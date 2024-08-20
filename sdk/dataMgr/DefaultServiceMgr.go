package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	circuit "github.com/rubyist/circuitbreaker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync"
	"time"
)

type ServiceInfoBuffer struct {
	Mutex       sync.Mutex
	InstanceDic map[int64]*InstanceInfo
	NameDic     map[string]*InstanceInfo
}

type DefaultServiceMgr struct {
	SkipSave bool

	config      *config.Config
	connManager connMgr.ConnManager
	mt          *metrics.MetricsManager

	serviceBuffer *util.SyncContainer[map[string]*ServiceInfoBuffer]
	routerBuffer  *util.SyncContainer[map[string]*ServiceRouterBuffer]

	watchChan *util.SyncContainer[map[string]*util.SyncContainer[[]chan<- *util.ServiceInfo]]
}

func (m *DefaultServiceMgr) RemoveInstance(serviceName string, instanceID int64) {
	service, ok := func() (service *ServiceInfoBuffer, ok bool) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok = m.serviceBuffer.Value[serviceName]
		if ok {
			service.Mutex.Lock()
		}
		return
	}()
	if !ok {
		return
	}
	defer service.Mutex.Unlock()
	info, ok := service.InstanceDic[instanceID]
	if !ok {
		return
	}
	delete(service.InstanceDic, instanceID)
	delete(service.NameDic, info.Name)
	util.Info("removeKVRouter %s Instance %v", serviceName, instanceID)
}

func (m *DefaultServiceMgr) AddInstance(serviceName string, info *pb.InstanceInfo) {
	if m.SkipSave {
		return
	}
	service := func() (service *ServiceInfoBuffer) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok := m.serviceBuffer.Value[serviceName]
		if !ok {
			service = &ServiceInfoBuffer{
				InstanceDic: make(map[int64]*InstanceInfo),
				NameDic:     make(map[string]*InstanceInfo),
			}
			m.serviceBuffer.Value[serviceName] = service
		}
		service.Mutex.Lock()
		return
	}()
	defer service.Mutex.Unlock()
	nowInfo, ok := service.NameDic[info.Name]
	if ok {
		delete(service.InstanceDic, nowInfo.InstanceID)
	}
	infoNow := &InstanceInfo{
		InstanceInfo: info,
		Breaker: circuit.NewBreakerWithOptions(&circuit.Options{
			ShouldTrip: circuit.ThresholdTripFunc(m.config.SDK.Breaker.ThresholdTrip),
			WindowTime: time.Duration(m.config.SDK.Breaker.WindowTime) * time.Second,
		}),
	}
	service.InstanceDic[info.InstanceID] = infoNow
	service.NameDic[info.Name] = infoNow
	util.Info("%s Add Instance %v", serviceName, info.InstanceID)
}

func (m *DefaultServiceMgr) GetServiceInfo(serviceName string) (*util.ServiceInfo, bool) {
	service, ok := func() (service *ServiceInfoBuffer, ok bool) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok = m.serviceBuffer.Value[serviceName]
		if ok {
			service.Mutex.Lock()
		}
		return
	}()
	if !ok {
		return nil, false
	}
	defer service.Mutex.Unlock()
	info := &util.ServiceInfo{
		Name:      serviceName,
		Instances: make([]util.DstInstanceInfo, 0, len(service.InstanceDic)),
	}
	for _, v := range service.InstanceDic {
		info.Instances = append(info.Instances, v)
	}
	return info, true
}

func (m *DefaultServiceMgr) GetInstanceInfo(serviceName string, instanceID int64) (*InstanceInfo, bool) {
	service, ok := func() (service *ServiceInfoBuffer, ok bool) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok = m.serviceBuffer.Value[serviceName]
		if ok {
			service.Mutex.Lock()
		}
		return
	}()
	if !ok {
		return nil, false
	}
	defer service.Mutex.Unlock()
	info, ok := service.InstanceDic[instanceID]
	return info, ok
}

func (m *DefaultServiceMgr) handleInstancePut(serviceName string, ev *clientv3.Event) (err error) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "InstanceKeyPut"}, err)
	}()
	info := &pb.InstanceInfo{}
	err = proto.Unmarshal(ev.Kv.Value, info)
	if err != nil {
		util.Error("err %v", err)
		return
	}
	m.AddInstance(serviceName, info)
	return
}

func (m *DefaultServiceMgr) handleInstanceDelete(serviceName string, ev *clientv3.Event) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "InstanceKeyDelete"}, nil)
	}()
	id := util.InstanceKey2InstanceID(string(ev.Kv.Key), serviceName)
	instanceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		util.Error("remove instance error: %v", err)
	} else {
		m.RemoveInstance(serviceName, instanceID)
	}
}

func (m *DefaultServiceMgr) handleInstanceService(cli *clientv3.Client, serviceName string) {
	m.tryAddServiceInfoChanList(serviceName)
	revision, err := m.initServiceInfo(cli, serviceName)
	if err != nil {
		util.Error("获取全量数据失败：%v", err)
	}
	go func() {
		rch := cli.Watch(context.Background(),
			util.ServiceAllInfoPrefix(serviceName),
			clientv3.WithPrefix(),
			clientv3.WithRev(revision),
		)
		for wresp := range rch {
			if m.SkipSave {
				continue
			}
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					err = m.handleInstancePut(serviceName, ev)
					if err != nil {
						util.Error("Put err: %v", err)
					}
				case clientv3.EventTypeDelete:
					m.handleInstanceDelete(serviceName, ev)
				}
			}
			m.sendServiceInfoAllChan(serviceName)
		}
	}()

}

func (m *DefaultServiceMgr) handleWatchService(serviceName string) {
	instanceCli, err := m.connManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		util.Error("%s 服务数据初始化失败", serviceName)
		return
	}
	m.handleInstanceService(instanceCli, serviceName)
	clis := m.connManager.GetAllServiceConn(connMgr.RoutersEtcd)
	for _, cli := range clis {
		m.handleWatchRouter(cli, serviceName)
	}
}

func (m *DefaultServiceMgr) tryAddServiceInfoChanList(serviceName string) *util.SyncContainer[[]chan<- *util.ServiceInfo] {
	var chans *util.SyncContainer[[]chan<- *util.ServiceInfo] = nil
	m.watchChan.Mutex.Lock()
	defer m.watchChan.Mutex.Unlock()
	if _, ok := m.watchChan.Value[serviceName]; ok {
		chans = m.watchChan.Value[serviceName]
	} else {
		chans = &util.SyncContainer[[]chan<- *util.ServiceInfo]{
			Value: make([]chan<- *util.ServiceInfo, 0),
		}
		m.watchChan.Value[serviceName] = chans
	}
	return chans
}

func (m *DefaultServiceMgr) sendServiceInfoAllChan(serviceName string) {
	var chanList *util.SyncContainer[[]chan<- *util.ServiceInfo]
	var ok bool
	func() {
		m.watchChan.Mutex.Lock()
		defer m.watchChan.Mutex.Unlock()
		chanList, ok = m.watchChan.Value[serviceName]
		if !ok {
			return
		}
		chanList.Mutex.Lock()
	}()
	if !ok {
		return
	}
	defer chanList.Mutex.Unlock()
	service, ok := m.GetServiceInfo(serviceName)
	if !ok {
		return
	}
	for _, ch := range chanList.Value {
		ch <- service
	}
}

func (m *DefaultServiceMgr) initServiceInfo(cli *clientv3.Client, serviceName string) (revision int64, err error) {
	resp, err := cli.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kv := range resp.Kvs {
		info := &pb.InstanceInfo{}
		err = proto.Unmarshal(kv.Value, info)
		if err != nil {
			continue
		}
		m.AddInstance(serviceName, info)
	}
	revision = resp.Header.Revision
	return
}

func (m *DefaultServiceMgr) startFlushInfo() {
	//如果处于离线模式
	//离线模式仅供性能测试使用
	if m.connManager == nil {
		return
	}
	for _, serviceName := range m.config.SDK.Discovery.DstService {
		m.handleWatchService(serviceName)
	}
}

func (m *DefaultServiceMgr) GetInstanceInfoByName(serviceName string, name string) (*InstanceInfo, bool) {
	service, ok := func() (service *ServiceInfoBuffer, ok bool) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok = m.serviceBuffer.Value[serviceName]
		if ok {
			service.Mutex.Lock()
		}
		return
	}()
	if !ok {
		return nil, false
	}
	defer service.Mutex.Unlock()
	info, ok := service.NameDic[name]
	return info, ok
}

func (m *DefaultServiceMgr) WatchServiceInfo(serviceName string) (<-chan *util.ServiceInfo, error) {
	serviceChan := m.tryAddServiceInfoChanList(serviceName)
	serviceChan.Mutex.Lock()
	defer serviceChan.Mutex.Unlock()
	ch := make(chan *util.ServiceInfo)
	serviceChan.Value = append(serviceChan.Value, ch)
	return ch, nil
}

func NewDefaultServiceMgr(
	config *config.Config,
	manager connMgr.ConnManager,
	mt *metrics.MetricsManager,
) *DefaultServiceMgr {
	mgr := &DefaultServiceMgr{
		SkipSave: false,

		config:      config,
		connManager: manager,
		mt:          mt,

		serviceBuffer: &util.SyncContainer[map[string]*ServiceInfoBuffer]{
			Value: make(map[string]*ServiceInfoBuffer),
		},
		routerBuffer: &util.SyncContainer[map[string]*ServiceRouterBuffer]{
			Value: make(map[string]*ServiceRouterBuffer),
		},
		watchChan: &util.SyncContainer[map[string]*util.SyncContainer[[]chan<- *util.ServiceInfo]]{
			Value: make(map[string]*util.SyncContainer[[]chan<- *util.ServiceInfo]),
		},
	}
	mgr.startFlushInfo()
	return mgr
}

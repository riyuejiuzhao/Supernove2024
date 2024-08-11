package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"log"
	"sync"
	"time"
)

type SyncContainer[T any] struct {
	Mutex sync.Mutex
	Value T
}

type ServiceInfoBuffer struct {
	Mutex       sync.Mutex
	InstanceDic map[int64]*pb.InstanceInfo
	AddressDic  map[string]*pb.InstanceInfo
}

type DefaultServiceMgr struct {
	config      *config.Config
	connManager connMgr.ConnManager
	mt          *metrics.MetricsManager

	serviceBuffer *SyncContainer[map[string]*ServiceInfoBuffer]
	routerBuffer  *SyncContainer[map[string]*ServiceRouterBuffer]
}

func (m *DefaultServiceMgr) RemoveInstanceByAddress(serviceName string, address string) {
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
	instanceInfo, ok := service.AddressDic[address]
	if !ok {
		return
	}
	delete(service.AddressDic, address)
	delete(service.InstanceDic, instanceInfo.InstanceID)
}

func (m *DefaultServiceMgr) AddInstance(serviceName string, info *pb.InstanceInfo) {
	service := func() (service *ServiceInfoBuffer) {
		m.serviceBuffer.Mutex.Lock()
		defer m.serviceBuffer.Mutex.Unlock()
		service, ok := m.serviceBuffer.Value[serviceName]
		if !ok {
			service = &ServiceInfoBuffer{
				AddressDic:  make(map[string]*pb.InstanceInfo),
				InstanceDic: make(map[int64]*pb.InstanceInfo),
			}
			m.serviceBuffer.Value[serviceName] = service
		}
		service.Mutex.Lock()
		return
	}()
	defer service.Mutex.Unlock()
	address := util.Address(info.Host, info.Port)
	nowInfo, ok := service.AddressDic[address]
	if ok && nowInfo.CreateTime > info.CreateTime {
		return
	}
	service.AddressDic[address] = info
	service.InstanceDic[info.InstanceID] = info
	util.Info("%s Add Instance %v %s", serviceName, info.InstanceID, address)
}

func (m *DefaultServiceMgr) GetServiceInfo(serviceName string) (*pb.ServiceInfo, bool) {
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
	info := &pb.ServiceInfo{
		ServiceName: serviceName,
		Instances:   make([]*pb.InstanceInfo, 0, len(service.InstanceDic)),
	}
	for _, v := range service.InstanceDic {
		info.Instances = append(info.Instances, v)
	}
	return info, true
}

func (m *DefaultServiceMgr) GetInstanceInfo(serviceName string, instanceID int64) (*pb.InstanceInfo, bool) {
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
	address := util.InstanceKey2Address(string(ev.Kv.Key), serviceName)
	m.RemoveInstanceByAddress(serviceName, address)
}

func (m *DefaultServiceMgr) handleWatchService(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	rch := cli.Watch(context.Background(), util.InstancePrefix(serviceName), clientv3.WithPrefix())
	go m.InitInstanceInfo(serviceName)
	for wresp := range rch {
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
	}
}

func (m *DefaultServiceMgr) InitInstanceInfo(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Get(context.Background(), util.InstancePrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
		util.Error("%v", err)
		return
	}
	for _, kv := range resp.Kvs {
		info := &pb.InstanceInfo{}
		err = proto.Unmarshal(kv.Value, info)
		if err != nil {
			util.Error("%v", err)
			continue
		}
		m.AddInstance(serviceName, info)
	}
}

func (m *DefaultServiceMgr) startFlushInfo() {
	for _, serviceName := range m.config.Global.Discovery.DstService {
		go m.handleWatchService(serviceName)
		go m.handleWatchKVRouter(serviceName)
		go m.handleWatchTargetRouter(serviceName)
	}
}

func NewDefaultServiceMgr(
	config *config.Config,
	manager connMgr.ConnManager,
	mt *metrics.MetricsManager,
) *DefaultServiceMgr {
	mgr := &DefaultServiceMgr{
		config:      config,
		connManager: manager,
		mt:          mt,

		serviceBuffer: &SyncContainer[map[string]*ServiceInfoBuffer]{
			Value: make(map[string]*ServiceInfoBuffer),
		},
		routerBuffer: &SyncContainer[map[string]*ServiceRouterBuffer]{
			Value: make(map[string]*ServiceRouterBuffer),
		},
	}
	mgr.startFlushInfo()
	return mgr
}

package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
	"sync"
	"time"
)

type KVRouterDictBuffer struct {
	dict map[string]*pb.KVRouterInfo
}

func (b *KVRouterDictBuffer) Add(info *pb.KVRouterInfo) {
	b.dict[info.Key] = info
}

func (b *KVRouterDictBuffer) Get(key string) (*pb.KVRouterInfo, bool) {
	info, ok := b.dict[key]
	return info, ok
}

func (b *KVRouterDictBuffer) Remove(key string) {
	delete(b.dict, key)
}

func (b *KVRouterDictBuffer) Clear() {
	clear(b.dict)
}

type KVRouterBuffer interface {
	Add(info *pb.KVRouterInfo)
	Get(key string) (*pb.KVRouterInfo, bool)
	Remove(key string)
	Clear()
}

type ServiceRouterBuffer struct {
	*pb.ServiceRouterInfo

	routerBufferLock sync.RWMutex
	kvRouterDic      KVRouterBuffer
	targetRouterDic  map[string]*pb.TargetRouterInfo
}

func (b *ServiceRouterBuffer) SetRouterInfo(info *pb.ServiceRouterInfo) {
	b.Revision = info.Revision
	b.ServiceRouterInfo = info

	b.kvRouterDic = &KVRouterDictBuffer{
		dict: make(map[string]*pb.KVRouterInfo),
	}
	b.targetRouterDic = make(map[string]*pb.TargetRouterInfo)
	for _, v := range info.KVRouters {
		b.kvRouterDic.Add(v)
	}
	for _, v := range info.TargetRouters {
		b.targetRouterDic[v.SrcInstanceID] = v
	}
}

func (b *ServiceRouterBuffer) TryGetKVRouter(key string, skipTimeCheck bool) (*pb.KVRouterInfo, bool) {
	b.routerBufferLock.RLock()
	defer b.routerBufferLock.RUnlock()
	r, ok := b.kvRouterDic.Get(key)
	if !ok || skipTimeCheck {
		return r, ok
	}
	if r.Timeout+r.CreateTime >= time.Now().Unix() {
		return r, ok
	}
	return nil, false
}

func (b *ServiceRouterBuffer) TryGetTargetRouter(srcID string, skipTimeCheck bool) (*pb.TargetRouterInfo, bool) {
	b.routerBufferLock.RLock()
	defer b.routerBufferLock.RUnlock()
	r, ok := b.targetRouterDic[srcID]
	if !ok || skipTimeCheck {
		return r, ok
	}
	if r.CreateTime+r.Timeout >= time.Now().Unix() {
		return r, ok
	}
	return nil, false
}

type SyncContainer[T any] struct {
	Mutex sync.RWMutex
	Value T
}

type ServiceInfoBuffer struct {
	*pb.ServiceInfo
	Mutex       sync.RWMutex
	InstanceDic map[string]*pb.InstanceInfo
}

func (b *ServiceInfoBuffer) SetServiceInfo(info *pb.ServiceInfo) {
	b.ServiceInfo = info
	b.InstanceDic = make(map[string]*pb.InstanceInfo)
	if info.Instances == nil {
		return
	}
	for _, i := range info.Instances {
		b.InstanceDic[i.InstanceID] = i
	}
}

type DefaultServiceMgr struct {
	config      *config.Config
	connManager connMgr.ConnManager

	serviceBuffer *SyncContainer[map[string]*ServiceInfoBuffer]

	healthBuffer *SyncContainer[map[string]*SyncContainer[map[string]*pb.InstanceHealthInfo]]

	routerBuffer *SyncContainer[map[string]*ServiceRouterBuffer]
}

func (m *DefaultServiceMgr) flushAllHealthInfoLocked() {
	for _, serviceName := range m.config.Global.Discovery.DstService {
		m.healthBuffer.Mutex.Lock()
		service, ok := m.healthBuffer.Value[serviceName]
		if !ok {
			service = &SyncContainer[map[string]*pb.InstanceHealthInfo]{}
			m.healthBuffer.Value[serviceName] = service
		}
		service.Mutex.Lock()
		m.healthBuffer.Mutex.Unlock()
		m.flushHealthInfo(serviceName, service)
		service.Mutex.Unlock()
	}
}

func (m *DefaultServiceMgr) flushAllRouterLocked() {
	for _, serviceName := range m.config.Global.Discovery.DstService {
		m.routerBuffer.Mutex.Lock()
		router, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			router = &ServiceRouterBuffer{ServiceRouterInfo: &pb.ServiceRouterInfo{
				ServiceName: serviceName,
				Revision:    0,
			}}
			m.routerBuffer.Value[serviceName] = router
		}
		router.routerBufferLock.Lock()
		m.routerBuffer.Mutex.Unlock()
		m.flushRouter(router)
		router.routerBufferLock.Unlock()
	}
}

func (m *DefaultServiceMgr) flushRouter(buffer *ServiceRouterBuffer) {
	disConn, err := m.connManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", buffer.ServiceName, err)
		return
	}
	cli := pb.NewDiscoveryServiceClient(disConn)

	request := &pb.GetRoutersRequest{
		ServiceName: buffer.ServiceName,
		Revision:    buffer.Revision,
	}

	reply, err := cli.GetRouters(context.Background(), request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", buffer.ServiceName, err)
		return
	}
	if reply.Router.Revision == request.Revision {
		//util.Info("无需更新本地路由缓存 %v", request.Revision)
		return
	}

	buffer.SetRouterInfo(reply.Router)
	util.Info("更新本地路由缓存 %v->%v", request.Revision, reply.Router.Revision)
}

func (m *DefaultServiceMgr) flushAllServiceLocked() {
	for _, serviceName := range m.config.Global.Discovery.DstService {
		m.serviceBuffer.Mutex.Lock()
		service, ok := m.serviceBuffer.Value[serviceName]
		if !ok {
			service = &ServiceInfoBuffer{}
			service.SetServiceInfo(&pb.ServiceInfo{ServiceName: serviceName, Revision: 0})
			m.serviceBuffer.Value[serviceName] = service
		}
		service.Mutex.Lock()
		m.serviceBuffer.Mutex.Unlock()
		m.flushService(service)
		service.Mutex.Unlock()
	}
}

func (m *DefaultServiceMgr) flushService(service *ServiceInfoBuffer) {
	disConn, err := m.connManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", service.ServiceName, err)
		return
	}
	cli := pb.NewDiscoveryServiceClient(disConn)

	request := &pb.GetInstancesRequest{
		ServiceName: service.ServiceName,
		Revision:    service.Revision,
	}

	reply, err := cli.GetInstances(context.Background(), request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", service.ServiceName, err)
		return
	}
	if reply.Service.Revision == request.Revision {
		util.Info("无需更新本地缓存 %v", request.Revision)
		return
	}

	service.SetServiceInfo(reply.Service)
	util.Info("更新本地缓存 %v->%v", request.Revision, reply.Service.Revision)
}

func (m *DefaultServiceMgr) GetServiceInfo(serviceName string) (*pb.ServiceInfo, bool) {
	m.serviceBuffer.Mutex.RLock()
	defer m.serviceBuffer.Mutex.RUnlock()
	service, ok := m.serviceBuffer.Value[serviceName]
	if !ok {
		return nil, false
	}
	service.Mutex.RLock()
	defer service.Mutex.RUnlock()
	return service.ServiceInfo, true
}

func (m *DefaultServiceMgr) GetInstanceInfo(serviceName string, instanceID string) (*pb.InstanceInfo, bool) {
	m.serviceBuffer.Mutex.RLock()
	defer m.serviceBuffer.Mutex.RUnlock()
	service, ok := m.serviceBuffer.Value[serviceName]
	if !ok {
		return nil, false
	}
	service.Mutex.RLock()
	defer service.Mutex.RUnlock()
	instance, ok := service.InstanceDic[instanceID]
	if !ok {
		return nil, false
	}
	return instance, true
}

func (m *DefaultServiceMgr) GetHealthInfo(serviceName string, instanceID string) (*pb.InstanceHealthInfo, bool) {
	m.healthBuffer.Mutex.RLock()
	defer m.healthBuffer.Mutex.RUnlock()
	serviceDic, ok := m.healthBuffer.Value[serviceName]
	if !ok {
		return nil, false
	}
	serviceDic.Mutex.RLock()
	defer serviceDic.Mutex.RUnlock()
	ins, ok := serviceDic.Value[instanceID]
	if !ok {
		return nil, false
	}
	return ins, true
}

func (m *DefaultServiceMgr) flushHealthInfo(serviceName string, serviceHealth *SyncContainer[map[string]*pb.InstanceHealthInfo]) {
	conn, err := m.connManager.GetServiceConn(connMgr.HealthCheck)
	if err != nil {
		util.Error("更新健康信息连接获取失败 err:%v", err)
		return
	}

	cli := pb.NewHealthServiceClient(conn)
	reply, err := cli.GetHealthInfo(context.Background(), &pb.GetHealthInfoRequest{ServiceName: serviceName})
	if err != nil {
		util.Error("更新健康信息rpc失败 err:%v", err)
		return
	}
	serviceHealthInfo := make(map[string]*pb.InstanceHealthInfo)
	for _, ins := range reply.HealthInfo.InstanceHealthInfo {
		serviceHealthInfo[ins.InstanceID] = ins
	}
	serviceHealth.Value = serviceHealthInfo
}

func (m *DefaultServiceMgr) GetTargetRouter(ServiceName string, SrcInstanceID string, skipTimeCheck bool) (*pb.TargetRouterInfo, bool) {
	m.routerBuffer.Mutex.RLock()
	defer m.routerBuffer.Mutex.RUnlock()

	b, ok := m.routerBuffer.Value[ServiceName]
	if !ok {
		return nil, false
	}
	return b.TryGetTargetRouter(SrcInstanceID, skipTimeCheck)
}

func (m *DefaultServiceMgr) GetKVRouter(ServiceName string, Key string, skipTimeCheck bool) (*pb.KVRouterInfo, bool) {
	m.routerBuffer.Mutex.RLock()
	defer m.routerBuffer.Mutex.RUnlock()

	b, ok := m.routerBuffer.Value[ServiceName]
	if !ok {
		return nil, false
	}
	return b.TryGetKVRouter(Key, skipTimeCheck)
}

func (m *DefaultServiceMgr) startFlushInfo() {
	m.flushAllServiceLocked()
	m.flushAllHealthInfoLocked()
	m.flushAllRouterLocked()
	go func() {
		for {
			time.Sleep(time.Duration(m.config.Global.Discovery.DefaultTimeout) * time.Second)
			m.flushAllHealthInfoLocked()
			m.flushAllRouterLocked()
			m.flushAllServiceLocked()
		}
	}()
}

func NewDefaultServiceMgr(config *config.Config, manager connMgr.ConnManager) *DefaultServiceMgr {
	mgr := &DefaultServiceMgr{
		config:      config,
		connManager: manager,
		serviceBuffer: &SyncContainer[map[string]*ServiceInfoBuffer]{
			Value: make(map[string]*ServiceInfoBuffer),
		},
		healthBuffer: &SyncContainer[map[string]*SyncContainer[map[string]*pb.InstanceHealthInfo]]{
			Value: make(map[string]*SyncContainer[map[string]*pb.InstanceHealthInfo]),
		},
		routerBuffer: &SyncContainer[map[string]*ServiceRouterBuffer]{
			Value: make(map[string]*ServiceRouterBuffer),
		},
	}
	mgr.startFlushInfo()
	return mgr
}

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

type ServiceRouterBuffer struct {
	routerBufferLock sync.RWMutex
	kvRouter         map[string]*pb.KVRouterInfo
	targetRouter     map[string]*pb.TargetRouterInfo
}

func (b *ServiceRouterBuffer) TryGetKVRouter(key string) (*pb.KVRouterInfo, bool) {
	b.routerBufferLock.RLock()
	defer b.routerBufferLock.RUnlock()
	r, ok := b.kvRouter[key]
	return r, ok
}

func (b *ServiceRouterBuffer) TryGetTargetRouter(srcID string) (*pb.TargetRouterInfo, bool) {
	b.routerBufferLock.RLock()
	defer b.routerBufferLock.RUnlock()
	r, ok := b.targetRouter[srcID]
	return r, ok
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

	routerBufferLock sync.RWMutex
	routerBuffer     map[string]*ServiceRouterBuffer
}

func (m *DefaultServiceMgr) flushAllHealthInfoLocked() {
	for _, serviceName := range m.config.Global.Discovery.DstService {
		m.healthBuffer.Mutex.Lock()
		service, ok := m.healthBuffer.Value[serviceName]
		if !ok {
			service = &SyncContainer[map[string]*pb.InstanceHealthInfo]{}
			m.healthBuffer.Value[serviceName] = service
		}
		m.healthBuffer.Mutex.Unlock()
		m.flushHealthInfoLocked(serviceName, service)
	}
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
		m.serviceBuffer.Mutex.Unlock()
		m.flushServiceLocked(service)
	}
}

func (m *DefaultServiceMgr) flushServiceLocked(service *ServiceInfoBuffer) {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	disConn, err := m.connManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", service.ServiceName, err)
		return
	}
	defer disConn.Close()
	cli := pb.NewDiscoveryServiceClient(disConn.Value())

	request := &pb.GetInstancesRequest{
		ServiceName: service.ServiceName,
		Revision:    service.Revision,
	}

	reply, err := cli.GetInstances(context.Background(), request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", service.ServiceName, err)
	}
	if reply.Revision == request.Revision {
		util.Info("无需更新本地缓存 %v", request.Revision)
		return
	}

	service.SetServiceInfo(&pb.ServiceInfo{
		Instances:   reply.Instances,
		Revision:    reply.Revision,
		ServiceName: request.ServiceName,
	})
	util.Info("更新本地缓存 %v->%v", request.Revision, reply.Revision)
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
	m.serviceBuffer.Mutex.RUnlock()
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

func (m *DefaultServiceMgr) flushHealthInfoLocked(serviceName string, serviceHealth *SyncContainer[map[string]*pb.InstanceHealthInfo]) {
	serviceHealth.Mutex.Lock()
	defer serviceHealth.Mutex.Unlock()

	conn, err := m.connManager.GetServiceConn(connMgr.HealthCheck)
	if err != nil {
		util.Error("更新健康信息连接获取失败 err:%v", err)
		return
	}
	defer conn.Close()

	cli := pb.NewHealthServiceClient(conn.Value())
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

func (m *DefaultServiceMgr) GetTargetRouter(ServiceName string, SrcInstanceID string) (*pb.TargetRouterInfo, bool) {
	m.routerBufferLock.RLock()
	defer m.routerBufferLock.RUnlock()

	b, ok := m.routerBuffer[ServiceName]
	if !ok {
		return nil, false
	}
	return b.TryGetTargetRouter(SrcInstanceID)
}

func (m *DefaultServiceMgr) GetKVRouter(ServiceName string, Key string) (*pb.KVRouterInfo, bool) {
	m.routerBufferLock.RLock()
	defer m.routerBufferLock.RUnlock()

	b, ok := m.routerBuffer[ServiceName]
	if !ok {
		return nil, false
	}
	return b.TryGetKVRouter(Key)
}

func (m *DefaultServiceMgr) startFlushInfo() {
	m.flushAllServiceLocked()
	m.flushAllHealthInfoLocked()
	go func() {
		for {
			time.Sleep(time.Duration(m.config.Global.Discovery.DefaultTimeout) * time.Second)
			m.flushAllHealthInfoLocked()
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
	}
	mgr.startFlushInfo()
	return mgr
}

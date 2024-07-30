package dataMgr

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
	"sync"
	"time"
)

type DefaultServiceMgr struct {
	config      *config.Config
	connManager connMgr.ConnManager

	serviceBufferLock sync.RWMutex
	serviceBuffer     map[string]*miniRouterProto.ServiceInfo

	healthBufferLock sync.RWMutex
	healthBuffer     map[string]map[string]*miniRouterProto.InstanceHealthInfo
}

func (m *DefaultServiceMgr) newGetInstancesRequest(serviceName string) *miniRouterProto.GetInstancesRequest {
	m.serviceBufferLock.RLock()
	defer m.serviceBufferLock.RUnlock()
	nowService, ok := m.serviceBuffer[serviceName]
	if !ok {
		nowService = &miniRouterProto.ServiceInfo{
			ServiceName: serviceName,
			Instances:   make([]*miniRouterProto.InstanceInfo, 0),
			Revision:    0}
		m.serviceBuffer[serviceName] = nowService
		m.healthBuffer[serviceName] = make(map[string]*miniRouterProto.InstanceHealthInfo)
	}
	request := miniRouterProto.GetInstancesRequest{
		ServiceName: serviceName,
		Revision:    nowService.Revision,
	}
	return &request
}

func (m *DefaultServiceMgr) flushService(serviceName string, reply *miniRouterProto.GetInstancesReply) {
	m.serviceBufferLock.Lock()
	defer m.serviceBufferLock.Unlock()
	instances := make([]*miniRouterProto.InstanceInfo, len(reply.Instances))
	for i, v := range reply.Instances {
		instances[i] = v
	}
	m.serviceBuffer[serviceName] = &miniRouterProto.ServiceInfo{
		Instances:   instances,
		Revision:    reply.Revision,
		ServiceName: serviceName,
	}
}

func (m *DefaultServiceMgr) tryFlushService(serviceName string) {
	disConn, err := m.connManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", serviceName, err)
		return
	}
	defer disConn.Close()
	cli := miniRouterProto.NewDiscoveryServiceClient(disConn.Value())

	request := m.newGetInstancesRequest(serviceName)

	reply, err := cli.GetInstances(context.Background(), request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", serviceName, err)
	}
	if reply.Revision == request.Revision {
		util.Info("无需更新本地缓存 %v", request.Revision)
		return
	}
	m.flushService(serviceName, reply)
	util.Info("更新本地缓存 %v->%v", request.Revision, reply.Revision)
}

func (m *DefaultServiceMgr) getServiceInstanceInBuffer(serviceName string) (*miniRouterProto.ServiceInfo, bool) {
	m.serviceBufferLock.RLock()
	defer m.serviceBufferLock.RUnlock()
	service, ok := m.serviceBuffer[serviceName]
	return service, ok
}

// GetServiceInstance 第一次取用的时候强制刷新，之后按时刷新
func (m *DefaultServiceMgr) GetServiceInstance(serviceName string) *miniRouterProto.ServiceInfo {
	service, ok := m.getServiceInstanceInBuffer(serviceName)
	if !ok {
		m.tryFlushService(serviceName)
		service, ok = m.getServiceInstanceInBuffer(serviceName)
	}
	return service
}

func (m *DefaultServiceMgr) GetHealthInfo(serviceName string, instanceID string) (*miniRouterProto.InstanceHealthInfo, bool) {
	m.healthBufferLock.RLock()
	defer m.healthBufferLock.RUnlock()
	serviceDic, ok := m.healthBuffer[serviceName]
	if !ok {
		return nil, false
	}
	instance, ok := serviceDic[instanceID]
	if !ok {
		return nil, false
	}

	return instance, true
}

func (m *DefaultServiceMgr) flushHealthInfo() {
	conn, err := m.connManager.GetServiceConn(connMgr.HealthCheck)
	if err != nil {
		util.Error("更新健康信息连接获取失败 err:%v", err)
		return
	}
	defer conn.Close()

	m.healthBufferLock.Lock()
	defer m.healthBufferLock.Unlock()

	cli := miniRouterProto.NewHealthServiceClient(conn.Value())
	serviceList := make([]string, 0, len(m.healthBuffer))
	for k, _ := range m.healthBuffer {
		serviceList = append(serviceList, k)
	}
	reply, err := cli.GetHealthInfo(context.Background(), &miniRouterProto.GetHealthInfoRequest{
		ServiceNames: serviceList,
	})
	if err != nil {
		util.Error("更新健康信息rpc失败 err:%v", err)
		return
	}
	newDic := make(map[string]map[string]*miniRouterProto.InstanceHealthInfo)
	for _, v := range reply.HealthInfos {
		serviceHealthInfo := make(map[string]*miniRouterProto.InstanceHealthInfo)
		for _, ins := range v.InstanceHealthInfo {
			serviceHealthInfo[ins.InstanceID] = ins
		}
		newDic[v.ServiceName] = serviceHealthInfo
	}
	m.healthBuffer = newDic
}

func (m *DefaultServiceMgr) getAllServiceName() []string {
	m.serviceBufferLock.RLock()
	defer m.serviceBufferLock.RUnlock()
	allName := make([]string, 0, len(m.serviceBuffer))
	for k, _ := range m.serviceBuffer {
		allName = append(allName, k)
	}
	return allName
}

func (m *DefaultServiceMgr) startFlushInfo() {
	go func() {
		for {
			m.flushHealthInfo()
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for _, name := range m.getAllServiceName() {
			m.tryFlushService(name)
		}
		time.Sleep(5 * time.Second)
	}()
}

func NewDefaultServiceMgr(config *config.Config, manager connMgr.ConnManager) *DefaultServiceMgr {
	mgr := &DefaultServiceMgr{
		config:        config,
		connManager:   manager,
		serviceBuffer: make(map[string]*miniRouterProto.ServiceInfo),
	}
	mgr.startFlushInfo()
	return mgr
}

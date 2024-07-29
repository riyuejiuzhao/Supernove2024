package dataMgr

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
)

type DefaultServiceMgr struct {
	config        *config.Config
	connManager   connMgr.ConnManager
	serviceBuffer map[string]*miniRouterProto.ServiceInfo
	healthBuffer  map[string]map[string]*miniRouterProto.InstanceHealthInfo
}

func (m *DefaultServiceMgr) FlushService(serviceName string) {
	nowService, ok := m.serviceBuffer[serviceName]
	if !ok {
		nowService = &miniRouterProto.ServiceInfo{
			ServiceName: serviceName,
			Instances:   make([]*miniRouterProto.InstanceInfo, 0),
			Revision:    0}
		m.serviceBuffer[serviceName] = nowService
	}
	disConn, err := m.connManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", serviceName, err)
		return
	}
	defer disConn.Close()
	cli := miniRouterProto.NewDiscoveryServiceClient(disConn.Value())
	request := miniRouterProto.GetInstancesRequest{
		ServiceName: serviceName,
		Revision:    nowService.Revision,
	}
	reply, err := cli.GetInstances(context.Background(), &request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", serviceName, err)
	}
	if reply.Revision == nowService.Revision {
		util.Info("无需更新本地缓存 %v", nowService.Revision)
		return
	}
	instances := make([]*miniRouterProto.InstanceInfo, len(reply.Instances))
	for i, v := range reply.Instances {
		instances[i] = v
	}
	m.serviceBuffer[serviceName] = &miniRouterProto.ServiceInfo{
		Instances:   instances,
		Revision:    reply.Revision,
		ServiceName: serviceName,
	}
	util.Info("更新本地缓存 %v->%v", nowService.Revision, reply.Revision)
}

func (m *DefaultServiceMgr) GetServiceInstance(serviceName string) *miniRouterProto.ServiceInfo {
	m.FlushService(serviceName)
	service := m.serviceBuffer[serviceName]
	return service
}

func (m *DefaultServiceMgr) GetHealthInfo(serviceName string, instanceID string) (*miniRouterProto.InstanceHealthInfo, bool) {
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

func NewDefaultServiceMgr(config *config.Config, manager connMgr.ConnManager) *DefaultServiceMgr {
	return &DefaultServiceMgr{
		config:        config,
		connManager:   manager,
		serviceBuffer: make(map[string]*miniRouterProto.ServiceInfo),
	}
}

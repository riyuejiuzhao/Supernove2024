package dataMgr

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
)

type DefaultServiceMgr struct {
	sdk.APIContext
	buffer   map[string]*miniRouterProto.ServiceInfo
	revision int64
}

func (m *DefaultServiceMgr) FlushService(serviceName string) {
	nowService, ok := m.buffer[serviceName]
	if !ok {
		nowService = &miniRouterProto.ServiceInfo{
			ServiceName: serviceName,
			Instances:   make([]*miniRouterProto.InstanceInfo, 0),
			Revision:    0}
	}
	m.buffer[serviceName] = nowService
	disConn, err := m.ConnManager.GetServiceConn(connMgr.Discovery)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, 无法获取链接, err: %v", serviceName, err)
		return
	}
	defer disConn.Close()
	cli := miniRouterProto.NewDiscoveryServiceClient(disConn.Value())
	request := miniRouterProto.GetInstancesRequest{
		ServiceName: serviceName,
		Revision:    m.revision,
	}
	reply, err := cli.GetInstances(context.Background(), &request)
	if err != nil {
		util.Error("更新服务 %v 缓冲数据失败, grpc错误, err: %v", serviceName, err)
	}
	instances := make([]*miniRouterProto.InstanceInfo, len(reply.Instances))
	for i, v := range reply.Instances {
		instances[i].Host = v.Host
		instances[i].Port = v.Port
		instances[i].InstanceID = v.InstanceID
	}
	m.buffer[serviceName] = &miniRouterProto.ServiceInfo{
		Instances:   instances,
		Revision:    reply.Revision,
		ServiceName: serviceName,
	}
}

func (m *DefaultServiceMgr) GetServiceInstance(serviceName string) *miniRouterProto.ServiceInfo {
	m.FlushService(serviceName)
	service := m.buffer[serviceName]
	return service
}

func (m *DefaultServiceMgr) GetRevision() int64 {
	return m.revision
}

func NewDefaultServiceMgr(config config.Config, manager connMgr.ConnManager) {

}

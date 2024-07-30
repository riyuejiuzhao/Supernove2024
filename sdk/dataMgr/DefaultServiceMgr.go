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
	config        *config.Config
	connManager   connMgr.ConnManager
	serviceBuffer map[string]*miniRouterProto.ServiceInfo

	healthBufferLock sync.RWMutex
	healthBuffer     map[string]map[string]*miniRouterProto.InstanceHealthInfo
}

func (m *DefaultServiceMgr) FlushService(serviceName string) {
	nowService, ok := m.serviceBuffer[serviceName]
	if !ok {
		nowService = &miniRouterProto.ServiceInfo{
			ServiceName: serviceName,
			Instances:   make([]*miniRouterProto.InstanceInfo, 0),
			Revision:    0}
		m.serviceBuffer[serviceName] = nowService
		m.healthBuffer[serviceName] = make(map[string]*miniRouterProto.InstanceHealthInfo)
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

func (m *DefaultServiceMgr) startFlushHealthInfo() {
	go func() {
		for {
			m.flushHealthInfo()
			time.Sleep(5 * time.Second)
		}
	}()
}

func NewDefaultServiceMgr(config *config.Config, manager connMgr.ConnManager) *DefaultServiceMgr {
	mgr := &DefaultServiceMgr{
		config:        config,
		connManager:   manager,
		serviceBuffer: make(map[string]*miniRouterProto.ServiceInfo),
	}
	mgr.startFlushHealthInfo()
	return mgr
}

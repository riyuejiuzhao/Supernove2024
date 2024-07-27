package main

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"fmt"
	"strconv"
	"sync"
)

type InstanceMgr struct {
	ServiceInfo        *miniRouterProto.ServiceInfo
	InstanceAddressDic map[string]*miniRouterProto.InstanceInfo
	InstanceIdDic      map[string]*miniRouterProto.InstanceInfo
}

type DefaultRegisterDataManager struct {
	rwMutex sync.RWMutex
	dict    map[string]*InstanceMgr
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func (m *DefaultRegisterDataManager) FlushService(info *miniRouterProto.ServiceInfo) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	// 直接把过去的删掉重新生成，因为service可能注销
	instanceMgr := &InstanceMgr{
		InstanceAddressDic: make(map[string]*miniRouterProto.InstanceInfo),
		InstanceIdDic:      make(map[string]*miniRouterProto.InstanceInfo),
		ServiceInfo:        info,
	}
	for _, v := range info.Instances {
		instanceMgr.InstanceAddressDic[InstanceAddress(v.Host, v.Port)] = v
		instanceMgr.InstanceIdDic[v.InstanceID] = v
	}
	m.dict[info.ServiceName] = instanceMgr
}

func (m *DefaultRegisterDataManager) TryGetInstanceByAddress(
	service string, address string,
) (*miniRouterProto.InstanceInfo, bool) {
	m.rwMutex.RLocker()
	defer m.rwMutex.RUnlock()
	instanceMgr, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	v, ok := instanceMgr.InstanceAddressDic[address]
	return v, ok
}

func (m *DefaultRegisterDataManager) TryGetInstanceByID(
	service string, instanceID string,
) (*miniRouterProto.InstanceInfo, bool) {
	m.rwMutex.RLocker()
	defer m.rwMutex.RUnlock()
	instanceMgr, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	v, ok := instanceMgr.InstanceIdDic[instanceID]
	return v, ok
}

func (m *DefaultRegisterDataManager) TryGetServiceInfo(service string) (*miniRouterProto.ServiceInfo, bool) {
	m.rwMutex.RLocker()
	defer m.rwMutex.RUnlock()

	ins, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	return ins.ServiceInfo, true
}

func (m *DefaultRegisterDataManager) AddInstances(
	serviceName string,
	host string,
	port int32,
	weight int32,
) (info *miniRouterProto.InstanceInfo) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	mgr, ok := m.dict[serviceName]
	if !ok {
		mgr = &InstanceMgr{
			InstanceAddressDic: make(map[string]*miniRouterProto.InstanceInfo),
			InstanceIdDic:      make(map[string]*miniRouterProto.InstanceInfo),
			ServiceInfo:        util.NewServiceInfo(serviceName),
		}
	}
	mgr.ServiceInfo.Revision += 1
	info = &miniRouterProto.InstanceInfo{
		InstanceID: strconv.FormatInt(mgr.ServiceInfo.Revision, 10),
		Host:       host,
		Port:       port,
		Weight:     weight,
	}
	mgr.InstanceIdDic[info.InstanceID] = info
	mgr.InstanceAddressDic[InstanceAddress(info.Host, info.Port)] = info
	mgr.ServiceInfo.Instances = append(mgr.ServiceInfo.Instances, info)
	return
}

// RegisterDataManager 用来快速查找当前是否存在某个Instance
// 实现中应当考虑并发
type RegisterDataManager interface {
	// TryGetInstanceByAddress 获取对应服务实例
	TryGetInstanceByAddress(service string, address string) (*miniRouterProto.InstanceInfo, bool)
	// TryGetInstanceByID 获取对应服务实例
	TryGetInstanceByID(service string, instanceID string) (*miniRouterProto.InstanceInfo, bool)
	// TryGetServiceInfo 获取服务信息
	TryGetServiceInfo(service string) (*miniRouterProto.ServiceInfo, bool)

	// AddInstances 增加一个Instances
	AddInstances(serviceName string, host string, port int32, weight int32) *miniRouterProto.InstanceInfo
	// FlushService 更新或者创建
	FlushService(info *miniRouterProto.ServiceInfo)
}

package register

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
}

type DefaultServiceBuffer struct {
	rwMutex sync.RWMutex
	dict    map[string]*InstanceMgr
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func (m *DefaultServiceBuffer) FlushService(info *miniRouterProto.ServiceInfo) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	nowInfo, ok := m.dict[info.ServiceName]
	if ok && nowInfo.ServiceInfo.Revision == info.Revision {
		return
	}
	// 直接把过去的删掉重新生成，因为service可能注销
	instanceMgr := &InstanceMgr{
		InstanceAddressDic: make(map[string]*miniRouterProto.InstanceInfo),
		//InstanceIdDic:      make(map[string]*miniRouterProto.InstanceInfo),
		ServiceInfo: info,
	}
	for _, v := range info.Instances {
		instanceMgr.InstanceAddressDic[InstanceAddress(v.Host, v.Port)] = v
		//instanceMgr.InstanceIdDic[v.InstanceID] = v
	}
	m.dict[info.ServiceName] = instanceMgr
}

func (m *DefaultServiceBuffer) TryGetInstanceByAddress(
	service string, address string,
) (*miniRouterProto.InstanceInfo, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	instanceMgr, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	v, ok := instanceMgr.InstanceAddressDic[address]
	return v, ok
}

func (m *DefaultServiceBuffer) TryGetServiceInfo(service string) (*miniRouterProto.ServiceInfo, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	ins, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	return ins.ServiceInfo, true
}

func (m *DefaultServiceBuffer) AddInstance(
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
			ServiceInfo:        util.NewServiceInfo(serviceName),
		}
		m.dict[serviceName] = mgr
	}
	mgr.ServiceInfo.Revision += 1
	info = &miniRouterProto.InstanceInfo{
		InstanceID: strconv.FormatInt(mgr.ServiceInfo.Revision, 10),
		Host:       host,
		Port:       port,
		Weight:     weight,
	}
	mgr.InstanceAddressDic[InstanceAddress(info.Host, info.Port)] = info
	mgr.ServiceInfo.Instances = append(mgr.ServiceInfo.Instances, info)
	return
}

func (m *DefaultServiceBuffer) RemoveInstance(
	serviceName string, host string, port int32,
) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	mgr, ok := m.dict[serviceName]
	if !ok {
		return
	}

	instance, ok := mgr.InstanceAddressDic[InstanceAddress(host, port)]
	if !ok {
		return
	}

	mgr.ServiceInfo.Revision += 1
	removeIndex := 0
	for i, v := range mgr.ServiceInfo.Instances {
		if v.InstanceID == instance.InstanceID {
			removeIndex = i
			break
		}
	}
	copy(mgr.ServiceInfo.Instances[removeIndex:],
		mgr.ServiceInfo.Instances[removeIndex+1:])
	mgr.ServiceInfo.Instances = mgr.ServiceInfo.Instances[:len(mgr.ServiceInfo.Instances)-1]
}

func NewDefaultServiceBuffer() ServiceBuffer {
	return &DefaultServiceBuffer{
		dict: make(map[string]*InstanceMgr),
	}
}

var (
	NewServiceBuffer = NewDefaultServiceBuffer()
)

// ServiceBuffer 用来快速查找当前是否存在某个Instance
// 实现中应当考虑并发
type ServiceBuffer interface {
	// TryGetInstanceByAddress 获取对应服务实例
	TryGetInstanceByAddress(service string, address string) (*miniRouterProto.InstanceInfo, bool)
	// TryGetServiceInfo 获取服务信息
	TryGetServiceInfo(service string) (*miniRouterProto.ServiceInfo, bool)

	// AddInstance 增加一个Instance
	AddInstance(serviceName string, host string, port int32, weight int32) *miniRouterProto.InstanceInfo
	// FlushService 更新或者创建
	FlushService(info *miniRouterProto.ServiceInfo)
	// RemoveInstance 删除一个Instance
	RemoveInstance(serviceName string, host string, port int32)
}

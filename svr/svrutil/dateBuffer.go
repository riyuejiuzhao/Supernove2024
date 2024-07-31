package svrutil

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"errors"
	"fmt"
	"strconv"
	"sync"
)

type InstanceMgr struct {
	ServiceInfo        *pb.ServiceInfo
	InstanceAddressDic map[string]*pb.InstanceInfo
}

type DefaultServiceBuffer struct {
	rwMutex sync.RWMutex
	dict    map[string]*InstanceMgr
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func (m *DefaultServiceBuffer) FlushService(info *pb.ServiceInfo) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	nowInfo, ok := m.dict[info.ServiceName]
	if ok && nowInfo.ServiceInfo.Revision == info.Revision {
		return
	}
	// 直接把过去的删掉重新生成，因为service可能注销
	instanceMgr := &InstanceMgr{
		InstanceAddressDic: make(map[string]*pb.InstanceInfo),
		ServiceInfo:        info,
	}
	for _, v := range info.Instances {
		instanceMgr.InstanceAddressDic[InstanceAddress(v.Host, v.Port)] = v
	}
	m.dict[info.ServiceName] = instanceMgr
}

func (m *DefaultServiceBuffer) TryGetInstanceByAddress(
	service string, address string,
) (*pb.InstanceInfo, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	instanceMgr, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	v, ok := instanceMgr.InstanceAddressDic[address]
	return v, ok
}

func (m *DefaultServiceBuffer) TryGetServiceInfo(service string) (*pb.ServiceInfo, bool) {
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
) (info *pb.InstanceInfo) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	mgr, ok := m.dict[serviceName]
	if !ok {
		mgr = &InstanceMgr{
			InstanceAddressDic: make(map[string]*pb.InstanceInfo),
			ServiceInfo:        util.NewServiceInfo(serviceName),
		}
		m.dict[serviceName] = mgr
	}
	mgr.ServiceInfo.Revision += 1
	info = &pb.InstanceInfo{
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
	serviceName string, host string, port int32, instanceID string,
) error {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	mgr, ok := m.dict[serviceName]
	if !ok {
		return nil
	}

	instance, ok := mgr.InstanceAddressDic[InstanceAddress(host, port)]
	if !ok {
		return nil
	}

	if instance.InstanceID != instanceID {
		return errors.New("删除的InstanceID和实际InstanceID不同，终止删除")
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
	return nil
}

func newDefaultServiceBuffer() ServiceBuffer {
	return &DefaultServiceBuffer{
		dict: make(map[string]*InstanceMgr),
	}
}

var (
	NewServiceBuffer = newDefaultServiceBuffer
)

// ServiceBuffer 用来快速查找当前是否存在某个Instance
// 实现中应当考虑并发
type ServiceBuffer interface {
	// TryGetInstanceByAddress 获取对应服务实例
	TryGetInstanceByAddress(service string, address string) (*pb.InstanceInfo, bool)
	// TryGetServiceInfo 获取服务信息
	TryGetServiceInfo(service string) (*pb.ServiceInfo, bool)

	// AddInstance 增加一个Instance
	AddInstance(serviceName string, host string, port int32, weight int32) *pb.InstanceInfo
	// FlushService 更新或者创建
	FlushService(info *pb.ServiceInfo)
	// RemoveInstance 删除一个Instance
	RemoveInstance(serviceName string, host string, port int32, instanceID string) error
}

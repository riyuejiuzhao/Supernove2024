package svrutil

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"errors"
	"fmt"
	"sync"
)

type InstanceMgr struct {
	*pb.ServiceInfo

	rwMutex            sync.RWMutex
	InstanceAddressDic map[string]*pb.InstanceInfo
	InstanceIDDic      map[string]*pb.InstanceInfo
}

func (m *InstanceMgr) Reset(info *pb.ServiceInfo) {
	m.ServiceInfo = info
	m.InstanceAddressDic = make(map[string]*pb.InstanceInfo)
	m.InstanceIDDic = make(map[string]*pb.InstanceInfo)
	for _, v := range info.Instances {
		m.InstanceAddressDic[InstanceAddress(v.Host, v.Port)] = v
		m.InstanceIDDic[v.InstanceID] = v
	}
}

func newInstanceMgr(serviceName string) *InstanceMgr {
	return &InstanceMgr{
		InstanceIDDic:      make(map[string]*pb.InstanceInfo),
		InstanceAddressDic: make(map[string]*pb.InstanceInfo),
		ServiceInfo:        util.NewServiceInfo(serviceName),
	}
}

type DefaultServiceBuffer struct {
	rwMutex sync.RWMutex
	dict    map[string]*InstanceMgr
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func (m *DefaultServiceBuffer) FlushService(info *pb.ServiceInfo) {
	instanceMgr := func() *InstanceMgr {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		instanceMgr, ok := m.dict[info.ServiceName]
		if !ok {
			instanceMgr = newInstanceMgr(info.ServiceName)
			m.dict[info.ServiceName] = instanceMgr
		}
		//在释放父锁之前先把自己锁住
		instanceMgr.rwMutex.Lock()
		return instanceMgr
	}()
	defer instanceMgr.rwMutex.Unlock()
	//版本相同就不更新了
	if instanceMgr.ServiceInfo.Revision == info.Revision {
		return
	}
	instanceMgr.Reset(info)
}

func (m *DefaultServiceBuffer) GetInstanceByAddress(
	service string, address string,
) (*pb.InstanceInfo, bool) {
	instanceMgr, ok := func() (*InstanceMgr, bool) {
		m.rwMutex.RLock()
		defer m.rwMutex.RUnlock()
		instanceMgr, ok := m.dict[service]
		if ok {
			instanceMgr.rwMutex.RLock()
		}
		return instanceMgr, ok
	}()
	if !ok {
		return nil, false
	}
	defer instanceMgr.rwMutex.RUnlock()
	v, ok := instanceMgr.InstanceAddressDic[address]
	return v, ok
}

func (m *DefaultServiceBuffer) GetServiceInfo(service string) (*pb.ServiceInfo, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	ins, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	return ins.ServiceInfo, true
}

func (m *DefaultServiceBuffer) AddInstance(serviceName string, host string,
	port int32, weight int32, instanceID *string,
) (info *pb.InstanceInfo) {
	mgr := func() *InstanceMgr {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.dict[serviceName]
		if !ok {
			mgr = newInstanceMgr(serviceName)
			m.dict[serviceName] = mgr
		}
		mgr.rwMutex.Lock()
		return mgr
	}()
	defer mgr.rwMutex.Unlock()
	mgr.ServiceInfo.Revision += 1
	address := InstanceAddress(host, port)
	var nowInstanceID string
	if instanceID != nil {
		nowInstanceID = *instanceID
	} else {
		nowInstanceID = address
	}
	info = &pb.InstanceInfo{
		InstanceID: nowInstanceID,
		Host:       host,
		Port:       port,
		Weight:     weight,
	}
	mgr.InstanceAddressDic[address] = info
	mgr.InstanceIDDic[nowInstanceID] = info
	mgr.ServiceInfo.Instances = append(mgr.ServiceInfo.Instances, info)
	return
}

func (m *DefaultServiceBuffer) RemoveInstance(
	serviceName string, host string, port int32, instanceID string,
) error {
	mgr, ok := func() (*InstanceMgr, bool) {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.dict[serviceName]
		if ok {
			mgr.rwMutex.Lock()
		}
		return mgr, ok
	}()
	if !ok {
		return nil
	}
	defer mgr.rwMutex.Unlock()
	address := InstanceAddress(host, port)
	instance, ok := mgr.InstanceAddressDic[address]
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
	delete(mgr.InstanceAddressDic, address)
	delete(mgr.InstanceIDDic, instanceID)
	return nil
}

func newDefaultServiceBuffer() ServiceBuffer {
	return &DefaultServiceBuffer{
		dict: make(map[string]*InstanceMgr),
	}
}

func (m *DefaultServiceBuffer) GetInstanceByID(service string, instanceID string) (*pb.InstanceInfo, bool) {
	mgr, ok := func() (*InstanceMgr, bool) {
		m.rwMutex.RLock()
		defer m.rwMutex.RUnlock()
		mgr, ok := m.dict[service]
		if !ok {
			return nil, false
		}
		mgr.rwMutex.RLock()
		return mgr, true
	}()
	if !ok {
		return nil, false
	}
	defer mgr.rwMutex.RUnlock()
	info, ok := mgr.InstanceIDDic[instanceID]
	return info, ok
}

var (
	NewServiceBuffer = newDefaultServiceBuffer
)

// ServiceBuffer 用来快速查找当前是否存在某个Instance
type ServiceBuffer interface {
	// GetInstanceByID 获取对应服务实例
	GetInstanceByID(service string, instanceID string) (*pb.InstanceInfo, bool)
	// GetInstanceByAddress 获取对应服务实例
	GetInstanceByAddress(service string, address string) (*pb.InstanceInfo, bool)
	// GetServiceInfo 获取服务信息
	GetServiceInfo(service string) (*pb.ServiceInfo, bool)

	// AddInstance 增加一个Instance
	AddInstance(serviceName string, host string, port int32, weight int32, instanceID *string) *pb.InstanceInfo
	// FlushService 更新或者创建
	FlushService(info *pb.ServiceInfo)
	// RemoveInstance 删除一个Instance
	RemoveInstance(serviceName string, host string, port int32, instanceID string) error
}

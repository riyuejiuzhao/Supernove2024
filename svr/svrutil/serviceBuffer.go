package svrutil

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"sync"
)

type InstanceMgr struct {
	*pb.ServiceInfo
	mutex sync.Mutex
}

func (m *InstanceMgr) Reset(info *pb.ServiceInfo) {
	m.ServiceInfo = info
}

func newInstanceMgr(serviceName string) *InstanceMgr {
	return &InstanceMgr{
		ServiceInfo: util.NewServiceInfo(serviceName),
	}
}

type DefaultServiceBuffer struct {
	mutex sync.Mutex
	dict  map[string]*InstanceMgr
}

func (m *DefaultServiceBuffer) FlushService(info *pb.ServiceInfo) {
	instanceMgr := func() *InstanceMgr {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		instanceMgr, ok := m.dict[info.ServiceName]
		if !ok {
			instanceMgr = newInstanceMgr(info.ServiceName)
			m.dict[info.ServiceName] = instanceMgr
		}
		//在释放父锁之前先把自己锁住
		instanceMgr.mutex.Lock()
		return instanceMgr
	}()
	defer instanceMgr.mutex.Unlock()
	//版本相同就不更新了
	if instanceMgr.ServiceInfo.Revision == info.Revision {
		return
	}
	instanceMgr.Reset(info)
}

func (m *DefaultServiceBuffer) GetServiceInfo(service string) (*pb.ServiceInfo, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ins, ok := m.dict[service]
	if !ok {
		return nil, false
	}
	return ins.ServiceInfo, true
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
type ServiceBuffer interface {
	// GetServiceInfo 获取服务信息
	GetServiceInfo(service string) (*pb.ServiceInfo, bool)
	// FlushService 更新或者创建
	FlushService(info *pb.ServiceInfo)
}

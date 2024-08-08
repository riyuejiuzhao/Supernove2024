package svrutil

import (
	"Supernove2024/pb"
	"sync"
)

type ServiceRouterBuffer struct {
	*pb.ServiceRouterInfo
	rwMutex sync.RWMutex
}

func newRouterBuffer(name string) *ServiceRouterBuffer {
	rt := &ServiceRouterBuffer{
		ServiceRouterInfo: &pb.ServiceRouterInfo{ServiceName: name},
	}
	return rt
}

func (b *ServiceRouterBuffer) Reset(info *pb.ServiceRouterInfo) {
	b.ServiceRouterInfo = info
}

type DefaultRouterBuffer struct {
	rwMutex sync.RWMutex
	buffer  map[string]*ServiceRouterBuffer
}

func (m *DefaultRouterBuffer) FlushService(info *pb.ServiceRouterInfo) {
	routerBuffer := func() *ServiceRouterBuffer {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		instanceMgr, ok := m.buffer[info.ServiceName]
		if !ok {
			instanceMgr = newRouterBuffer(info.ServiceName)
			m.buffer[info.ServiceName] = instanceMgr
		}
		//在释放父锁之前先把自己锁住
		instanceMgr.rwMutex.Lock()
		return instanceMgr
	}()
	defer routerBuffer.rwMutex.Unlock()
	//版本相同就不更新了
	if routerBuffer.Revision == info.Revision {
		return
	}
	routerBuffer.Reset(info)
}

func (m *DefaultRouterBuffer) GetServiceRouter(service string) (*pb.ServiceRouterInfo, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	serviceInfo, ok := m.buffer[service]
	if !ok {
		return nil, false
	}
	return serviceInfo.ServiceRouterInfo, true
}

func newDefaultRouterBuffer() RouterBuffer {
	return &DefaultRouterBuffer{
		buffer: make(map[string]*ServiceRouterBuffer),
	}
}

var (
	NewRouterBuffer = newDefaultRouterBuffer
)

// RouterBuffer 用来快速查找当前是否存在某个Router
type RouterBuffer interface {
	GetServiceRouter(service string) (*pb.ServiceRouterInfo, bool)
	FlushService(info *pb.ServiceRouterInfo)
}

package svrutil

import (
	"Supernove2024/pb"
	"sync"
)

type ServiceRouterBuffer struct {
	*pb.ServiceRouterInfo

	rwMutex         sync.RWMutex
	KVRouterDic     map[string]*pb.KVRouterInfo
	TargetRouterDic map[string]*pb.TargetRouterInfo
}

func newRouterBuffer(name string) *ServiceRouterBuffer {
	rt := &ServiceRouterBuffer{
		ServiceRouterInfo: &pb.ServiceRouterInfo{ServiceName: name},
		KVRouterDic:       make(map[string]*pb.KVRouterInfo),
		TargetRouterDic:   make(map[string]*pb.TargetRouterInfo),
	}
	return rt
}

func (b *ServiceRouterBuffer) Reset(info *pb.ServiceRouterInfo) {
	b.ServiceRouterInfo = info
	clear(b.KVRouterDic)
	clear(b.TargetRouterDic)
	for _, v := range info.KVRouters {
		b.KVRouterDic[v.Key] = v
	}
	for _, v := range info.TargetRouters {
		b.TargetRouterDic[v.SrcInstanceID] = v
	}
}

type DefaultRouterBuffer struct {
	rwMutex sync.RWMutex
	buffer  map[string]*ServiceRouterBuffer
}

func (m *DefaultRouterBuffer) GetKVRouter(service string, key string) (*pb.KVRouterInfo, bool) {
	instanceMgr, ok := func() (*ServiceRouterBuffer, bool) {
		m.rwMutex.RLock()
		defer m.rwMutex.RUnlock()
		instanceMgr, ok := m.buffer[service]
		if ok {
			instanceMgr.rwMutex.RLock()
		}
		return instanceMgr, ok
	}()
	if !ok {
		return nil, false
	}
	defer instanceMgr.rwMutex.RUnlock()
	v, ok := instanceMgr.KVRouterDic[key]
	return v, ok
}

func (m *DefaultRouterBuffer) GetTargetRouter(service string, srcInstanceID string) (*pb.TargetRouterInfo, bool) {
	instanceMgr, ok := func() (*ServiceRouterBuffer, bool) {
		m.rwMutex.RLock()
		defer m.rwMutex.RUnlock()
		instanceMgr, ok := m.buffer[service]
		if ok {
			instanceMgr.rwMutex.RLock()
		}
		return instanceMgr, ok
	}()
	if !ok {
		return nil, false
	}
	defer instanceMgr.rwMutex.RUnlock()
	v, ok := instanceMgr.TargetRouterDic[srcInstanceID]
	return v, ok
}

func (m *DefaultRouterBuffer) AddKVRouter(serviceName string, key string, dstInstanceID string) *pb.KVRouterInfo {
	mgr := func() *ServiceRouterBuffer {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.buffer[serviceName]
		if !ok {
			mgr = newRouterBuffer(serviceName)
			m.buffer[serviceName] = mgr
		}
		mgr.rwMutex.Lock()
		return mgr
	}()
	defer mgr.rwMutex.Unlock()
	mgr.ServiceRouterInfo.Revision += 1
	info := &pb.KVRouterInfo{
		Key:           key,
		DstInstanceID: dstInstanceID,
	}
	mgr.KVRouterDic[key] = info
	mgr.ServiceRouterInfo.KVRouters = append(mgr.ServiceRouterInfo.KVRouters, info)
	return info
}

func (m *DefaultRouterBuffer) AddTargetRouter(serviceName string, srcInstanceID string, dstInstanceID string) *pb.TargetRouterInfo {
	mgr := func() *ServiceRouterBuffer {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.buffer[serviceName]
		if !ok {
			mgr = newRouterBuffer(serviceName)
			m.buffer[serviceName] = mgr
		}
		mgr.rwMutex.Lock()
		return mgr
	}()
	defer mgr.rwMutex.Unlock()
	mgr.ServiceRouterInfo.Revision += 1
	info := &pb.TargetRouterInfo{
		SrcInstanceID: srcInstanceID,
		DstInstanceID: dstInstanceID,
	}
	mgr.TargetRouterDic[srcInstanceID] = info
	mgr.ServiceRouterInfo.TargetRouters = append(mgr.ServiceRouterInfo.TargetRouters, info)
	return info
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

func (m *DefaultRouterBuffer) RemoveKVRouter(serviceName string, Key string) error {
	mgr, ok := func() (*ServiceRouterBuffer, bool) {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.buffer[serviceName]
		if ok {
			mgr.rwMutex.Lock()
		}
		return mgr, ok
	}()
	if !ok {
		return nil
	}
	defer mgr.rwMutex.Unlock()
	_, ok = mgr.KVRouterDic[Key]
	if !ok {
		return nil
	}
	mgr.Revision += 1
	removeIndex := 0
	for i, v := range mgr.KVRouters {
		if v.Key == Key {
			removeIndex = i
			break
		}
	}
	copy(mgr.KVRouters[removeIndex:], mgr.KVRouters[removeIndex+1:])
	mgr.KVRouters = mgr.KVRouters[:len(mgr.KVRouters)-1]
	delete(mgr.KVRouterDic, Key)
	return nil
}

func (m *DefaultRouterBuffer) RemoveTargetRouter(serviceName string, srcInstanceID string) error {
	mgr, ok := func() (*ServiceRouterBuffer, bool) {
		m.rwMutex.Lock()
		defer m.rwMutex.Unlock()
		mgr, ok := m.buffer[serviceName]
		if ok {
			mgr.rwMutex.Lock()
		}
		return mgr, ok
	}()
	if !ok {
		return nil
	}
	defer mgr.rwMutex.Unlock()
	_, ok = mgr.TargetRouterDic[srcInstanceID]
	if !ok {
		return nil
	}
	mgr.Revision += 1
	removeIndex := 0
	for i, v := range mgr.TargetRouters {
		if v.SrcInstanceID == srcInstanceID {
			removeIndex = i
			break
		}
	}
	copy(mgr.TargetRouters[removeIndex:], mgr.TargetRouters[removeIndex+1:])
	mgr.TargetRouters = mgr.TargetRouters[:len(mgr.TargetRouters)-1]
	delete(mgr.TargetRouterDic, srcInstanceID)
	return nil
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

	GetKVRouter(service string, key string) (*pb.KVRouterInfo, bool)
	GetTargetRouter(service string, srcInstanceID string) (*pb.TargetRouterInfo, bool)

	AddKVRouter(serviceName string, key string, dstInstanceID string) *pb.KVRouterInfo
	AddTargetRouter(serviceName string, srcInstanceID string, dstInstanceID string) *pb.TargetRouterInfo

	FlushService(info *pb.ServiceRouterInfo)

	RemoveKVRouter(serviceName string, Key string) error
	RemoveTargetRouter(serviceName string, srcInstanceID string) error
}

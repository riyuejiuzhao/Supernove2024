package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"log"
	"sync"
	"time"
)

type KVRouterBuffer interface {
	Add(info *pb.KVRouterInfo)
	Get(key string) (*pb.KVRouterInfo, bool)
	Remove(key string)
	Clear()
}

type ServiceRouterBuffer struct {
	Mutex           sync.Mutex
	KvRouterDic     map[string]*pb.KVRouterInfo
	TargetRouterDic map[int64]*pb.TargetRouterInfo
}

func (m *DefaultServiceMgr) GetTargetRouter(ServiceName string, SrcInstanceID int64) (*pb.TargetRouterInfo, bool) {
	b, ok := func() (b *ServiceRouterBuffer, ok bool) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok = m.routerBuffer.Value[ServiceName]
		if !ok {
			return nil, false
		}
		b.Mutex.Lock()
		return
	}()
	if !ok {
		return nil, false
	}
	defer b.Mutex.Unlock()
	info, ok := b.TargetRouterDic[SrcInstanceID]
	return info, ok
}

func (m *DefaultServiceMgr) GetKVRouter(ServiceName string, Key string) (*pb.KVRouterInfo, bool) {
	b, ok := func() (b *ServiceRouterBuffer, ok bool) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok = m.routerBuffer.Value[ServiceName]
		if !ok {
			return nil, false
		}
		b.Mutex.Lock()
		return
	}()
	if !ok {
		return nil, false
	}
	defer b.Mutex.Unlock()
	info, ok := b.KvRouterDic[Key]
	return info, ok
}

func (m *DefaultServiceMgr) AddTargetRouter(serviceName string, info *pb.TargetRouterInfo) {
	b := func() (b *ServiceRouterBuffer) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			b = &ServiceRouterBuffer{
				KvRouterDic:     make(map[string]*pb.KVRouterInfo),
				TargetRouterDic: make(map[int64]*pb.TargetRouterInfo),
			}
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	nowInfo, ok := b.TargetRouterDic[info.SrcInstanceID]
	//保留更新的那个
	if ok && nowInfo.CreateTime > info.CreateTime {
		return
	}
	b.TargetRouterDic[info.SrcInstanceID] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) AddKVRouter(serviceName string, info *pb.KVRouterInfo) {
	b := func() (b *ServiceRouterBuffer) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			b = &ServiceRouterBuffer{
				KvRouterDic:     make(map[string]*pb.KVRouterInfo),
				TargetRouterDic: make(map[int64]*pb.TargetRouterInfo),
			}
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	nowInfo, ok := b.KvRouterDic[info.Key]
	//保留更新的那个
	if ok && nowInfo.CreateTime > info.CreateTime {
		return
	}
	b.KvRouterDic[info.Key] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) RemoveKVRouter(ServiceName string, Key string) {
	b, ok := func() (b *ServiceRouterBuffer, ok bool) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok = m.routerBuffer.Value[ServiceName]
		if !ok {
			return nil, false
		}
		b.Mutex.Lock()
		return
	}()
	if !ok {
		return
	}
	defer b.Mutex.Unlock()
	delete(b.KvRouterDic, Key)
}

func (m *DefaultServiceMgr) RemoveTargetRouter(ServiceName string, SrcInstanceID int64) {
	b, ok := func() (b *ServiceRouterBuffer, ok bool) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok = m.routerBuffer.Value[ServiceName]
		if !ok {
			return nil, false
		}
		b.Mutex.Lock()
		return
	}()
	if !ok {
		return
	}
	defer b.Mutex.Unlock()
	delete(b.TargetRouterDic, SrcInstanceID)
}

func (m *DefaultServiceMgr) handleKVRouterDelete(serviceName string, ev *clientv3.Event) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "KVRouterDelete"}, nil)
	}()
	key := util.KVRouterKey2Key(string(ev.Kv.Key), serviceName)
	m.RemoveKVRouter(serviceName, key)
}

func (m *DefaultServiceMgr) handleKVRouterPut(serviceName string, ev *clientv3.Event) (err error) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "KVRouterPut"}, err)
	}()
	info := &pb.KVRouterInfo{}
	err = proto.Unmarshal(ev.Kv.Value, info)
	if err != nil {
		util.Error("err %v", err)
		return
	}
	m.AddKVRouter(serviceName, info)
	return
}

func (m *DefaultServiceMgr) handleWatchKVRouter(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	rch := cli.Watch(context.Background(), util.RouterKVPrefix(serviceName), clientv3.WithPrefix())
	go m.InitKVRouter(serviceName)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				err = m.handleKVRouterPut(serviceName, ev)
				if err != nil {
					util.Error("kv router err: %v", err)
				}
			case clientv3.EventTypeDelete:
			}
		}
	}
}

func (m *DefaultServiceMgr) handleWatchTargetRouter(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	rch := cli.Watch(context.Background(), util.RouterTargetPrefix(serviceName), clientv3.WithPrefix())
	go m.InitTargetRouter(serviceName)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				info := &pb.TargetRouterInfo{}
				err = proto.Unmarshal(ev.Kv.Value, info)
				if err != nil {
					util.Error("err %v", err)
					continue
				}
				m.AddTargetRouter(serviceName, info)
			case clientv3.EventTypeDelete:
				id, err := util.TargetRouterKey2InstanceID(string(ev.Kv.Key), serviceName)
				if err != nil {
					util.Error("err %v", err)
					continue
				}
				m.RemoveTargetRouter(serviceName, id)
			}
		}
	}
}

func (m *DefaultServiceMgr) InitKVRouter(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Get(context.Background(), util.RouterKVPrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
		util.Error("%v", err)
		return
	}
	for _, kv := range resp.Kvs {
		info := &pb.KVRouterInfo{}
		err = proto.Unmarshal(kv.Value, info)
		if err != nil {
			util.Error("%v", err)
			continue
		}
		m.AddKVRouter(serviceName, info)
	}
}

func (m *DefaultServiceMgr) InitTargetRouter(serviceName string) {
	cli, err := m.connManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.Get(context.Background(), util.RouterTargetPrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
		util.Error("%v", err)
		return
	}
	for _, kv := range resp.Kvs {
		info := &pb.TargetRouterInfo{}
		err = proto.Unmarshal(kv.Value, info)
		if err != nil {
			util.Error("%v", err)
			continue
		}
		m.AddTargetRouter(serviceName, info)
	}
}

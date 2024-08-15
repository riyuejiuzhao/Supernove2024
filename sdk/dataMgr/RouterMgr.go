package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"context"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
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
	Mutex sync.Mutex

	KvRouterIdDic map[int64]*pb.KVRouterInfo
	KvRouterDic   map[string]*pb.KVRouterInfo

	TargetRouterIdDic map[int64]*pb.TargetRouterInfo
	TargetRouterDic   map[string]*pb.TargetRouterInfo
}

func (m *DefaultServiceMgr) GetTargetRouter(ServiceName string, SrcInstanceName string) (*pb.TargetRouterInfo, bool) {
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
	info, ok := b.TargetRouterDic[SrcInstanceName]
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
				KvRouterDic:       make(map[string]*pb.KVRouterInfo),
				KvRouterIdDic:     make(map[int64]*pb.KVRouterInfo),
				TargetRouterDic:   make(map[string]*pb.TargetRouterInfo),
				TargetRouterIdDic: make(map[int64]*pb.TargetRouterInfo),
			}
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	nowInfo, ok := b.TargetRouterDic[info.SrcInstanceName]
	//保留更新的那个
	if ok && nowInfo.CreateTime > info.CreateTime {
		return
	}
	if ok {
		delete(b.TargetRouterIdDic, nowInfo.RouterID)
	}
	b.TargetRouterDic[info.SrcInstanceName] = info
	b.TargetRouterIdDic[info.RouterID] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) AddKVRouter(serviceName string, info *pb.KVRouterInfo) {
	b := func() (b *ServiceRouterBuffer) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			b = &ServiceRouterBuffer{
				KvRouterDic:       make(map[string]*pb.KVRouterInfo),
				KvRouterIdDic:     make(map[int64]*pb.KVRouterInfo),
				TargetRouterDic:   make(map[string]*pb.TargetRouterInfo),
				TargetRouterIdDic: make(map[int64]*pb.TargetRouterInfo),
			}
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	dic := make(map[string]string)
	for i := 0; i < len(info.Key); i++ {
		dic[info.Key[i]] = info.Val[i]
	}
	js, err := json.Marshal(dic)
	if err != nil {
		util.Error("add kv router failed", err)
		return
	}
	jss := string(js)
	nowInfo, ok := b.KvRouterDic[jss]
	//保留更新的那个
	if ok && nowInfo.CreateTime > info.CreateTime {
		return
	}
	if ok {
		delete(b.KvRouterIdDic, nowInfo.RouterID)
	}
	b.KvRouterDic[jss] = info
	b.KvRouterIdDic[info.RouterID] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) RemoveKVRouter(ServiceName string, RouterID int64) {
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
	info, ok := b.KvRouterIdDic[RouterID]
	dic := make(map[string]string)
	for i := 0; i < len(info.Key); i++ {
		dic[info.Key[i]] = info.Val[i]
	}
	js, err := json.Marshal(dic)
	if err != nil {
		util.Error("remove kv router failed", err)
		return
	}
	jss := string(js)
	delete(b.KvRouterDic, jss)
	delete(b.KvRouterIdDic, RouterID)
	util.Info("%s Delete KV Router %s", ServiceName, jss)
}

func (m *DefaultServiceMgr) RemoveTargetRouter(ServiceName string, RouterID int64) {
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
	info, ok := b.TargetRouterIdDic[RouterID]
	if !ok {
		return
	}
	delete(b.TargetRouterIdDic, RouterID)
	delete(b.TargetRouterDic, info.SrcInstanceName)
	util.Info("%s Delete Target Router %s", ServiceName, info.SrcInstanceName)
}

func (m *DefaultServiceMgr) handleKVRouterDelete(serviceName string, ev *clientv3.Event) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "KVRouterDelete"}, nil)
	}()
	key, err := util.KVRouterKey2RouterID(string(ev.Kv.Key), serviceName)
	if err != nil {
		util.Error("%s delete kv router err %v", serviceName, err)
		return
	}
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

func (m *DefaultServiceMgr) handleWatchKVRouter(cli *clientv3.Client, serviceName string) {
	revision, err := m.initKVRouter(cli, serviceName)
	if err != nil {
		util.Error("KV路由获取错误", err)
	}
	rch := cli.Watch(context.Background(),
		util.RouterKVPrefix(serviceName),
		clientv3.WithPrefix(),
		clientv3.WithRev(revision),
	)
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					err := m.handleKVRouterPut(serviceName, ev)
					if err != nil {
						util.Error("kv router err: %v", err)
					}
				case clientv3.EventTypeDelete:
					id, err := util.KVRouterKey2RouterID(string(ev.Kv.Key), serviceName)
					if err != nil {
						util.Error("kv router delete err: %v", err)
					}
					m.RemoveKVRouter(serviceName, id)
				}
			}
		}
	}()

}

func (m *DefaultServiceMgr) handleWatchTargetRouter(cli *clientv3.Client, serviceName string) {
	revision, err := m.initTargetRouter(cli, serviceName)
	if err != nil {
		util.Error("获取全量TargetRouter出错", err)
	}
	rch := cli.Watch(context.Background(),
		util.RouterTargetPrefix(serviceName),
		clientv3.WithPrefix(),
		clientv3.WithRev(revision))
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					info := &pb.TargetRouterInfo{}
					err := proto.Unmarshal(ev.Kv.Value, info)
					if err != nil {
						util.Error("err %v", err)
						continue
					}
					m.AddTargetRouter(serviceName, info)
				case clientv3.EventTypeDelete:
					id, err := util.TargetRouterKey2RouterID(string(ev.Kv.Key), serviceName)
					if err != nil {
						util.Error("解析RouterID err: %v", err)
						continue
					}
					m.RemoveTargetRouter(serviceName, id)
				}
			}
		}
	}()
}

func (m *DefaultServiceMgr) initKVRouter(cli *clientv3.Client, serviceName string) (revision int64, err error) {
	resp, err := cli.Get(context.Background(), util.RouterKVPrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
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
	err = nil
	revision = resp.Header.Revision
	return
}

func (m *DefaultServiceMgr) initTargetRouter(cli *clientv3.Client, serviceName string) (revision int64, err error) {
	resp, err := cli.Get(context.Background(), util.RouterTargetPrefix(serviceName), clientv3.WithPrefix())
	if err != nil {
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
	err = nil
	revision = resp.Header.Revision
	return
}

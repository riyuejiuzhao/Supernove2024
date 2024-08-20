package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type RouterMapNode struct {
	KVRouterInfo []*pb.KVRouterInfo
	Nodes        map[string]*RouterMapNode
}

func newRouterMapNode() *RouterMapNode {
	return &RouterMapNode{
		KVRouterInfo: make([]*pb.KVRouterInfo, 0),
		Nodes:        make(map[string]*RouterMapNode),
	}
}

func (n *RouterMapNode) Add(index int, Tags []string, info *pb.KVRouterInfo) {
	if index >= len(Tags) {
		n.KVRouterInfo = append(n.KVRouterInfo, info)
		return
	}
	nowTag := Tags[index]
	value, ok := info.Dic[nowTag]
	if !ok {
		n.KVRouterInfo = append(n.KVRouterInfo, info)
		return
	}
	next, ok := n.Nodes[value]
	if !ok {
		next = newRouterMapNode()
		n.Nodes[value] = next
	}
	next.Add(index+1, Tags, info)
}

func (n *RouterMapNode) Find(index int, Tags []string, Keys map[string]string) ([]*pb.KVRouterInfo, bool) {
	if index >= len(Tags) {
		return n.KVRouterInfo, len(n.KVRouterInfo) != 0
	}

	nowTag := Tags[index]
	value, ok := Keys[nowTag]
	if !ok {
		return n.KVRouterInfo, len(n.KVRouterInfo) != 0
	}
	next, ok := n.Nodes[value]
	if !ok {
		return n.KVRouterInfo, len(n.KVRouterInfo) != 0
	}
	result, ok := next.Find(index+1, Tags, Keys)
	if !ok {
		return n.KVRouterInfo, len(n.KVRouterInfo) != 0
	}
	return result, ok
}

// Remove 返回值通知父节点自身可否删除
func (n *RouterMapNode) Remove(index int, Tags []string, info *pb.KVRouterInfo) bool {
	if index >= len(Tags) {
		i := 0
		for ; i < len(n.KVRouterInfo); i++ {
			if n.KVRouterInfo[i].RouterID != info.RouterID {
				continue
			}
			n.KVRouterInfo = append(n.KVRouterInfo[:i], n.KVRouterInfo[i+1:]...)
			break
		}
		return len(n.KVRouterInfo) == 0
	}
	nowTag := Tags[index]
	value := info.Dic[nowTag]
	next, ok := n.Nodes[value]
	//没有能够完全匹配的，那就不删除任何路由
	if !ok {
		return false
	}
	ok = next.Remove(index+1, Tags, info)
	//子路由存在不能删除的
	//那么自身也不能删除
	if !ok {
		return false
	}
	//子路由可以删除，先删除子路由
	delete(n.Nodes, value)
	//下面没有任何子路由了，本身也不是路由就可以让父节点把自己也删掉
	return len(n.Nodes) == 0 && len(n.KVRouterInfo) == 0
}

type RouterMapRoot struct {
	Info *pb.RouterTableInfo
	Root *RouterMapNode
}

type ServiceRouterBuffer struct {
	Mutex sync.Mutex

	KvRouterTable *RouterMapRoot

	KvRouterIdDic map[int64]*pb.KVRouterInfo

	TargetRouterIdDic map[int64]*pb.TargetRouterInfo
	TargetRouterDic   map[string]*pb.TargetRouterInfo
}

func (b *RouterMapRoot) findKVRouter(Keys map[string]string) ([]*pb.KVRouterInfo, bool) {
	if b.Info == nil {
		util.Warn("查询路由缺少路由表")
		return nil, false
	}
	return b.Root.Find(0, b.Info.Tags, Keys)
}

func (b *RouterMapRoot) removeKVRouter(info *pb.KVRouterInfo) {
	if b.Info == nil {
		util.Warn("删除路由缺少路由表")
		return
	}
	b.Root.Remove(0, b.Info.Tags, info)
}

func (b *RouterMapRoot) addKVRouter(info *pb.KVRouterInfo) {
	if b.Info == nil {
		util.Warn("添加路由缺少路由表")
		return
	}
	b.Root.Add(0, b.Info.Tags, info)
}

func newServiceRouterBuffer(info *pb.RouterTableInfo) (buffer *ServiceRouterBuffer) {
	buffer = &ServiceRouterBuffer{
		KvRouterTable: &RouterMapRoot{
			Info: info,
			Root: newRouterMapNode(),
		},
		KvRouterIdDic:     make(map[int64]*pb.KVRouterInfo),
		TargetRouterDic:   make(map[string]*pb.TargetRouterInfo),
		TargetRouterIdDic: make(map[int64]*pb.TargetRouterInfo),
	}
	return
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

func (m *DefaultServiceMgr) GetKVRouter(ServiceName string, Keys map[string]string) ([]*pb.KVRouterInfo, bool) {
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
	return b.KvRouterTable.findKVRouter(Keys)
}

func (m *DefaultServiceMgr) AddTargetRouter(serviceName string, info *pb.TargetRouterInfo) {
	if m.SkipSave {
		return
	}
	b := func() (b *ServiceRouterBuffer) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			b = newServiceRouterBuffer(nil)
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	nowInfo, ok := b.TargetRouterDic[info.SrcInstanceName]
	if ok {
		delete(b.TargetRouterIdDic, nowInfo.RouterID)
	}
	b.TargetRouterDic[info.SrcInstanceName] = info
	b.TargetRouterIdDic[info.RouterID] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) AddRouterTable(info *pb.RouterTableInfo) {
	if m.SkipSave {
		return
	}
	b, ok := func() (b *ServiceRouterBuffer, ok bool) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok = m.routerBuffer.Value[info.ServiceName]
		if !ok {
			b = newServiceRouterBuffer(info)
			m.routerBuffer.Value[info.ServiceName] = b
		} else {
			b.Mutex.Lock()
		}
		return
	}()
	if !ok {
		util.Info("添加路由表%s", info.ServiceName)
		return
	}
	//原本存在需要重新加载一次路由了
	defer b.Mutex.Unlock()
	b.KvRouterTable.Info = info
	for _, kvInfo := range b.KvRouterIdDic {
		b.KvRouterTable.addKVRouter(kvInfo)
	}
	util.Info("添加路由表%s", info.ServiceName)
}

func (m *DefaultServiceMgr) AddKVRouter(serviceName string, info *pb.KVRouterInfo) {
	if m.SkipSave {
		return
	}
	b := func() (b *ServiceRouterBuffer) {
		m.routerBuffer.Mutex.Lock()
		defer m.routerBuffer.Mutex.Unlock()
		b, ok := m.routerBuffer.Value[serviceName]
		if !ok {
			util.Warn("路由表注册前，发生了KV 路由注册")
			b = newServiceRouterBuffer(nil)
			m.routerBuffer.Value[serviceName] = b
		}
		b.Mutex.Lock()
		return
	}()
	defer b.Mutex.Unlock()
	b.KvRouterTable.addKVRouter(info)
	b.KvRouterIdDic[info.RouterID] = info
	util.Info("%s Add Router %s", serviceName, info)
}

func (m *DefaultServiceMgr) RemoveRouterTable(ServiceName string) {
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
	b.KvRouterTable.Info = nil
	util.Info("%s Delete Router Table", ServiceName)
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
	if !ok {
		return
	}
	b.KvRouterTable.removeKVRouter(info)
	delete(b.KvRouterIdDic, RouterID)
	util.Info("%s Delete KV Router %s", ServiceName, info.Dic)
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

func (m *DefaultServiceMgr) handleRouterDelete(serviceName string, ev *clientv3.Event) (err error) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "RouterDelete"}, err)
	}()

	var id int64
	key := string(ev.Kv.Key)
	if util.IsRouterTable(serviceName, key) {
		m.RemoveRouterTable(serviceName)
		return
	} else if util.IsKVRouter(serviceName, key) {
		id, err = util.KVRouterKey2RouterID(string(ev.Kv.Key), serviceName)
		if err != nil {
			util.Error("%s delete kv router err %v", serviceName, err)
			return
		}
		m.RemoveKVRouter(serviceName, id)
		return
	} else if util.IsDstRouter(serviceName, key) {
		id, err = util.TargetRouterKey2RouterID(string(ev.Kv.Key), serviceName)
		if err != nil {
			util.Error("解析RouterID err: %v", err)
			return
		}
		m.RemoveTargetRouter(serviceName, id)
		return
	}
	err = fmt.Errorf("不识别的路由类型")
	return
}

func (m *DefaultServiceMgr) handleRouterPut(serviceName string, ev *clientv3.Event) (err error) {
	begin := time.Now()
	defer func() {
		m.mt.MetricsUpload(begin, prometheus.Labels{"Method": "RouterPut"}, err)
	}()
	key := string(ev.Kv.Key)
	if util.IsRouterTable(serviceName, key) {
		info := &pb.RouterTableInfo{}
		err = proto.Unmarshal(ev.Kv.Value, info)
		if err != nil {
			return
		}
		m.AddRouterTable(info)
		return
	} else if util.IsKVRouter(serviceName, key) {
		info := &pb.KVRouterInfo{}
		err = proto.Unmarshal(ev.Kv.Value, info)
		if err != nil {
			return
		}
		m.AddKVRouter(serviceName, info)
		return
	} else if util.IsDstRouter(serviceName, key) {
		info := &pb.TargetRouterInfo{}
		err = proto.Unmarshal(ev.Kv.Value, info)
		if err != nil {
			return
		}
		m.AddTargetRouter(serviceName, info)
		return
	}
	err = fmt.Errorf("不识别的路由类型")
	return
}

func (m *DefaultServiceMgr) handleWatchRouter(cli *clientv3.Client, serviceName string) {
	revision, err := m.initRouter(cli, serviceName)
	if err != nil {
		util.Error("路由获取错误 err: %v", err)
	}
	go func() {
		rch := cli.Watch(context.Background(),
			util.ServiceAllInfoPrefix(serviceName),
			clientv3.WithPrefix(),
			clientv3.WithRev(revision),
		)
		for wresp := range rch {
			if m.SkipSave {
				continue
			}
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					err := m.handleRouterPut(serviceName, ev)
					if err != nil {
						util.Error("router put err: %v", err)
					}
				case clientv3.EventTypeDelete:
					err := m.handleRouterDelete(serviceName, ev)
					if err != nil {
						util.Error("router del err: %v", err)
					}
				}
			}
		}
	}()
}

func (m *DefaultServiceMgr) initRouter(cli *clientv3.Client, serviceName string) (revision int64, err error) {
	resp, err := cli.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kv := range resp.Kvs {
		if util.IsRouterTable(serviceName, string(kv.Key)) {
			info := &pb.RouterTableInfo{}
			err = proto.Unmarshal(kv.Value, info)
			if err != nil {
				util.Error("%v", err)
				continue
			}
			m.AddRouterTable(info)
		} else if util.IsKVRouter(serviceName, string(kv.Key)) {
			info := &pb.KVRouterInfo{}
			err = proto.Unmarshal(kv.Value, info)
			if err != nil {
				util.Error("%v", err)
				continue
			}
			m.AddKVRouter(serviceName, info)
		} else if util.IsDstRouter(serviceName, string(kv.Key)) {
			info := &pb.TargetRouterInfo{}
			err = proto.Unmarshal(kv.Value, info)
			if err != nil {
				util.Error("%v", err)
				continue
			}
			m.AddTargetRouter(serviceName, info)
		} else {
			util.Error("不识别的Key %v", string(kv.Key))
		}

	}
	err = nil
	revision = resp.Header.Revision
	return
}

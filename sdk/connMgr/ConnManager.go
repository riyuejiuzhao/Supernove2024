package connMgr

import (
	"Supernove2024/sdk/config"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
)

type ServiceType int32

const (
	InstancesEtcd = iota
	RoutersEtcd
	ServiceTypeCount
)

// ConnManager 管理链接
type ConnManager interface {
	GetAllServiceConn(service ServiceType) []*clientv3.Client
	// GetServiceConn 指定服务的链接
	GetServiceConn(service ServiceType, key string) (*clientv3.Client, error)
}

var (
	connMgr        ConnManager = nil
	NewConnManager             = newDefaultConnManager
)

var connMutex sync.Mutex

func Instance() (ConnManager, error) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if connMgr == nil {
		cfg, err := config.GlobalConfig()
		if err != nil {
			return nil, errors.New("创建连接管理器失败")
		}
		newMgr, err := NewConnManager(cfg)
		if err != nil {
			return nil, err
		}
		connMgr = newMgr
	}
	return connMgr, nil
}

package connMgr

import (
	"Supernove2024/sdk/config"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceType int32

const (
	Etcd = iota
	ServiceTypeCount
)

// ConnManager 管理链接
type ConnManager interface {
	// GetServiceConn 指定服务的链接
	GetServiceConn(service ServiceType) (*clientv3.Client, error)
}

var (
	connMgr        ConnManager = nil
	NewConnManager             = newDefaultConnManager
)

func Instance() (ConnManager, error) {
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

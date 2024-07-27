package connMgr

import (
	"Supernove2024/sdk/config"
	"errors"
	"github.com/shimingyah/pool"
)

type ServiceType int32

const (
	Discovery ServiceType = iota
	HealthCheck
	Register

	ServiceTypeCount
)

// ConnManager 管理GRPC链接
type ConnManager interface {
	// GetServiceConn 指定服务的链接
	GetServiceConn(service ServiceType) (pool.Conn, error)
	// GetConn 指定地址的链接
	GetConn(service ServiceType, address string) (pool.Conn, error)
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
		connMgr = NewConnManager(cfg)
	}
	return connMgr, nil
}

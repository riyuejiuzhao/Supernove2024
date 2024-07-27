package connMgr

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/util"
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

type AddressPoolDic map[string]pool.Pool

type DefaultConnManager struct {
	config  *config.Config
	poolDic map[ServiceType]AddressPoolDic
}

// ConnManager 管理GRPC链接
type ConnManager interface {
	// GetServiceConn 指定服务的链接
	GetServiceConn(service ServiceType) (pool.Conn, error)
	// GetConn 指定地址的链接
	GetConn(service ServiceType, address string) (pool.Conn, error)
}

var connMgr ConnManager = nil
var NewConnManager = 

func Instance() (ConnManager, error) {
	if connMgr == nil {
		cfg, err := config.GlobalConfig()
		if err != nil {
			return nil, errors.New("创建连接管理器失败")
		}
		connMgr = newDefaultConnManager(cfg.Global.RegisterService,
			cfg.Global.DiscoverService, cfg.Global.HealthService)
	}
	return connMgr, nil
}

func newAddressPoolDic(serviceConfig []config.ServiceConfig) AddressPoolDic {
	poolDic := make(map[string]pool.Pool)
	for _, service := range serviceConfig {
		address := service.String()
		p, err := pool.New(service.String(), DefaultOptions)
		if err != nil {
			util.Error("创建连接池失败，地址：%s", address)
			continue
		}
		util.Info("创建连接池，地址: %s", address)
		poolDic[address] = p
	}
	return poolDic
}

func newDefaultConnManager(register []config.ServiceConfig,
	discovery []config.ServiceConfig,
	health []config.ServiceConfig) ConnManager {
	poolDic := make(map[ServiceType]AddressPoolDic)
	poolDic[Discovery] = newAddressPoolDic(discovery)
	poolDic[Register] = newAddressPoolDic(register)
	poolDic[HealthCheck] = newAddressPoolDic(health)
	util.Info("初始化GrpcConnManager成功")
	return &DefaultConnManager{poolDic: poolDic}
}

// GetServiceConn 指定服务的链接
func (m *DefaultConnManager) GetServiceConn(service ServiceType) (pool.Conn, error) {
	if service >= ServiceTypeCount {
		return nil, errors.New("指定的服务不存在")
	}
	servicesDic := m.poolDic[service]
	p := util.RandomDicValue(servicesDic)
	return p.Get()
}

// GetConn 指定地址的链接
func (m *DefaultConnManager) GetConn(service ServiceType, address string) (pool.Conn, error) {
	p, ok := m.poolDic[service][address]
	if !ok {
		newPool, err := pool.New(address, DefaultOptions)
		if err != nil {
			return nil, err
		}
		util.Info("创建连接池 %s", address)
		m.poolDic[service][address] = newPool
		p = newPool
	}
	conn, err := p.Get()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

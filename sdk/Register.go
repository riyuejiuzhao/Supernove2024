package sdk

import (
	"Supernove2024/miniRouterProto"
	config2 "Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/util"
	"context"
)

type RegisterCli struct {
	config      *config2.Config
	connManager *connMgr.GrpcConnManager
}

type RegisterInstanceResult struct {
	InstanceID string
	Existed    bool
}

func (c *RegisterCli) Register(service ServiceBaseInfo) (*RegisterInstanceResult, error) {
	conn, err := c.connManager.GetServiceConn(connMgr.Register)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rpcCli := miniRouterProto.NewRegisterServiceClient(conn.Value())
	request := miniRouterProto.RegisterRequest{
		ServiceName: service.Name,
		Host:        service.Host,
		Port:        service.Port,
		Weight:      c.config.Global.Register.DefaultWeight}
	reply, err := rpcCli.Register(context.Background(), &request)
	if err != nil {
		return nil, err
	}
	util.Info("注册服务: Name: %v, Host: %v, Port: %v, Weight: %v, InstanceID: %v, Exited: %v",
		request.ServiceName, request.Host, request.Port, request.Weight,
		reply.InstanceID, reply.Existed)
	return &RegisterInstanceResult{InstanceID: reply.InstanceID, Existed: reply.Existed}, nil
}

func (c *RegisterCli) Deregister(service ServiceBaseInfo) error {
	conn, err := c.connManager.GetServiceConn(connMgr.Register)
	if err != nil {
		return err
	}
	defer conn.Close()
	rpcCli := miniRouterProto.NewRegisterServiceClient(conn.Value())
	request := miniRouterProto.DeregisterRequest{
		ServiceName: service.Name,
		Host:        service.Host,
		Port:        service.Port,
	}
	_, err = rpcCli.Deregister(context.Background(), &request)
	if err != nil {
		return err
	}
	util.Info("注销服务: Name: %v, Host: %v, Port: %v, InstanceID: %v, Exited: %v",
		request.ServiceName, request.Host, request.Port)
	return nil
}

// RegisterAPI 功能：
// 服务注册
type RegisterAPI interface {
	Register(service ServiceBaseInfo) (*RegisterInstanceResult, error)
	Deregister(service ServiceBaseInfo) error
}

func NewRegisterAPI() (RegisterAPI, error) {
	globalConfig, err := config2.GlobalConfig()
	if err != nil {
		return nil, err
	}
	//随机选择一个配置中的DiscoverySvr
	connManager, err := connMgr.Instance()
	if err != nil {
		return nil, err
	}
	api := RegisterCli{connManager: connManager,
		config: globalConfig}
	util.Info("成功创建RegisterAPI")
	return &api, nil
}

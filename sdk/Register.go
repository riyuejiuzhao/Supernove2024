package sdk

import (
	"Supernove2024/miniRouterProto"
	"context"
)

type RegisterCli struct {
	registerSvrAddress string
	config             *Config
	connManager        *GrpcConnManager
}

type RegisterInstanceResult struct {
	InstanceID string
	Existed    bool
}

type ServiceInfo struct {
	Name string
	Host string
	Port int32
}

func (c *RegisterCli) Register(service ServiceInfo) (*RegisterInstanceResult, error) {
	conn, err := c.connManager.GetConn(c.registerSvrAddress)
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
	Info("注册服务: Name: %v, Host: %v, Port: %v, Weight: %v, InstanceID: %v, Exited: %v",
		request.ServiceName, request.Host, request.Port, request.Weight,
		reply.InstanceID, reply.Existed)
	return &RegisterInstanceResult{InstanceID: reply.InstanceID, Existed: reply.Existed}, nil
}

func (c *RegisterCli) Deregister(service ServiceInfo) error {
	conn, err := c.connManager.GetConn(c.registerSvrAddress)
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
	Info("注销服务: Name: %v, Host: %v, Port: %v, InstanceID: %v, Exited: %v",
		request.ServiceName, request.Host, request.Port)
	return nil
}

// RegisterAPI 功能：
// 服务注册
type RegisterAPI interface {
	Register(service ServiceInfo) (*RegisterInstanceResult, error)
	Deregister(service ServiceInfo) error
}

func NewRegisterAPI() (RegisterAPI, error) {
	globalConfig, err := GlobalConfig()
	if err != nil {
		return nil, err
	}
	//随机选择一个配置中的DiscoverySvr
	serviceHost := RandomItem(globalConfig.Global.RegisterService)
	api := RegisterCli{registerSvrAddress: serviceHost.String(),
		connManager: NewConnManager(globalConfig.Global.RegisterService),
		config:      globalConfig}
	Info("成功创建RegisterAPI，连接到%s", api.registerSvrAddress)
	return &api, nil
}

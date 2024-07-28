package sdk

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/util"
	"context"
)

type RegisterCli struct {
	APIContext
}

type RegisterArgv struct {
	ServiceName string
	Host        string
	Port        int32
	Weight      int32
}

type RegisterResult struct {
	InstanceID string
	Existed    bool
}

type DeregisterArgv struct {
	ServiceName string
	Host        string
	Port        int32
	InstanceID  string
}

func (c *RegisterCli) Register(service *RegisterArgv) (*RegisterResult, error) {
	conn, err := c.ConnManager.GetServiceConn(connMgr.Register)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rpcCli := miniRouterProto.NewRegisterServiceClient(conn.Value())
	var weight int32
	if service.Weight > 0 {
		weight = service.Weight
	} else {
		weight = c.Config.Global.Register.DefaultWeight
	}
	request := miniRouterProto.RegisterRequest{
		ServiceName: service.ServiceName,
		Host:        service.Host,
		Port:        service.Port,
		Weight:      weight,
	}
	reply, err := rpcCli.Register(context.Background(), &request)
	if err != nil {
		return nil, err
	}
	util.Info("注册服务: ServiceName: %v, Host: %v, Port: %v, Weight: %v, InstanceID: %v, Exited: %v",
		request.ServiceName, request.Host, request.Port, request.Weight,
		reply.InstanceID, reply.Existed)
	return &RegisterResult{InstanceID: reply.InstanceID, Existed: reply.Existed}, nil
}

func (c *RegisterCli) Deregister(service *DeregisterArgv) error {
	conn, err := c.ConnManager.GetServiceConn(connMgr.Register)
	if err != nil {
		return err
	}
	defer conn.Close()
	rpcCli := miniRouterProto.NewRegisterServiceClient(conn.Value())
	request := miniRouterProto.DeregisterRequest{
		ServiceName: service.ServiceName,
		Host:        service.Host,
		Port:        service.Port,
		InstanceID:  service.InstanceID,
	}
	_, err = rpcCli.Deregister(context.Background(), &request)
	if err != nil {
		return err
	}
	util.Info("注销服务: ServiceName: %s, Host: %s, Port: %v, InstanceID: %s",
		request.ServiceName, request.Host, request.Port, request.InstanceID)
	return nil
}

// RegisterAPI 功能：
// 服务注册
type RegisterAPI interface {
	Register(service *RegisterArgv) (*RegisterResult, error)
	Deregister(service *DeregisterArgv) error
}

func NewRegisterAPI() (RegisterAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	return &RegisterCli{*ctx}, nil
}

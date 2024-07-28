package main

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
)

type RegisterContext struct {
	Ctx         context.Context
	Request     *miniRouterProto.RegisterRequest
	ServiceHash string
	Address     string
}

func (c *RegisterContext) GetContext() context.Context {
	return c.Ctx
}
func (c *RegisterContext) GetServiceName() string {
	return c.Request.ServiceName
}
func (c *RegisterContext) GetServiceHash() string {
	return c.ServiceHash
}

func newRegisterContext(ctx context.Context, request *miniRouterProto.RegisterRequest) *RegisterContext {
	return &RegisterContext{
		Ctx:         ctx,
		ServiceHash: ServiceHash(request.ServiceName),
		Request:     request,
		Address:     InstanceAddress(request.Host, request.Port),
	}
}

// Register 注册一个新服务实例
func (r *RegisterSvr) Register(
	ctx context.Context,
	request *miniRouterProto.RegisterRequest,
) (*miniRouterProto.RegisterReply, error) {
	registerCtx := newRegisterContext(ctx, request)

	mutex := r.rMutex.NewMutex(ServiceInfoLockName(request.ServiceName))
	err := mutex.Lock()
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	defer util.TryUnlock(mutex)

	err = r.flushBuffer(registerCtx)
	if err != nil {
		return nil, err
	}

	instanceInfo, ok := r.mgr.TryGetInstanceByAddress(request.ServiceName, registerCtx.Address)
	if ok {
		//如果存在，那么就直接返回
		return &miniRouterProto.RegisterReply{
			InstanceID: instanceInfo.InstanceID,
			Existed:    true,
		}, nil
	}

	//如果不存在，那么添加新实例
	instanceInfo = r.mgr.AddInstance(request.ServiceName, request.Host, request.Port, request.Weight)
	//获取调整后的结果
	serviceInfo, ok := r.mgr.TryGetServiceInfo(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("没有成功创建ServiceInfo, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}
	//写入redis
	err = r.sendServiceToRedis(registerCtx, serviceInfo)
	if err != nil {
		return nil, err
	}

	return &miniRouterProto.RegisterReply{InstanceID: instanceInfo.InstanceID, Existed: false}, nil
}

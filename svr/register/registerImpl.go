package register

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
)

type RegisterContext struct {
	Request     *miniRouterProto.RegisterRequest
	ServiceHash string
	Address     string
}

func (c *RegisterContext) GetServiceName() string {
	return c.Request.ServiceName
}
func (c *RegisterContext) GetServiceHash() string {
	return c.ServiceHash
}

func newRegisterContext(request *miniRouterProto.RegisterRequest) *RegisterContext {
	return &RegisterContext{
		ServiceHash: svrutil.ServiceHash(request.ServiceName),
		Request:     request,
		Address:     svrutil.InstanceAddress(request.Host, request.Port),
	}
}

// Register 注册一个新服务实例
func (r *Server) Register(
	_ context.Context,
	request *miniRouterProto.RegisterRequest,
) (*miniRouterProto.RegisterReply, error) {
	registerCtx := newRegisterContext(request)

	mutex, err := r.LockRedisService(request.ServiceName)
	if err != nil {
		return nil, err
	}
	defer util.TryUnlock(mutex)

	err = r.FlushBuffer(registerCtx)
	if err != nil {
		return nil, err
	}

	instanceInfo, ok := r.Mgr.TryGetInstanceByAddress(request.ServiceName, registerCtx.Address)
	if ok {
		//如果存在，那么就直接返回
		return &miniRouterProto.RegisterReply{
			InstanceID: instanceInfo.InstanceID,
			Existed:    true,
		}, nil
	}

	//如果不存在，那么添加新实例
	instanceInfo = r.Mgr.AddInstance(request.ServiceName, request.Host, request.Port, request.Weight)
	//获取调整后的结果
	serviceInfo, ok := r.Mgr.TryGetServiceInfo(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("没有成功创建ServiceInfo, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}
	//写入redis
	err = r.SendServiceToRedis(registerCtx, serviceInfo)
	if err != nil {
		return nil, err
	}

	return &miniRouterProto.RegisterReply{InstanceID: instanceInfo.InstanceID, Existed: false}, nil
}

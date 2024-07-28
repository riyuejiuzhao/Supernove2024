package main

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
)

type DeregisterContext struct {
	ctx     context.Context
	hash    string
	request *miniRouterProto.DeregisterRequest
}

func (d *DeregisterContext) GetContext() context.Context {
	return d.ctx
}

func (d *DeregisterContext) GetServiceName() string {
	return d.request.ServiceName
}

func (d *DeregisterContext) GetServiceHash() string {
	return d.hash
}

func (r *RegisterSvr) Deregister(ctx context.Context,
	request *miniRouterProto.DeregisterRequest,
) (*miniRouterProto.DeregisterReply, error) {
	deregisterCtx := &DeregisterContext{
		request: request,
		hash:    ServiceHash(request.ServiceName),
		ctx:     ctx,
	}

	mutex := r.rMutex.NewMutex(ServiceInfoLockName(request.ServiceName))
	err := mutex.Lock()
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	defer util.TryUnlock(mutex)

	//更新缓存
	err = r.flushBuffer(deregisterCtx)
	if err != nil {
		return nil, err
	}

	//删除本地缓存
	r.mgr.RemoveInstance(request.ServiceName, request.Host, request.Port)
	serviceInfo, ok := r.mgr.TryGetServiceInfo(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("ServiceInfo丢失, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}

	//写入redis
	err = r.sendServiceToRedis(deregisterCtx, serviceInfo)
	if err != nil {
		return nil, err
	}

	return &miniRouterProto.DeregisterReply{}, nil
}

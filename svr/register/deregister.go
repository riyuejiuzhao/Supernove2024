package main

import (
	"Supernove2024/miniRouterProto"
	"context"
)

type DeregisterContext struct {
}

/*
GetContext() context.Context
GetServiceName() string
GetServiceHash() string

GetServiceInfo() *miniRouterProto.ServiceInfo
SetServiceInfo(*miniRouterProto.ServiceInfo)
*/

func (r *RegisterSvr) checkInstanceExist(ctx *DeregisterContext) {
	r.checkBufVersion(ctx)
}

func (r *RegisterSvr) Deregister(context.Context,
	*miniRouterProto.DeregisterRequest,
) (*miniRouterProto.DeregisterReply, error) {

}

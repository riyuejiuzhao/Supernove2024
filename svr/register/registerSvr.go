package main

import (
	"Supernove2024/miniRouterProto"
	"context"
	"errors"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceKey, serviceName)
}

type RegisterSvrContext interface {
	GetContext() context.Context
	GetServiceName() string
	GetServiceHash() string
}

// 如果Revision和Redis不同就刷新
func (r *RegisterSvr) flushBuffer(ctx RegisterSvrContext) error {
	redisRevision, err := r.rdb.HGet(ctx.GetContext(),
		ctx.GetServiceHash(), ServiceRevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	serviceInfo, ok := r.mgr.TryGetServiceInfo(ctx.GetServiceName())
	if !ok || serviceInfo.Revision != redisRevision {
		//需要更新本地缓存
		infoBytes, err := r.rdb.HGet(ctx.GetContext(), ctx.GetServiceHash(), ServiceInfoFiled).Bytes()
		if err != nil {
			return err
		}
		serviceInfo = &miniRouterProto.ServiceInfo{}
		err = proto.Unmarshal(infoBytes, serviceInfo)
		if err != nil {
			return err
		}
		r.mgr.FlushService(serviceInfo)
	}
	return nil
}

type RegisterSvr struct {
	mgr    RegisterDataManager
	rdb    *redis.Client
	rMutex *redsync.Redsync
	miniRouterProto.UnimplementedRegisterServiceServer
}

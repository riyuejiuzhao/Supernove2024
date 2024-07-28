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

// ServiceInfoLockName 对一个资源的分布式锁给一个名字
func ServiceInfoLockName(serviceName string) string {
	return fmt.Sprintf("Register%s", serviceName)
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

// 将数据发送到Redis
func (r *RegisterSvr) sendServiceToRedis(ctx RegisterSvrContext, info *miniRouterProto.ServiceInfo) error {
	bytes, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	txPipeline := r.rdb.TxPipeline()
	txPipeline.HSet(ctx.GetContext(), ctx.GetServiceHash(), ServiceRevisionFiled,
		info.Revision)
	txPipeline.HSet(ctx.GetContext(), ctx.GetServiceHash(), ServiceInfoFiled, bytes)
	_, err = txPipeline.Exec(ctx.GetContext())
	if err != nil {
		return err
	}
	return nil
}

type RegisterSvr struct {
	mgr    RegisterDataManager
	rdb    *redis.Client
	rMutex *redsync.Redsync
	miniRouterProto.UnimplementedRegisterServiceServer
}

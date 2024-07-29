package svrutil

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"google.golang.org/protobuf/proto"
	"log"
)

const (
	ServiceKey           = "Service"
	ServiceInfoFiled     = "Info"
	ServiceRevisionFiled = "Revision"
)

type BaseServer struct {
	Mgr      ServiceBuffer
	Rdb      *redis.Client
	RedMutex *redsync.Redsync
}

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceKey, serviceName)
}

// ServiceInfoLockName 对一个资源的分布式锁给一个名字
func ServiceInfoLockName(serviceName string) string {
	return fmt.Sprintf("Register%s", serviceName)
}

func NewBaseSvr(redisAddress string, redisPassword string, redisDB int) *BaseServer {
	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalln(err)
	} else {
		util.Info("redis PONG")
	}

	return &BaseServer{
		Mgr:      NewServiceBuffer(),
		Rdb:      rdb,
		RedMutex: rs,
	}
}

type SvrContext interface {
	GetServiceName() string
	GetServiceHash() string
}

// FlushBuffer 如果Revision和Redis不同就刷新
func (r *BaseServer) FlushBuffer(ctx SvrContext) error {
	redisRevision, err := r.Rdb.HGet(ctx.GetServiceHash(), ServiceRevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	serviceInfo, ok := r.Mgr.TryGetServiceInfo(ctx.GetServiceName())
	if !ok || serviceInfo.Revision != redisRevision {
		originRevision := serviceInfo.Revision
		//需要更新本地缓存
		infoBytes, err := r.Rdb.HGet(ctx.GetServiceHash(), ServiceInfoFiled).Bytes()
		if err != nil {
			return err
		}
		serviceInfo = &miniRouterProto.ServiceInfo{}
		err = proto.Unmarshal(infoBytes, serviceInfo)
		if err != nil {
			return err
		}
		r.Mgr.FlushService(serviceInfo)
		util.Info("刷新%s缓存,%v->%v", serviceInfo.ServiceName, originRevision, redisRevision)
	}
	return nil
}

// SendServiceToRedis 将数据发送到Redis
func (r *BaseServer) SendServiceToRedis(ctx SvrContext, info *miniRouterProto.ServiceInfo) error {
	bytes, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(ctx.GetServiceHash(), ServiceRevisionFiled, info.Revision)
	txPipeline.HSet(ctx.GetServiceHash(), ServiceInfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return err
	}
	util.Info("更新redis: %s, %v", info.ServiceName, info.Revision)
	return nil
}

func (r *BaseServer) LockRedisService(serviceName string) (*redsync.Mutex, error) {
	mutex := r.RedMutex.NewMutex(ServiceInfoLockName(serviceName))
	err := mutex.Lock()
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	return mutex, nil
}

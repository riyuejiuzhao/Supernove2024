package register

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
)

const (
	ServiceKey           = "Service"
	ServiceInfoFiled     = "Info"
	ServiceRevisionFiled = "Revision"
)

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceKey, serviceName)
}

type SvrContext interface {
	GetContext() context.Context
	GetServiceName() string
	GetServiceHash() string
}

// ServiceInfoLockName 对一个资源的分布式锁给一个名字
func ServiceInfoLockName(serviceName string) string {
	return fmt.Sprintf("Register%s", serviceName)
}

// 如果Revision和Redis不同就刷新
func (r *Server) flushBuffer(ctx SvrContext) error {
	redisRevision, err := r.rdb.HGet(ctx.GetServiceHash(), ServiceRevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	serviceInfo, ok := r.mgr.TryGetServiceInfo(ctx.GetServiceName())
	if !ok || serviceInfo.Revision != redisRevision {
		//需要更新本地缓存
		infoBytes, err := r.rdb.HGet(ctx.GetServiceHash(), ServiceInfoFiled).Bytes()
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
func (r *Server) sendServiceToRedis(ctx SvrContext, info *miniRouterProto.ServiceInfo) error {
	bytes, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	txPipeline := r.rdb.TxPipeline()
	txPipeline.HSet(ctx.GetServiceHash(), ServiceRevisionFiled, info.Revision)
	txPipeline.HSet(ctx.GetServiceHash(), ServiceInfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return err
	}
	return nil
}

type Server struct {
	mgr    ServiceBuffer
	rdb    *redis.Client
	rMutex *redsync.Redsync
	miniRouterProto.UnimplementedRegisterServiceServer
}

func NewSvr(rdb *redis.Client, rs *redsync.Redsync) *Server {
	return &Server{
		mgr:    NewDefaultServiceBuffer(),
		rdb:    rdb,
		rMutex: rs,
	}
}

func SetupServer(address string, redisAddress string, redisPassword string, redisDB int) {
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

	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	miniRouterProto.RegisterRegisterServiceServer(grpcServer, NewSvr(rdb, rs))
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

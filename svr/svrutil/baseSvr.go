package svrutil

import (
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"sync/atomic"
)

const (
	MethodTag  = "Method"
	ServiceTag = "Service"

	AllServiceChannel = "AllService"
	ServiceChannel    = "Chan.Service"

	//健康信息 redis key
	HealthHashKey            = "Hash.Health"
	HealthTtlFiled           = "TTL"
	HealthLastHeartBeatField = "LastHeartBeat"

	//服务信息 redis key
	ServiceHashKey = "Hash.Service"

	//路由信息
	RouterHashKey = "Hash.Router"

	InfoField     = "Info"
	RevisionFiled = "Revision"
)

// BaseServer 基本服务器，保持和Redis链接的能力
type BaseServer struct {
	Rdb        *redis.Client
	RedMutex   *redsync.Redsync
	GrpcServer *grpc.Server
	PromReg    *prometheus.Registry

	SendRedisCount  *prometheus.CounterVec
	RecvRedisCount  *prometheus.CounterVec
	RpcRequestCount *prometheus.CounterVec
	RpcReplyCount   *prometheus.CounterVec
}

func ServiceChan(service string) string {
	return fmt.Sprintf("%s.%s", ServiceChannel, service)
}

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceHashKey, serviceName)
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func HealthHash(serviceName string, instanceID string) string {
	return fmt.Sprintf("%s.%s.%s", HealthHashKey, serviceName, instanceID)
}

func RouterIsKvField(field string) bool {
	startIndex := len(InfoField) + 1
	return field[startIndex:startIndex+2] == "KV"
}

func RouterIsDstField(field string) bool {
	startIndex := len(InfoField) + 1
	return field[startIndex:startIndex+3] == "Dst"
}

func RouterKVInfoField(key string) string {
	return fmt.Sprintf("%s.KV.%s", InfoField, key)
}

func RouterDstInfoField(key string) string {
	return fmt.Sprintf("%s.Dst.%s", InfoField, key)
}

func RouterHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", RouterHashKey, serviceName)
}

func (s *BaseServer) Setup(ctx context.Context, address string, metricsAddress string) {
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		<-ctx.Done()
		s.GrpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()
	go func() {
		m := http.NewServeMux()
		m.Handle("/metrics", promhttp.HandlerFor(
			s.PromReg, promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}))
		if err := http.ListenAndServe(metricsAddress, m); err != nil {
			log.Fatalln(err)
		}
	}()
	if err = s.GrpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

func (s *BaseServer) MethodMetricsUpload(
	label prometheus.Labels,
	request *int64,
	reply *int64,
	sendRedis *int64,
	recvRedis *int64) {
	var temp int64

	if reply != nil {
		temp = atomic.LoadInt64(reply)
		atomic.AddInt64(reply, -temp)
		s.RpcReplyCount.With(label).Add(float64(temp))
	}

	if request != nil {
		temp = atomic.LoadInt64(request)
		atomic.AddInt64(request, -temp)
		s.RpcRequestCount.With(label).Add(float64(temp))
	}

	if recvRedis != nil {
		temp = atomic.LoadInt64(recvRedis)
		atomic.AddInt64(recvRedis, -temp)
		s.RecvRedisCount.With(label).Add(float64(temp))
	}

	if sendRedis != nil {
		temp = atomic.LoadInt64(sendRedis)
		atomic.AddInt64(sendRedis, -temp)
		s.SendRedisCount.With(label).Add(float64(temp))
	}
}

func NewBaseSvr(redisAddress string, redisPassword string, redisDB int) *BaseServer {
	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	//redis 监控
	redisSend := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "SendRedisCount",
	}, []string{MethodTag, ServiceTag})
	redisRecv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RecvRedisCount",
	}, []string{MethodTag, ServiceTag})

	//创建grpc监控
	rpcSendCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RpcRequestCount",
	}, []string{MethodTag, ServiceTag})
	rpcRecvCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RpcReplyCount",
	}, []string{MethodTag, ServiceTag})
	srvMetrics := grpcprom.NewServerMetrics()
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics, redisSend, redisRecv, rpcSendCount, rpcRecvCount)

	//创建grpc服务器
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(srvMetrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(srvMetrics.StreamServerInterceptor()))

	srvMetrics.InitializeMetrics(grpcServer)

	return &BaseServer{
		Rdb:        rdb,
		RedMutex:   rs,
		GrpcServer: grpcServer,
		PromReg:    reg,

		SendRedisCount:  redisSend,
		RecvRedisCount:  redisRecv,
		RpcRequestCount: rpcSendCount,
		RpcReplyCount:   rpcRecvCount,
	}
}

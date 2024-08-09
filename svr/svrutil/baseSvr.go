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
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"net/http"
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

	RedisSend    *prometheus.CounterVec
	RedisRecv    *prometheus.CounterVec
	RpcSendCount *prometheus.CounterVec
	RpcRecvCount *prometheus.CounterVec
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

func (s *BaseServer) MetricsUpload(
	service string,
	method string,
	request proto.Message,
	reply proto.Message,
) {
	if reply == nil {
		return
	}
	bytes, err := proto.Marshal(reply)
	if err != nil {
		return
	}
	s.RpcSendCount.With(
		prometheus.Labels{
			ServiceTag: service,
			MethodTag:  method,
		},
	).Add(float64(len(bytes)))
	bytes, err = proto.Marshal(request)
	if err != nil {
		return
	}
	s.RpcRecvCount.With(
		prometheus.Labels{
			ServiceTag: service,
			MethodTag:  method,
		},
	).Add(float64(len(bytes)))
}

func NewBaseSvr(redisAddress string, redisPassword string, redisDB int) *BaseServer {
	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	//redis 监控
	redisSend := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RedisSend",
	}, []string{MethodTag, ServiceTag})
	redisRecv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RedisRecv",
	}, []string{MethodTag, ServiceTag})

	//创建grpc监控
	rpcSendCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RpcSendCount",
	}, []string{MethodTag, ServiceTag})
	rpcRecvCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "RpcRecvCount",
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

		RedisSend:    redisSend,
		RedisRecv:    redisRecv,
		RpcSendCount: rpcSendCount,
		RpcRecvCount: rpcRecvCount,
	}
}

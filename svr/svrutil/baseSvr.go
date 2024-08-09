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
)

const (
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

func NewBaseSvr(redisAddress string, redisPassword string, redisDB int) *BaseServer {
	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	//创建监控
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

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
	}
}

package main

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"context"
	"flag"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	ServiceKey           = "Service"
	ServiceInfoFiled     = "Info"
	ServiceRevisionFiled = "Revision"
)

func main() {
	rpcAddress := flag.String("rpc", "localhost:9090", "rpc address")
	redisAddress := flag.String("redis-addr", "localhost:6379", "redis address")
	redisPassword := flag.String("redis-password", "", "redis password")
	redisDB := flag.Int("redis-db", 0, "DB number")

	util.Info("rpc address: %s", *rpcAddress)
	util.Info("redis address: %s", *redisAddress)
	util.Info("redis password: %s", *redisPassword)
	util.Info("redis db: %v", *redisDB)

	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: *redisAddress, Password: *redisPassword, DB: *redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("没有找到redis服务")
	} else {
		util.Info("redis PONG")
	}

	//创建rpc服务器
	lis, err := net.Listen("tcp", *rpcAddress)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	miniRouterProto.RegisterRegisterServiceServer(grpcServer,
		&RegisterSvr{
			mgr:    &DefaultRegisterDataManager{},
			rdb:    rdb,
			rMutex: rs,
		})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

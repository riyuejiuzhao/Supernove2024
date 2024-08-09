package main

import (
	"Supernove2024/svr/register"
	"Supernove2024/util"
	"context"
	"flag"
)

func main() {
	rpcAddress := flag.String("rpc", "localhost:9090", "rpc address")
	metricsAddress := flag.String("metrics", "localhost:6060", "rpc address")
	redisAddress := flag.String("redis-addr", "localhost:6379", "redis address")
	redisPassword := flag.String("redis-password", "", "redis password")
	redisDB := flag.Int("redis-db", 0, "DB number")
	flag.Parse()

	util.Info("rpc address: %s", *rpcAddress)
	util.Info("redis address: %s", *redisAddress)
	util.Info("redis password: %s", *redisPassword)
	util.Info("redis db: %v", *redisDB)

	register.SetupServer(
		context.Background(),
		*rpcAddress,
		*metricsAddress,
		*redisAddress,
		*redisPassword,
		*redisDB,
	)
}

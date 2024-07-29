package health

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type Server struct {
	*svrutil.BaseServer
	miniRouterProto.UnimplementedHealthServiceServer
}

// GetHealthInfo 同步健康数据
func (s *Server) GetHealthInfo(_ context.Context,
	req *miniRouterProto.GetHealthInfoRequest,
) (*miniRouterProto.GetHealthInfoReply, error) {
	healthSetKey := svrutil.ServiceSetKey(req.ServiceName)
	members, err := s.Rdb.SMembers(healthSetKey).Result()
	if err != nil {
		return nil, err
	}
	ttlSlice := make([]*redis.StringCmd, len(members))

	heatBeatSlice := make([]*redis.StringCmd, len(members))
	txPipeline := s.Rdb.TxPipeline()
	for i, instanceID := range members {
		hash := svrutil.HealthHash(req.ServiceName, instanceID)
		ttlSlice[i] = txPipeline.HGet(hash, svrutil.HealthTtlFiled)
		heatBeatSlice[i] = txPipeline.HGet(hash, svrutil.HealthLastHeartBeatField)
	}
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}

	healthInfo := make([]*miniRouterProto.HealthInfo, 0, len(members))
	for i := 0; i < len(members); i++ {
		healthBeat, err := heatBeatSlice[i].Int64()
		if err != nil {
			util.Error("解析LastHeartBeat出错，InstanceID: %v, err: %v", members[i], err)
			continue
		}
		ttl, err := ttlSlice[i].Int64()
		if err != nil {
			util.Error("解析TTL出错，InstanceID: %v, err: %v", members[i], err)
			continue
		}
		healthInfo = append(healthInfo, &miniRouterProto.HealthInfo{TTL: ttl, LastHeartBeat: healthBeat})
	}
	return &miniRouterProto.GetHealthInfoReply{HealthInfos: healthInfo}, err
}

func (s *Server) HeartBeat(_ context.Context, req *miniRouterProto.HeartBeatRequest) (*miniRouterProto.HeartBeatReply, error) {
	lastHeartBeat := time.Now().Unix()
	key := svrutil.HealthHash(req.ServiceName, req.InstanceID)
	s.Rdb.HSet(key, "LastHeartBeat", lastHeartBeat)
	return &miniRouterProto.HeartBeatReply{}, nil
}

func SetupServer(address string, redisAddress string, redisPassword string, redisDB int) {
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer()
	miniRouterProto.RegisterHealthServiceServer(grpcServer,
		&Server{BaseServer: svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

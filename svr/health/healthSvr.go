package health

import (
	"Supernove2024/pb"
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
	*svrutil.BufferServer
	pb.UnimplementedHealthServiceServer
}

// GetHealthInfo 同步健康数据
func (s *Server) GetHealthInfo(_ context.Context,
	req *pb.GetHealthInfoRequest,
) (*pb.GetHealthInfoReply, error) {
	//批量获取ID
	hash := svrutil.ServiceHash(req.ServiceName)
	err := s.FlushBufferLocked(hash, req.ServiceName)
	if err != nil {
		return nil, err
	}

	serviceInfo, ok := s.Mgr.TryGetServiceInfo(req.ServiceName)
	if !ok {
		return nil, err
	}

	instanceInfos := make([]*pb.InstanceHealthInfo, 0, len(serviceInfo.Instances))
	nowTtlSlice := make([]*redis.StringCmd, 0, len(serviceInfo.Instances))
	nowHeatBeatSlice := make([]*redis.StringCmd, 0, len(serviceInfo.Instances))
	txPipeline := s.Rdb.TxPipeline()
	for _, instance := range serviceInfo.Instances {
		hash := svrutil.HealthHash(req.ServiceName, instance.InstanceID)
		nowTtlSlice = append(nowTtlSlice, txPipeline.HGet(hash, svrutil.HealthTtlFiled))
		nowHeatBeatSlice = append(nowHeatBeatSlice, txPipeline.HGet(hash, svrutil.HealthLastHeartBeatField))
	}
	_, err = txPipeline.Exec()
	if err != nil {
		util.Error("%v", err)
		return nil, err
	}
	for j := 0; j < len(serviceInfo.Instances); j++ {
		ttl, err := nowTtlSlice[j].Int64()
		if err != nil {
			util.Error("get ttl err: %v", err)
			continue
		}
		heartBeat, err := nowHeatBeatSlice[j].Int64()
		if err != nil {
			util.Error("get lastHeartBeat err: %v", err)
			continue
		}
		instanceInfos = append(instanceInfos, &pb.InstanceHealthInfo{TTL: ttl, LastHeartBeat: heartBeat})
	}
	return &pb.GetHealthInfoReply{HealthInfo: &pb.ServiceHealthInfo{ServiceName: req.ServiceName, InstanceHealthInfo: instanceInfos}}, err
}

func (s *Server) HeartBeat(_ context.Context, req *pb.HeartBeatRequest) (*pb.HeartBeatReply, error) {
	lastHeartBeat := time.Now().Unix()
	key := svrutil.HealthHash(req.ServiceName, req.InstanceID)
	s.Rdb.HSet(key, svrutil.HealthLastHeartBeatField, lastHeartBeat)
	return &pb.HeartBeatReply{}, nil
}

func SetupServer(address string, redisAddress string, redisPassword string, redisDB int) {
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHealthServiceServer(grpcServer,
		&Server{BufferServer: svrutil.NewBufferSvr(redisAddress, redisPassword, redisDB)})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

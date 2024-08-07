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
	"strings"
	"time"
)

type Server struct {
	*svrutil.BufferServer
	pb.UnimplementedHealthServiceServer
}

// GetHealthInfo 同步健康数据
func (s *Server) GetHealthInfo(_ context.Context,
	req *pb.GetHealthInfoRequest,
) (reply *pb.GetHealthInfoReply, err error) {
	match := svrutil.HealthHash(req.ServiceName, "*")
	prefix := svrutil.HealthHash(req.ServiceName, "")
	reply = nil

	cursor := uint64(0)
	allResult := make([][]string, 0)
	result, cursor, err := s.Rdb.Scan(cursor, match, 10000).Result()
	allResult = append(allResult, result)
	count := len(result)
	if err != nil {
		return
	}
	for cursor != 0 {
		result, cursor, err = s.Rdb.Scan(cursor, match, 10000).Result()
		if err != nil {
			return
		}
		allResult = append(allResult, result)
		count += len(result)
	}

	allInstances := make([]string, 0, count)
	for _, r := range allResult {
		for _, i := range r {
			allInstances = append(allInstances, strings.TrimPrefix(i, prefix))
		}
	}
	instanceInfos := make([]*pb.InstanceHealthInfo, 0, count)
	nowTtlSlice := make([]*redis.StringCmd, 0, count)
	nowHeatBeatSlice := make([]*redis.StringCmd, 0, count)
	pipeline := s.Rdb.Pipeline()
	for _, instance := range allInstances {
		hash := svrutil.HealthHash(req.ServiceName, instance)
		nowTtlSlice = append(nowTtlSlice, pipeline.HGet(hash, svrutil.HealthTtlFiled))
		nowHeatBeatSlice = append(nowHeatBeatSlice, pipeline.HGet(hash, svrutil.HealthLastHeartBeatField))
	}
	_, err = pipeline.Exec()
	if err != nil {
		util.Error("%v", err)
		return nil, err
	}
	for j, ins := range allInstances {
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
		instanceInfos = append(instanceInfos, &pb.InstanceHealthInfo{
			InstanceID: ins, TTL: ttl, LastHeartBeat: heartBeat})
	}
	return &pb.GetHealthInfoReply{HealthInfo: &pb.ServiceHealthInfo{ServiceName: req.ServiceName, InstanceHealthInfo: instanceInfos}}, err
}

func (s *Server) HeartBeat(_ context.Context, req *pb.HeartBeatRequest) (*pb.HeartBeatReply, error) {
	lastHeartBeat := time.Now().Unix()
	key := svrutil.HealthHash(req.ServiceName, req.InstanceID)
	s.Rdb.HSet(key, svrutil.HealthLastHeartBeatField, lastHeartBeat)
	return &pb.HeartBeatReply{}, nil
}

func SetupServer(ctx context.Context, address string, redisAddress string, redisPassword string, redisDB int) {
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHealthServiceServer(grpcServer,
		&Server{BufferServer: svrutil.NewBufferSvr(redisAddress, redisPassword, redisDB)})
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

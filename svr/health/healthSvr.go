package health

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"github.com/go-redis/redis"
	"strings"
	"time"
)

type Server struct {
	*svrutil.BaseServer
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

func (s *Server) HeartBeat(_ context.Context, request *pb.HeartBeatRequest) (reply *pb.HeartBeatReply, err error) {
	lastHeartBeat := time.Now().Unix()
	key := svrutil.HealthHash(request.ServiceName, request.InstanceID)
	s.Rdb.HSet(key, svrutil.HealthLastHeartBeatField, lastHeartBeat)
	reply, err = &pb.HeartBeatReply{}, nil
	return
}

func SetupServer(
	ctx context.Context,
	address string,
	metricsAddress string,
	redisAddress string,
	redisPassword string,
	redisDB int,
) {
	baseSvr := svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)
	pb.RegisterHealthServiceServer(baseSvr.GrpcServer,
		&Server{BaseServer: baseSvr})
	baseSvr.Setup(ctx, address, metricsAddress)
}

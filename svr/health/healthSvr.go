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
	*svrutil.BaseServer
	pb.UnimplementedHealthServiceServer
}

// GetHealthInfo 同步健康数据
func (s *Server) GetHealthInfo(_ context.Context,
	req *pb.GetHealthInfoRequest,
) (*pb.GetHealthInfoReply, error) {
	lockedServices := make([]string, 0, len(req.ServiceNames))
	//给所有需要的上锁
	for _, name := range req.ServiceNames {
		lockName := svrutil.ServiceHealthInfoLockName(name)
		mutex, err := s.LockRedisService(lockName)
		if err != nil {
			util.Error("get lock %s failed %v", lockName, err)
			continue
		}
		lockedServices = append(lockedServices, name)
		defer svrutil.TryUnlock(mutex)
	}

	//批量获取ID
	txPipeline := s.Rdb.TxPipeline()
	smembers := make([]*redis.StringSliceCmd, 0, len(lockedServices))
	for _, name := range lockedServices {
		healthSetKey := svrutil.ServiceSetKey(name)
		smembers = append(smembers, txPipeline.SMembers(healthSetKey))
	}
	_, err := txPipeline.Exec()
	if err != nil {
		util.Error("%v", err)
		return nil, err
	}

	ttlSlice := make([][]*redis.StringCmd, 0, len(smembers))
	heatBeatSlice := make([][]*redis.StringCmd, 0, len(smembers))
	healthInfos := make([]*pb.ServiceHealthInfo, 0, len(smembers))
	txPipeline = s.Rdb.TxPipeline()
	for idx, mem := range smembers {
		ids, err := mem.Result()
		nowServiceName := lockedServices[idx]
		if err != nil {
			util.Error("get instances for %s failed", nowServiceName)
			continue
		}
		instanceInfos := make([]*pb.InstanceHealthInfo, 0, len(ids))
		nowTtlSlice := make([]*redis.StringCmd, 0, len(ids))
		nowHeatBeatSlice := make([]*redis.StringCmd, 0, len(ids))
		for _, id := range ids {
			hash := svrutil.HealthHash(lockedServices[idx], id)
			nowTtlSlice = append(nowTtlSlice, txPipeline.HGet(hash, svrutil.HealthTtlFiled))
			nowHeatBeatSlice = append(nowHeatBeatSlice, txPipeline.HGet(hash, svrutil.HealthLastHeartBeatField))
			instanceInfos = append(instanceInfos, &pb.InstanceHealthInfo{InstanceID: id})
		}
		healthInfos = append(healthInfos, &pb.ServiceHealthInfo{
			ServiceName:        nowServiceName,
			InstanceHealthInfo: instanceInfos,
		})
		ttlSlice = append(ttlSlice, nowTtlSlice)
		heatBeatSlice = append(heatBeatSlice, nowHeatBeatSlice)
	}
	_, err = txPipeline.Exec()
	if err != nil {
		util.Error("%v", err)
		return nil, err
	}

	for i, healthInfo := range healthInfos {
		nowInstances := make([]*pb.InstanceHealthInfo, 0, len(healthInfo.InstanceHealthInfo))
		for j, instanceHealthInfo := range healthInfo.InstanceHealthInfo {
			ttl, err := ttlSlice[i][j].Int64()
			if err != nil {
				util.Error("get ttl err: %v", err)
				continue
			}
			heartBeat, err := heatBeatSlice[i][j].Int64()
			if err != nil {
				util.Error("get lastHeartBeat err: %v", err)
				continue
			}
			instanceHealthInfo.TTL = ttl
			instanceHealthInfo.LastHeartBeat = heartBeat
			nowInstances = append(nowInstances, instanceHealthInfo)
		}
		healthInfo.InstanceHealthInfo = nowInstances
	}

	return &pb.GetHealthInfoReply{HealthInfos: healthInfos}, err
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
		&Server{BaseServer: svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

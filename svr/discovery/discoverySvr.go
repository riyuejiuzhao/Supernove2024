package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
)

type Server struct {
	*svrutil.BufferServer
	svrutil.ServiceBuffer
	pb.UnimplementedDiscoveryServiceServer
}

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (reply *pb.GetInstancesReply, err error) {
	hash := svrutil.ServiceHash(request.ServiceName)
	reply = nil
	//汇报报文
	defer func() { util.Info("GetInstancesReply:%v", reply) }()

	allResult := make([][]string, 0, 10)
	count := 0
	serviceInfo, ok := s.ServiceBuffer.GetServiceInfo(request.ServiceName)
	if ok && serviceInfo.Revision == request.Revision {
		reply = &pb.GetInstancesReply{
			Service: &pb.ServiceInfo{
				ServiceName: request.ServiceName,
				Revision:    request.Revision,
				Instances:   make([]*pb.InstanceInfo, 0),
			},
		}
		return
	}
	var revision int64
	if !ok {
		revision = 0
	} else {
		revision = serviceInfo.Revision
	}
	needUpdate, err := func() (bool, error) {
		mutex, err := s.LockRedis(svrutil.ServiceInfoLockName(request.ServiceName))
		if err != nil {
			util.Error("lock redis failed err:%v", err)
			return false, err
		}
		defer svrutil.TryUnlock(mutex)

		latestRevision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
		if latestRevision == revision {
			return false, nil
		}

		cursor := uint64(0)
		result, cursor, err := s.Rdb.HScan(hash, cursor, svrutil.InfoFieldMatch, 10000).Result()
		if err != nil {
			return false, err
		}
		allResult = append(allResult, result)
		count = len(result)
		for cursor != 0 {
			result, cursor, err = s.Rdb.HScan(hash, cursor, svrutil.InfoFieldMatch, 10000).Result()
			if err != nil {
				return false, err
			}
			count += len(result)
			allResult = append(allResult, result)
		}
		return true, nil
	}()

	if err != nil {
		return
	}

	if needUpdate {
		serviceInfo = &pb.ServiceInfo{
			ServiceName: request.ServiceName,
			Revision:    revision,
			Instances:   make([]*pb.InstanceInfo, 0, count),
		}

		for i := 0; i < len(allResult); i += 1 {
			nowResult := allResult[i]
			for j := 0; j < len(nowResult); j += 2 {
				bytes := []byte(nowResult[j+1])
				instance := &pb.InstanceInfo{}
				err = proto.Unmarshal(bytes, instance)
				if err != nil {
					return
				}
				serviceInfo.Instances = append(serviceInfo.Instances, instance)
			}
		}
		s.ServiceBuffer.FlushService(serviceInfo)
	}
	reply = &pb.GetInstancesReply{Service: serviceInfo}
	return
}

func (s *Server) GetRouters(_ context.Context, request *pb.GetRoutersRequest) (reply *pb.GetRoutersReply, err error) {
	defer func() {
		if reply == nil {
			return
		}
		util.Info("GetRouters: %v", reply)
	}()

	hash := svrutil.RouterHash(request.ServiceName)

	err = s.FlushRouterBufferLocked(hash, request.ServiceName)
	if err != nil {
		return
	}

	serviceInfo, ok := s.RouterBuffer.GetServiceRouter(request.ServiceName)
	if !ok {
		reply = &pb.GetRoutersReply{
			Router: &pb.ServiceRouterInfo{
				ServiceName:   request.ServiceName,
				Revision:      int64(0),
				TargetRouters: make([]*pb.TargetRouterInfo, 0),
				KVRouters:     make([]*pb.KVRouterInfo, 0),
			},
		}
		return
	}

	if serviceInfo.Revision == request.Revision {
		reply = &pb.GetRoutersReply{
			Router: &pb.ServiceRouterInfo{
				ServiceName:   request.ServiceName,
				Revision:      request.Revision,
				TargetRouters: make([]*pb.TargetRouterInfo, 0),
				KVRouters:     make([]*pb.KVRouterInfo, 0),
			},
		}
		return
	}

	reply = &pb.GetRoutersReply{
		Router: serviceInfo,
	}
	return
}

func SetupServer(ctx context.Context, address string, redisAddress string, redisPassword string, redisDB int) {
	baseSvr := svrutil.NewBufferSvr(redisAddress, redisPassword, redisDB)

	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(grpcServer,
		&Server{BufferServer: baseSvr, ServiceBuffer: svrutil.NewServiceBuffer()})

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

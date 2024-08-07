package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	*svrutil.BufferServer
	pb.UnimplementedDiscoveryServiceServer
}

func (s *Server) GetServices(_ context.Context, _ *pb.GetServicesRequest) (*pb.GetServicesReply, error) {
	nowKeys, c, err := s.Rdb.Scan(0, svrutil.ServiceHash("*"), 1000).Result()
	if err != nil {
		return nil, err
	}
	keys := util.Map(nowKeys, svrutil.ServiceHashToServiceName)
	for c != 0 {
		nowKeys, c, err = s.Rdb.Scan(0, svrutil.ServiceHash("*"), 1000).Result()
		keys = append(keys, util.Map(nowKeys, svrutil.ServiceHashToServiceName)...)
	}
	return &pb.GetServicesReply{ServiceName: keys}, nil
}

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (reply *pb.GetInstancesReply, err error) {
	hash := svrutil.ServiceHash(request.ServiceName)

	reply = nil
	err = s.FlushServiceBufferLocked(hash, request.ServiceName)
	if err != nil {
		return
	}
	//汇报报文
	defer func() { util.Info("GetInstancesReply:%v", reply) }()

	serviceInfo, ok := s.InstanceBuffer.GetServiceInfo(request.ServiceName)
	if !ok {
		reply = &pb.GetInstancesReply{
			Service: &pb.ServiceInfo{
				ServiceName: request.ServiceName,
				Revision:    int64(0),
				Instances:   make([]*pb.InstanceInfo, 0),
			},
		}
		err = nil
		return
	}

	if serviceInfo.Revision == request.Revision {
		err = nil
		reply = &pb.GetInstancesReply{
			Service: &pb.ServiceInfo{
				ServiceName: request.ServiceName,
				Revision:    request.Revision,
				Instances:   make([]*pb.InstanceInfo, 0),
			},
		}
		return
	}

	err = nil
	reply = &pb.GetInstancesReply{
		Service: serviceInfo,
	}
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
		&Server{BufferServer: baseSvr})

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

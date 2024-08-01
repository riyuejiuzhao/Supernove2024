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

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesReply, error) {
	hash := svrutil.ServiceHash(request.ServiceName)

	err := s.FlushServiceBufferLocked(hash, request.ServiceName)
	if err != nil {
		return nil, err
	}

	serviceInfo, ok := s.InstanceBuffer.GetServiceInfo(request.ServiceName)
	if !ok {
		return &pb.GetInstancesReply{
			Instances: make([]*pb.InstanceInfo, 0),
			Revision:  int64(0),
		}, nil
	}

	if serviceInfo.Revision == request.Revision {
		return &pb.GetInstancesReply{
			Instances: make([]*pb.InstanceInfo, 0),
			Revision:  serviceInfo.Revision,
		}, nil
	}

	return &pb.GetInstancesReply{
		Instances: serviceInfo.Instances,
		Revision:  serviceInfo.Revision}, nil

}

func SetupServer(address string, redisAddress string, redisPassword string, redisDB int) {
	baseSvr := svrutil.NewBufferSvr(redisAddress, redisPassword, redisDB)

	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(grpcServer,
		&Server{BufferServer: baseSvr})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

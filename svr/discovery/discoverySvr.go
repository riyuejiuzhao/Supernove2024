package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	*svrutil.BufferServer
	pb.UnimplementedDiscoveryServiceServer
}

type DiscoveryContext struct {
	hash    string
	request *pb.GetInstancesRequest
}

func (c *DiscoveryContext) GetServiceName() string {
	return c.request.ServiceName
}

func (c *DiscoveryContext) GetServiceHash() string {
	return c.hash
}

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesReply, error) {
	disCtx := &DiscoveryContext{
		hash:    svrutil.ServiceHash(request.ServiceName),
		request: request,
	}

	mutex, err := s.LockRedisService(request.ServiceName)
	if err != nil {
		return nil, err
	}
	defer svrutil.TryUnlock(mutex)

	err = s.FlushBuffer(disCtx)
	if err != nil {
		return nil, err
	}

	serviceInfo, ok := s.Mgr.TryGetServiceInfo(request.ServiceName)
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

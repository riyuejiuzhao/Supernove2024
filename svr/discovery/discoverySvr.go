package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"strconv"
)

type Server struct {
	*svrutil.BaseServer
	svrutil.RouterBuffer
	svrutil.ServiceBuffer
	pb.UnimplementedDiscoveryServiceServer
}

func (s *Server) GetInstances(request *pb.GetInstancesRequest, stream pb.DiscoveryService_GetInstancesServer) (err error) {
	hash := svrutil.ServiceHash(request.ServiceName)
	serviceInfo, ok := s.ServiceBuffer.GetServiceInfo(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		err = stream.Send(&pb.GetInstancesReply{
			ServiceName: request.ServiceName,
			Instances:   make([]*pb.InstanceInfo, 0),
		})
		return
	} else if err != nil {
		return
	}
	var result map[string]string
	if !ok || revision != serviceInfo.Revision {
		result, err = s.Rdb.HGetAll(hash).Result()
		if err != nil {
			return
		}
		serviceInfo = &pb.ServiceInfo{
			ServiceName: request.ServiceName,
			Revision:    0,
			Instances:   make([]*pb.InstanceInfo, 0),
		}

		for k, v := range result {
			if k == svrutil.RevisionFiled {
				revision, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return
				}
				serviceInfo.Revision = revision
			} else {
				bytes := []byte(v)
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

	if serviceInfo.Revision == request.Revision {
		err = stream.Send(&pb.GetInstancesReply{
			ServiceName: request.ServiceName,
			Revision:    request.Revision,
			Instances:   make([]*pb.InstanceInfo, 0),
		})
		return
	}

	instanceList := make([]*pb.InstanceInfo, 0, 100)
	for _, instance := range serviceInfo.Instances {
		instanceList = append(instanceList, instance)
		if len(instanceList) >= 100 {
			err = stream.Send(&pb.GetInstancesReply{
				ServiceName: serviceInfo.ServiceName,
				Revision:    serviceInfo.Revision,
				Instances:   instanceList,
			})
			if err != nil {
				return
			}
			instanceList = make([]*pb.InstanceInfo, 0, 100)
		}
	}
	if len(instanceList) != 0 {
		err = stream.Send(&pb.GetInstancesReply{
			ServiceName: serviceInfo.ServiceName,
			Revision:    serviceInfo.Revision,
			Instances:   instanceList,
		})
		if err != nil {
			return
		}
	}
	return
}

func (s *Server) GetRouters(request *pb.GetRoutersRequest, stream pb.DiscoveryService_GetRoutersServer) (err error) {
	hash := svrutil.RouterHash(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		err = stream.Send(&pb.GetRoutersReply{
			ServiceName:   request.ServiceName,
			Revision:      0,
			KvRouters:     make([]*pb.KVRouterInfo, 0),
			TargetRouters: make([]*pb.TargetRouterInfo, 0),
		})
		return
	} else if err != nil {
		return
	}
	routerInfo, ok := s.RouterBuffer.GetServiceRouter(request.ServiceName)
	var result map[string]string
	if !ok || routerInfo.Revision != revision {
		result, err = s.Rdb.HGetAll(hash).Result()
		if err != nil {
			return
		}

		routerInfo = &pb.ServiceRouterInfo{
			ServiceName:   request.ServiceName,
			Revision:      0,
			TargetRouters: make([]*pb.TargetRouterInfo, 0),
			KVRouters:     make([]*pb.KVRouterInfo, 0),
		}

		for k, v := range result {
			if k == svrutil.RevisionFiled {
				revision, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return
				}
				routerInfo.Revision = revision
			} else if svrutil.RouterIsDstField(k) {
				info := &pb.TargetRouterInfo{}
				err = proto.Unmarshal([]byte(v), info)
				if err != nil {
					return
				}
				routerInfo.TargetRouters = append(routerInfo.TargetRouters, info)
			} else if svrutil.RouterIsKvField(k) {
				info := &pb.KVRouterInfo{}
				err = proto.Unmarshal([]byte(v), info)
				if err != nil {
					return
				}
				routerInfo.KVRouters = append(routerInfo.KVRouters, info)
			} else {
				err = errors.New("错误的field")
				return
			}
		}
		s.RouterBuffer.FlushService(routerInfo)
	}

	if routerInfo.Revision == request.Revision {
		err = stream.Send(&pb.GetRoutersReply{
			ServiceName:   routerInfo.ServiceName,
			Revision:      routerInfo.Revision,
			TargetRouters: make([]*pb.TargetRouterInfo, 0),
			KvRouters:     make([]*pb.KVRouterInfo, 0),
		})
		return
	}

	kvRouters := make([]*pb.KVRouterInfo, 0, 100)
	tRouters := make([]*pb.TargetRouterInfo, 0, 100)
	for _, v := range routerInfo.KVRouters {
		kvRouters = append(kvRouters, v)
		if len(kvRouters)+len(tRouters) >= 100 {
			err = stream.Send(&pb.GetRoutersReply{
				ServiceName:   routerInfo.ServiceName,
				Revision:      routerInfo.Revision,
				TargetRouters: tRouters,
				KvRouters:     kvRouters,
			})
			if err != nil {
				return
			}
			kvRouters = make([]*pb.KVRouterInfo, 0, 100)
			tRouters = make([]*pb.TargetRouterInfo, 0, 100)
		}
	}

	for _, v := range routerInfo.TargetRouters {
		tRouters = append(tRouters, v)
		if len(kvRouters)+len(tRouters) >= 100 {
			err = stream.Send(&pb.GetRoutersReply{
				ServiceName:   routerInfo.ServiceName,
				Revision:      routerInfo.Revision,
				TargetRouters: tRouters,
				KvRouters:     kvRouters,
			})
			if err != nil {
				return
			}
			kvRouters = make([]*pb.KVRouterInfo, 0, 100)
			tRouters = make([]*pb.TargetRouterInfo, 0, 100)
		}
	}

	if len(tRouters) > 0 {
		err = stream.Send(&pb.GetRoutersReply{
			ServiceName:   routerInfo.ServiceName,
			Revision:      routerInfo.Revision,
			TargetRouters: tRouters,
			KvRouters:     kvRouters,
		})
		if err != nil {
			return
		}
	}

	return
}

func SetupServer(ctx context.Context, address string, redisAddress string, redisPassword string, redisDB int) {
	baseSvr := svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)

	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDiscoveryServiceServer(grpcServer,
		&Server{BaseServer: baseSvr,
			ServiceBuffer: svrutil.NewServiceBuffer(),
			RouterBuffer:  svrutil.NewRouterBuffer(),
		})

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

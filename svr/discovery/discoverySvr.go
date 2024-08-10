package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"context"
	"errors"
	"github.com/go-redis/redis"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync/atomic"
)

type Server struct {
	*svrutil.BaseServer
	RouterBuffer  svrutil.RouterBuffer
	ServiceBuffer svrutil.ServiceBuffer

	GetRoutersRecvRedis int64
	GetRoutersReply     int64
	GetRoutersRequest   int64

	GetInstancesRecvRedis int64
	GetInstanceReply      int64
	GetInstanceRequest    int64

	pb.UnimplementedDiscoveryServiceServer
}

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (reply *pb.GetInstancesReply, err error) {
	defer func() {
		bytes, err := proto.Marshal(request)
		if err == nil {
			atomic.AddInt64(&s.GetInstanceRequest, int64(len(bytes)))
		}
		bytes, err = proto.Marshal(reply)
		if err == nil {
			atomic.AddInt64(&s.GetInstanceReply, int64(len(bytes)))
		}
	}()

	hash := svrutil.ServiceHash(request.ServiceName)
	reply = nil
	serviceInfo, ok := s.ServiceBuffer.GetServiceInfo(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
	atomic.AddInt64(&s.GetInstancesRecvRedis, 8)

	if errors.Is(err, redis.Nil) {
		err = nil
		reply = &pb.GetInstancesReply{
			Service: &pb.ServiceInfo{
				ServiceName: request.ServiceName,
				Revision:    request.Revision,
				Instances:   make([]*pb.InstanceInfo, 0),
			},
		}
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
			atomic.AddInt64(&s.GetInstancesRecvRedis, int64(len(k)+len(v)))
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
		reply = &pb.GetInstancesReply{
			Service: &pb.ServiceInfo{
				ServiceName: request.ServiceName,
				Revision:    request.Revision,
				Instances:   make([]*pb.InstanceInfo, 0),
			},
		}
		return
	}

	if revision == serviceInfo.Revision {
		reply = &pb.GetInstancesReply{
			Service: serviceInfo,
		}
		return
	}

	reply = &pb.GetInstancesReply{Service: serviceInfo}
	return
}

func (s *Server) GetRouters(_ context.Context, request *pb.GetRoutersRequest) (reply *pb.GetRoutersReply, err error) {
	defer func() {
		bytes, err := proto.Marshal(request)
		if err == nil {
			atomic.AddInt64(&s.GetRoutersRequest, int64(len(bytes)))
		}
		bytes, err = proto.Marshal(reply)
		if err == nil {
			atomic.AddInt64(&s.GetRoutersReply, int64(len(bytes)))
		}
	}()

	hash := svrutil.RouterHash(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
	atomic.AddInt64(&s.GetRoutersRecvRedis, 8)

	if errors.Is(err, redis.Nil) {
		reply = &pb.GetRoutersReply{
			Router: &pb.ServiceRouterInfo{
				ServiceName:   request.ServiceName,
				Revision:      request.Revision,
				TargetRouters: make([]*pb.TargetRouterInfo, 0),
				KVRouters:     make([]*pb.KVRouterInfo, 0),
			},
		}
		err = nil
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
			atomic.AddInt64(&s.GetRoutersRecvRedis, int64(len(k)+len(v)))
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
	}

	if routerInfo.Revision == request.Revision {
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

	s.RouterBuffer.FlushService(routerInfo)
	reply = &pb.GetRoutersReply{Router: routerInfo}
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
	server := &Server{
		BaseServer:    baseSvr,
		ServiceBuffer: svrutil.NewServiceBuffer(),
		RouterBuffer:  svrutil.NewRouterBuffer(),
	}
	pb.RegisterDiscoveryServiceServer(baseSvr.GrpcServer, server)

	baseSvr.Setup(ctx, address, metricsAddress)
}

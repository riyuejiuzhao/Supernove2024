package discovery

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"context"
	"errors"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type Server struct {
	*svrutil.BaseServer
	svrutil.RouterBuffer
	svrutil.ServiceBuffer
	pb.UnimplementedDiscoveryServiceServer
}

func (s *Server) GetInstances(_ context.Context, request *pb.GetInstancesRequest) (reply *pb.GetInstancesReply, err error) {
	defer func() {
		const (
			Service = "Discovery"
			Method  = "GetInstances"
		)
		if reply == nil {
			return
		}
		bytes, err := proto.Marshal(reply)
		if err != nil {
			return
		}
		s.RpcSendCount.With(
			prometheus.Labels{
				svrutil.ServiceTag: Service,
				svrutil.MethodTag:  Method,
			},
		).Add(float64(len(bytes)))
		bytes, err = proto.Marshal(request)
		if err != nil {
			return
		}
		s.RpcRecvCount.With(
			prometheus.Labels{
				svrutil.ServiceTag: Service,
				svrutil.MethodTag:  Method,
			},
		).Add(float64(len(bytes)))
	}()

	hash := svrutil.ServiceHash(request.ServiceName)
	reply = nil
	serviceInfo, ok := s.ServiceBuffer.GetServiceInfo(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
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
		const (
			Service = "Discovery"
			Method  = "GetRouters"
		)
		s.MetricsUpload(Service, Method, request, reply)
	}()
	hash := svrutil.RouterHash(request.ServiceName)
	revision, err := s.Rdb.HGet(hash, svrutil.RevisionFiled).Int64()
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

// 订阅信息获取，增量获取，降低流量压力
/*
func (s *Server) Channel() {
	allServicePul := s.Rdb.Subscribe(svrutil.AllServiceChannel)
	_, err := allServicePul.Receive()
	if err != nil {
		log.Fatalf("AllServiceChannel failed %v", err)
	}

	allServiceChan := allServicePul.Channel()
	go func() {
		for msg := range allServiceChan {
			newService := msg.Payload
			s.Rdb.Subscribe("")
		}
	}()
}
*/

func SetupServer(
	ctx context.Context,
	address string,
	metricsAddress string,
	redisAddress string,
	redisPassword string,
	redisDB int,
) {
	baseSvr := svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)
	pb.RegisterDiscoveryServiceServer(baseSvr.GrpcServer,
		&Server{BaseServer: baseSvr,
			ServiceBuffer: svrutil.NewServiceBuffer(),
			RouterBuffer:  svrutil.NewRouterBuffer(),
		})

	baseSvr.Setup(ctx, address, metricsAddress)
}

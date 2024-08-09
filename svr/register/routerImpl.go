package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"google.golang.org/protobuf/proto"
	"time"
)

func (s *Server) AddKVRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	routerKvInfo := request.KvRouterInfo
	routerKvInfo.CreateTime = time.Now().Unix()
	infoField := svrutil.RouterKVInfoField(routerKvInfo.Key)
	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(routerKvInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	pipeline := s.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HSet(hash, infoField, bytes)
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s", request.ServiceName)
	return &pb.AddRouterReply{}, nil
}

func (s *Server) AddTargetRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	targetRouterInfo := request.TargetRouterInfo
	targetRouterInfo.CreateTime = time.Now().Unix()
	infoField := svrutil.RouterDstInfoField(targetRouterInfo.SrcInstanceID)
	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(targetRouterInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	pipeline := s.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HSet(hash, infoField, bytes)
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.AddRouterReply{}, nil
}

func (s *Server) AddRouter(_ context.Context, request *pb.AddRouterRequest) (reply *pb.AddRouterReply, err error) {
	defer func() {
		const (
			Service = "Register"
			Method  = "RemoveRouter"
		)
		s.MetricsUpload(Service, Method, request, reply)
	}()
	hash := svrutil.RouterHash(request.ServiceName)
	switch request.RouterType {
	case util.KVRouterType:
		reply, err = s.AddKVRouter(hash, request)
		return
	case util.TargetRouterType:
		reply, err = s.AddTargetRouter(hash, request)
		return
	default:
		reply, err = nil, errors.New("不支持的路由类型")
		return
	}
}

func (s *Server) RemoveTargetRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	//写入redis
	pipeline := s.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HDel(hash, svrutil.RouterDstInfoField(request.TargetRouterInfo.SrcInstanceID))
	_, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.RemoveRouterReply{}, nil
}

func (s *Server) RemoveKVRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	info := request.KvRouterInfo
	//写入redis
	pipeline := s.Rdb.TxPipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HDel(hash, svrutil.RouterKVInfoField(info.Key))
	_, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.RemoveRouterReply{}, nil
}

func (s *Server) RemoveRouter(_ context.Context, request *pb.RemoveRouterRequest) (reply *pb.RemoveRouterReply, err error) {
	defer func() {
		const (
			Service = "Register"
			Method  = "RemoveRouter"
		)
		s.MetricsUpload(Service, Method, request, reply)
	}()
	hash := svrutil.ServiceHash(request.ServiceName)
	switch request.RouterType {
	case util.KVRouterType:
		reply, err = s.RemoveKVRouter(hash, request)
		return
	case util.TargetRouterType:
		reply, err = s.RemoveTargetRouter(hash, request)
		return
	default:
		reply, err = nil, errors.New("不能识别的路由类型")
		return
	}
}

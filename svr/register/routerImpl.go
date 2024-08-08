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

func (r *Server) AddKVRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	routerKvInfo := request.KvRouterInfo
	routerKvInfo.CreateTime = time.Now().Unix()
	infoField := svrutil.RouterKVInfoField(routerKvInfo.Key)
	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(routerKvInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	pipeline := r.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HSet(hash, infoField, bytes)
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s", request.ServiceName)
	return &pb.AddRouterReply{}, nil
}

func (r *Server) AddTargetRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	targetRouterInfo := request.TargetRouterInfo
	targetRouterInfo.CreateTime = time.Now().Unix()
	infoField := svrutil.RouterDstInfoField(targetRouterInfo.SrcInstanceID)
	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(targetRouterInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	pipeline := r.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HSet(hash, infoField, bytes)
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.AddRouterReply{}, nil
}

func (r *Server) AddRouter(_ context.Context, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	defer util.Info(request.String())
	hash := svrutil.RouterHash(request.ServiceName)
	switch request.RouterType {
	case util.KVRouterType:
		return r.AddKVRouter(hash, request)
	case util.TargetRouterType:
		return r.AddTargetRouter(hash, request)
	default:
		return nil, errors.New("不支持的路由类型")
	}
}

func (r *Server) RemoveTargetRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	//写入redis
	pipeline := r.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HDel(hash, svrutil.RouterDstInfoField(request.TargetRouterInfo.SrcInstanceID))
	_, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.RemoveRouterReply{}, nil
}

func (r *Server) RemoveKVRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	info := request.KvRouterInfo
	//写入redis
	pipeline := r.Rdb.TxPipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HDel(hash, svrutil.RouterKVInfoField(info.Key))
	_, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return &pb.RemoveRouterReply{}, nil
}

func (r *Server) RemoveRouter(_ context.Context, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	hash := svrutil.ServiceHash(request.ServiceName)
	defer util.Info(request.String())
	//更新缓存
	switch request.RouterType {
	case util.KVRouterType:
		return r.RemoveKVRouter(hash, request)
	case util.TargetRouterType:
		return r.RemoveTargetRouter(hash, request)
	default:
		return nil, errors.New("不能识别的路由类型")
	}
}

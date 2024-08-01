package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
)

func (r *Server) AddKVRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	routerKvInfo := request.KvRouterInfo
	_, ok := r.RouterBuffer.GetKVRouter(request.ServiceName, routerKvInfo.Key)
	if ok {
		return nil, errors.New("该路由已经被注册了")
	}
	//如果不存在，那么添加新实例
	routerKvInfo = r.RouterBuffer.AddKVRouter(request.ServiceName, routerKvInfo.Key, routerKvInfo.DstInstanceID)
	//获取调整后的结果
	routerInfo, ok := r.RouterBuffer.GetServiceRouter(request.ServiceName)
	if !ok {
		err := errors.New(fmt.Sprintf("没有成功创建ServiceInfo, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}

	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(routerInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(hash, svrutil.RevisionFiled, routerInfo.Revision)
	txPipeline.HSet(hash, svrutil.InfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s, %v", routerInfo.ServiceName, routerInfo.Revision)
	return &pb.AddRouterReply{}, nil
}

func (r *Server) AddTargetRouter(hash string, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	targetRouterInfo := request.TargetRouterInfo
	_, ok := r.RouterBuffer.GetTargetRouter(request.ServiceName, targetRouterInfo.SrcInstanceID)
	if ok {
		return nil, errors.New("该路由已经被注册了")
	}
	//如果不存在，那么添加新实例
	targetRouterInfo = r.RouterBuffer.AddTargetRouter(request.ServiceName,
		targetRouterInfo.SrcInstanceID, targetRouterInfo.DstInstanceID)
	//获取调整后的结果
	routerInfo, ok := r.RouterBuffer.GetServiceRouter(request.ServiceName)
	if !ok {
		err := errors.New(fmt.Sprintf("没有成功创建ServiceInfo, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}

	//写入redis 数据和第一次的健康信息
	bytes, err := proto.Marshal(routerInfo)
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(hash, svrutil.RevisionFiled, routerInfo.Revision)
	txPipeline.HSet(hash, svrutil.InfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s, %v", routerInfo.ServiceName, routerInfo.Revision)
	return &pb.AddRouterReply{}, nil
}

func (r *Server) AddRouter(_ context.Context, request *pb.AddRouterRequest) (*pb.AddRouterReply, error) {
	hash := svrutil.RouterHash(request.ServiceName)

	mutex, err := r.LockRedis(svrutil.RouterInfoLockName(request.ServiceName))
	if err != nil {
		util.Error("lock redis failed err:%v", err)
		return nil, err
	}
	defer svrutil.TryUnlock(mutex)

	err = r.FlushRouterBuffer(hash, request.ServiceName)
	if err != nil {
		util.Error("flush buffer failed err:%v", err)
		return nil, err
	}

	switch request.RouterType {
	case util.KVRouterType:
		return r.AddKVRouter(hash, request)
	case util.TargetRouterType:
		return r.AddTargetRouter(hash, request)
	}
	return nil, errors.New("不识别的路由类型")
}

func (r *Server) RemoveTargetRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	info := request.KvRouterInfo
	//删除本地缓存
	err := r.RouterBuffer.RemoveKVRouter(request.ServiceName, info.Key)
	if err != nil {
		return nil, err
	}
	routerInfo, ok := r.RouterBuffer.GetServiceRouter(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("ServiceInfo丢失, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}
	//写入redis
	bytes, err := proto.Marshal(routerInfo)
	if err != nil {
		return nil, err
	}
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(hash, svrutil.RevisionFiled, routerInfo.Revision)
	txPipeline.HSet(hash, svrutil.InfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s, %v", routerInfo.ServiceName, routerInfo.Revision)

	return &pb.RemoveRouterReply{}, nil
}

func (r *Server) RemoveKVRouter(hash string, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	info := request.KvRouterInfo
	//删除本地缓存
	err := r.RouterBuffer.RemoveKVRouter(request.ServiceName, info.Key)
	if err != nil {
		return nil, err
	}
	routerInfo, ok := r.RouterBuffer.GetServiceRouter(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("ServiceInfo丢失, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}
	//写入redis
	bytes, err := proto.Marshal(routerInfo)
	if err != nil {
		return nil, err
	}
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(hash, svrutil.RevisionFiled, routerInfo.Revision)
	txPipeline.HSet(hash, svrutil.InfoFiled, bytes)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis路由: %s, %v", routerInfo.ServiceName, routerInfo.Revision)

	return &pb.RemoveRouterReply{}, nil
}

func (r *Server) RemoveRouter(_ context.Context, request *pb.RemoveRouterRequest) (*pb.RemoveRouterReply, error) {
	hash := svrutil.ServiceHash(request.ServiceName)

	mutex, err := r.LockRedis(svrutil.ServiceInfoLockName(request.ServiceName))
	if err != nil {
		return nil, err
	}
	defer svrutil.TryUnlock(mutex)

	//更新缓存
	err = r.FlushServiceBuffer(hash, request.ServiceName)
	if err != nil {
		return nil, err
	}

	switch request.RouterType {
	case util.KVRouterType:
		return r.RemoveKVRouter(hash, request)
	case util.TargetRouterType:
		return r.RemoveTargetRouter(hash, request)
	}
	return nil, errors.New("不能识别的路由类型")
}

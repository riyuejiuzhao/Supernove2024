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

type DeregisterContext struct {
	hash    string
	request *pb.DeregisterRequest
}

func (d *DeregisterContext) GetServiceName() string {
	return d.request.ServiceName
}

func (d *DeregisterContext) GetServiceHash() string {
	return d.hash
}

func (r *Server) Deregister(_ context.Context,
	request *pb.DeregisterRequest,
) (*pb.DeregisterReply, error) {
	deregisterCtx := &DeregisterContext{
		request: request,
		hash:    svrutil.ServiceHash(request.ServiceName),
	}

	mutex, err := r.LockRedisService(svrutil.ServiceInfoLockName(request.ServiceName))
	if err != nil {
		return nil, err
	}
	defer svrutil.TryUnlock(mutex)

	//更新缓存
	err = r.FlushBuffer(deregisterCtx)
	if err != nil {
		return nil, err
	}

	//删除本地缓存
	err = r.Mgr.RemoveInstance(request.ServiceName,
		request.Host, request.Port, request.InstanceID)
	if err != nil {
		return nil, err
	}
	serviceInfo, ok := r.Mgr.TryGetServiceInfo(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("ServiceInfo丢失, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}

	//写入redis
	healthKey := svrutil.HealthHash(serviceInfo.ServiceName, request.InstanceID)
	serviceSetName := svrutil.ServiceSetKey(serviceInfo.ServiceName)
	bytes, err := proto.Marshal(serviceInfo)
	if err != nil {
		return nil, err
	}
	//这里需要给健康信息也上锁
	healthMutex, err := r.LockRedisService(svrutil.ServiceHealthInfoLockName(request.ServiceName))
	if err != nil {
		util.Error("create lock err:%v", err)
		return nil, err
	}
	defer svrutil.TryUnlock(healthMutex)
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(deregisterCtx.GetServiceHash(), svrutil.ServiceRevisionFiled, serviceInfo.Revision)
	txPipeline.HSet(deregisterCtx.GetServiceHash(), svrutil.ServiceInfoFiled, bytes)
	// 健康信息
	txPipeline.SRem(serviceSetName, request.InstanceID)
	txPipeline.Del(healthKey)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis: %s, %v", serviceInfo.ServiceName, serviceInfo.Revision)

	return &pb.DeregisterReply{}, nil
}

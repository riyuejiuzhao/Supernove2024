package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
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

	serviceInfo, ok := r.InstanceBuffer.GetServiceInfo(request.ServiceName)
	if !ok {
		return &pb.DeregisterReply{}, nil
	}
	originRevision := serviceInfo.Revision

	//删除本地缓存
	err = r.InstanceBuffer.RemoveInstance(request.ServiceName,
		request.Host, request.Port, request.InstanceID)
	if err != nil {
		return nil, err
	}
	serviceInfo, ok = r.InstanceBuffer.GetServiceInfo(request.ServiceName)
	if !ok {
		return &pb.DeregisterReply{}, nil
	}

	if originRevision == serviceInfo.Revision {
		return &pb.DeregisterReply{}, nil
	}

	//写入redis
	healthKey := svrutil.HealthHash(serviceInfo.ServiceName, request.InstanceID)
	bytes, err := proto.Marshal(serviceInfo)
	if err != nil {
		return nil, err
	}
	//这里需要给健康信息也上锁
	healthMutex, err := r.LockRedis(svrutil.HealthInfoLockName(request.ServiceName))
	if err != nil {
		util.Error("create lock err:%v", err)
		return nil, err
	}
	defer svrutil.TryUnlock(healthMutex)
	txPipeline := r.Rdb.TxPipeline()
	txPipeline.HSet(hash, svrutil.RevisionFiled, serviceInfo.Revision)
	txPipeline.HSet(hash, svrutil.InfoFiled, bytes)
	// 健康信息
	txPipeline.Del(healthKey)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis: %s, %v", serviceInfo.ServiceName, serviceInfo.Revision)

	return &pb.DeregisterReply{}, nil
}

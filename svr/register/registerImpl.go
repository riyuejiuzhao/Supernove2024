package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"
)

type RegisterContext struct {
	Request     *pb.RegisterRequest
	ServiceHash string
	Address     string
}

func (c *RegisterContext) GetServiceName() string {
	return c.Request.ServiceName
}
func (c *RegisterContext) GetServiceHash() string {
	return c.ServiceHash
}

// Register 注册一个新服务实例
func (r *Server) Register(
	_ context.Context,
	request *pb.RegisterRequest,
) (*pb.RegisterReply, error) {
	address := svrutil.InstanceAddress(request.Host, request.Port)
	hash := svrutil.ServiceHash(request.ServiceName)

	mutex, err := r.LockRedis(svrutil.ServiceInfoLockName(request.ServiceName))
	if err != nil {
		util.Error("lock redis failed err:%v", err)
		return nil, err
	}
	defer svrutil.TryUnlock(mutex)

	err = r.FlushServiceBuffer(hash, request.ServiceName)
	if err != nil {
		util.Error("flush buffer failed err:%v", err)
		return nil, err
	}

	instanceInfo, ok := r.InstanceBuffer.GetInstanceByAddress(request.ServiceName, address)
	if ok {
		//如果存在那么直接返回数据，不进行修改
		return nil, errors.New("该地址已经被注册了")
	}
	if request.InstanceID != nil {
		instanceInfo, ok = r.InstanceBuffer.GetInstanceByID(request.ServiceName, *request.InstanceID)
		if ok {
			return nil, errors.New("该ID已经被注册了")
		}
	}

	//如果不存在，那么添加新实例
	instanceInfo = r.InstanceBuffer.AddInstance(request.ServiceName, request.Host, request.Port, request.Weight, request.InstanceID)
	//获取调整后的结果
	serviceInfo, ok := r.InstanceBuffer.GetServiceInfo(request.ServiceName)
	if !ok {
		err = errors.New(fmt.Sprintf("没有成功创建ServiceInfo, name:%s", request.ServiceName))
		util.Error("err: %v", err)
		return nil, err
	}

	//写入redis 数据和第一次的健康信息
	healthKey := svrutil.HealthHash(serviceInfo.ServiceName, instanceInfo.InstanceID)
	bytes, err := proto.Marshal(serviceInfo)
	if err != nil {
		util.Error("err: %v", err)
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
	txPipeline.HSet(healthKey, svrutil.HealthLastHeartBeatField, time.Now().Unix())
	txPipeline.HSet(healthKey, svrutil.HealthTtlFiled, request.TTL)
	_, err = txPipeline.Exec()
	if err != nil {
		return nil, err
	}
	util.Info("更新redis: %s, %v", serviceInfo.ServiceName, serviceInfo.Revision)

	return &pb.RegisterReply{InstanceID: instanceInfo.InstanceID}, nil
}

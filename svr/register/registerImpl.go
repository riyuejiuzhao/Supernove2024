package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
	"github.com/go-redis/redis"
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
) (reply *pb.RegisterReply, err error) {
	reply = nil
	address := svrutil.InstanceAddress(request.Host, request.Port)
	hash := svrutil.ServiceHash(request.ServiceName)
	set := svrutil.ServiceSet(request.ServiceName)
	var instanceID string
	if request.InstanceID != nil {
		instanceID = *request.InstanceID
	} else {
		instanceID = address
	}

	mutex, err := r.LockRedis(svrutil.ServiceInfoLockName(request.ServiceName))
	if err != nil {
		util.Error("lock redis failed err:%v", err)
		return
	}
	defer svrutil.TryUnlock(mutex)

	pipeline := r.Rdb.Pipeline()
	revisionCmd := pipeline.HGet(hash, svrutil.RevisionFiled)
	sIsMemCmd := pipeline.SIsMember(set, address)
	hExistCmd := pipeline.HExists(hash, instanceID)
	_, err = pipeline.Exec()
	var revision int64
	if errors.Is(err, redis.Nil) {
		revision = 0
	} else if err != nil {
		return
	} else {
		if sIsMemCmd.Val() || hExistCmd.Val() {
			err = errors.New("实例已经存在")
			return
		}
		revision, err = revisionCmd.Int64()
		if err != nil {
			return
		}
		revision += 1
	}

	instanceInfo := &pb.InstanceInfo{
		InstanceID: instanceID,
		Host:       request.Host,
		Port:       request.Port,
		Weight:     request.Weight,
	}

	bytes, err := proto.Marshal(instanceInfo)
	if err != nil {
		return
	}
	//写入redis 数据和第一次的健康信息
	healthKey := svrutil.HealthHash(request.ServiceName, instanceInfo.InstanceID)
	pipeline = r.Rdb.Pipeline()
	pipeline.HSet(hash, svrutil.RevisionFiled, revision)
	pipeline.HSet(hash, svrutil.InstanceIDFiled(instanceID), bytes)
	// 健康信息
	pipeline.HSet(healthKey, svrutil.HealthLastHeartBeatField, time.Now().Unix())
	pipeline.HSet(healthKey, svrutil.HealthTtlFiled, request.TTL)
	_, err = pipeline.Exec()
	if err != nil {
		return
	}
	util.Info("更新redis: %s, %v", request.ServiceName, revision)
	reply = &pb.RegisterReply{InstanceID: instanceInfo.InstanceID}
	return
}

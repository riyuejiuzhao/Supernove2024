package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
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
	var instanceID string
	if request.InstanceID != nil {
		instanceID = *request.InstanceID
	} else {
		instanceID = address
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
	pipeline := r.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HSet(hash, address, bytes)
	// 健康信息
	pipeline.HSet(healthKey, svrutil.HealthLastHeartBeatField, time.Now().Unix())
	pipeline.HSet(healthKey, svrutil.HealthTtlFiled, request.TTL)
	_, err = pipeline.Exec()
	if err != nil {
		util.Error("register pipeline err: %v", err)
		return
	}

	util.Info("更新redis: %s, %v", request.ServiceName, instanceInfo)
	reply = &pb.RegisterReply{InstanceID: instanceInfo.InstanceID}
	return
}

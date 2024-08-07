package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"errors"
)

func (r *Server) Deregister(_ context.Context,
	request *pb.DeregisterRequest,
) (reply *pb.DeregisterReply, err error) {
	reply = nil
	address := svrutil.InstanceAddress(request.Host, request.Port)
	infoField := svrutil.InstanceIDFiled(request.InstanceID)
	hash := svrutil.ServiceHash(request.ServiceName)
	set := svrutil.ServiceSet(request.ServiceName)

	mutex, err := r.LockRedis(svrutil.ServiceInfoLockName(request.ServiceName))
	if err != nil {
		return
	}
	defer svrutil.TryUnlock(mutex)

	pipeline := r.Rdb.Pipeline()
	revisionCmd := pipeline.HGet(hash, svrutil.RevisionFiled)
	sIsMem := pipeline.SIsMember(set, address)
	hExist := pipeline.HExists(hash, infoField)
	_, err = pipeline.Exec()
	if err != nil {
		return
	}
	revision, err := revisionCmd.Int64()
	if err != nil {
		return
	}
	//存在就删除
	if sIsMem.Val() || hExist.Val() {
		revision += 1
		healthKey := svrutil.HealthHash(request.ServiceName, request.InstanceID)

		pipeline = r.Rdb.Pipeline()
		pipeline.HSet(hash, svrutil.RevisionFiled, revision)
		hDelCmd := pipeline.HDel(hash, infoField)
		sRemCmd := pipeline.SRem(set, address)
		pipeline.Del(healthKey)
		_, err = pipeline.Exec()
		if err != nil {
			return
		}
		if hDelCmd.Val() != sRemCmd.Val() {
			err = errors.New("已删除，但是删除数据不一致")
			return
		}
	}
	util.Info("更新redis: %s, %v", request.ServiceName, revision)
	reply = &pb.DeregisterReply{}
	return
}

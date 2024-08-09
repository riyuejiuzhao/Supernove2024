package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
)

func (s *Server) Deregister(_ context.Context,
	request *pb.DeregisterRequest,
) (reply *pb.DeregisterReply, err error) {
	defer func() {
		const (
			Service = "Register"
			Method  = "Deregister"
		)
		s.MetricsUpload(Service, Method, request, reply)
	}()

	reply = nil
	address := svrutil.InstanceAddress(request.Host, request.Port)
	hash := svrutil.ServiceHash(request.ServiceName)
	healthKey := svrutil.HealthHash(request.ServiceName, request.InstanceID)

	pipeline := s.Rdb.Pipeline()
	pipeline.HIncrBy(hash, svrutil.RevisionFiled, 1)
	pipeline.HDel(hash, address)
	pipeline.Del(healthKey)
	_, err = pipeline.Exec()
	if err != nil {
		return
	}
	util.Info("更新redis: %s", request.ServiceName)
	reply = &pb.DeregisterReply{}
	return
}

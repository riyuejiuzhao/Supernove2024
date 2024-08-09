package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"context"
)

type Server struct {
	*svrutil.BaseServer
	pb.UnimplementedRegisterServiceServer
}

func SetupServer(
	ctx context.Context,
	address string,
	metricsAddress string,
	redisAddress string,
	redisPassword string,
	redisDB int,
) {
	baseSvr := svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)
	pb.RegisterRegisterServiceServer(baseSvr.GrpcServer,
		&Server{BaseServer: baseSvr})
	baseSvr.Setup(ctx, address, metricsAddress)
}

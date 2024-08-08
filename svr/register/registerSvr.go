package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"Supernove2024/util"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	*svrutil.BaseServer
	pb.UnimplementedRegisterServiceServer
}

func SetupServer(ctx context.Context, address string, redisAddress string, redisPassword string, redisDB int) {
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		util.Info("Stop grpc ser")
	}()
	pb.RegisterRegisterServiceServer(grpcServer,
		&Server{BaseServer: svrutil.NewBaseSvr(redisAddress, redisPassword, redisDB)})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}

}

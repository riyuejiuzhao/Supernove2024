package register

import (
	"Supernove2024/pb"
	"Supernove2024/svr/svrutil"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	*svrutil.BufferServer
	pb.UnimplementedRegisterServiceServer
}

func SetupServer(address string, redisAddress string, redisPassword string, redisDB int) {
	baseSvr := svrutil.NewBufferSvr(redisAddress, redisPassword, redisDB)
	//创建rpc服务器
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRegisterServiceServer(grpcServer,
		&Server{BufferServer: baseSvr})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

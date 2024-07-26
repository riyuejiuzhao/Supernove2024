package sdk

import (
	"google.golang.org/grpc"
	"log"
	"net"
)

type Register struct {
}

func NewRegisterAPI() {
	 = GlobalConfig()
	//创建rpc服务器
	lis, err := net.Listen("tcp", *rpcAddress)
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterTopicServiceServer(grpcServer, &TopicServer{})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

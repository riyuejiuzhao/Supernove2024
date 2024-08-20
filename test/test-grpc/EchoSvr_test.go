package test_grpc

import (
	grpc_sdk "Supernove2024/grpc-sdk"
	"Supernove2024/sdk/metrics"
	"Supernove2024/svr"
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"testing"
	"time"
)

type EchoSvr struct {
	Key    string
	status string
	*UnimplementedEchoServiceServer
}

func (s *EchoSvr) Echo(c context.Context, request *Request) (reply *Reply, err error) {
	mt, err := metrics.Instance()
	mt.InstanceRequestCount.With(prometheus.Labels{"Name": s.Key}).Inc()
	if s.Key != "CKey" {
		reply = &Reply{Content: request.Content}
		util.Info("%s recv %s", s.Key, request.Content)
	} else {
		if s.status == "" {
			s.status = "Sleep"
			go func() {
				time.Sleep(60 * time.Second)
				s.status = "Wake"
			}()
			return nil, fmt.Errorf("熔断实例")
		}
		if s.status == "Sleep" {
			return nil, fmt.Errorf("熔断实例")
		}
		reply = &Reply{Content: request.Content}
		return
	}
	return
}

func EchoSetupClient() {
	conn, err := grpc_sdk.NewClient("EchoService",
		grpc_sdk.WithGrpcDialOption(
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		))
	if err != nil {
		log.Fatalln(err)
	}

	client := NewEchoServiceClient(conn)
	count := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), 1000*time.Second)
		md := grpc_sdk.NewMetaData()
		md.SetRouterType(grpc_sdk.RandomRouterType)
		ctx = md.NewOutgoingContext(ctx)
		_, err = client.Echo(ctx, &Request{Content: fmt.Sprintf("%v", count)})
		if err != nil {
			util.Error("%v", err)
		} else {
			count += 1
		}
	}
}

func EchoSetup(selfKey, address string, opts ...grpc_sdk.ServerOption) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	svr, err := grpc_sdk.NewServer(opts...)
	if err != nil {
		log.Fatalln(err)
	}
	RegisterEchoServiceServer(svr, &EchoSvr{Key: selfKey})
	go func() {
		if err := svr.Serve(lis); err != nil {
			log.Fatalln(err)
		}
	}()

}

func TestGrpc(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("mini-router-svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	for _, addr := range []string{
		"127.0.0.1:2301",
		"127.0.0.1:2311",
		"127.0.0.1:2321",
	} {
		util.ClearEtcd(addr, t)
	}

	go func() {
		EchoSetup("AKey", "127.0.0.1:20000",
			grpc_sdk.WithInstanceName("AKey"))
	}()
	go func() {
		EchoSetup("BKey", "127.0.0.1:20001",
			grpc_sdk.WithInstanceName("BKey"))
	}()
	go func() {
		EchoSetup("CKey", "127.0.0.1:20002",
			grpc_sdk.WithInstanceName("CKey"))
	}()

	time.Sleep(1 * time.Second)
	EchoSetupClient()
}

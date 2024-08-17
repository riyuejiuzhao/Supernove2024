package test_grpc

import (
	grpc_sdk "Supernove2024/grpc-sdk"
	"Supernove2024/svr"
	"Supernove2024/util"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

type EchoSvr struct {
	Key string
	*UnimplementedEchoServiceServer
}

func (s *EchoSvr) Echo(c context.Context, request *Request) (reply *Reply, err error) {
	util.Info("%s recv %s", s.Key, request.Content)
	reply = &Reply{Content: request.Content}
	return
}

const jAddress = "127.0.0.1:6831"

func EchoSetup(selfKey, otherKey, address string, opts ...grpc_sdk.ServerOption) {
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
		util.Info("%s send %s-%v", selfKey, selfKey, count)
		ctx, _ := context.WithTimeout(context.Background(), 1000*time.Second)
		md := grpc_sdk.NewMetaData()
		md.SetRouterType(grpc_sdk.KVRouterType)
		err = md.AddRouterTag("Key", otherKey)
		if err != nil {
			util.Error("add router key err %v", err)
			continue
		}
		ctx = md.NewOutgoingContext(ctx)
		_, err = client.Echo(ctx, &Request{Content: fmt.Sprintf("%s-%v", selfKey, count)})
		if err != nil {
			util.Error("%v", err)
		} else {
			count += 1
		}
		time.Sleep(1 * time.Second)
	}
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("AKey", "BKey", "127.0.0.1:20000",
			grpc_sdk.WithInstanceName("AKey"),
			grpc_sdk.WithRouterTables([]string{"Key"}),
			grpc_sdk.WithInboundRouters([]grpc_sdk.InboundRouterOption{
				{
					NextRouterType: util.WeightedRouterType,
					Tags:           map[string]string{"Key": "AKey"},
					InstancesName:  []string{"AKey"},
				},
			}))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("BKey", "CKey", "127.0.0.1:20001",
			grpc_sdk.WithInstanceName("BKey"),
			grpc_sdk.WithRouterTables([]string{"Key"}),
			grpc_sdk.WithInboundRouters([]grpc_sdk.InboundRouterOption{
				{
					NextRouterType: util.WeightedRouterType,
					Tags:           map[string]string{"Key": "BKey"},
					InstancesName:  []string{"BKey"},
				},
			}),
		)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("CKey", "AKey", "127.0.0.1:20002",
			grpc_sdk.WithInstanceName("CKey"),
			grpc_sdk.WithRouterTables([]string{"Key"}),
			grpc_sdk.WithInboundRouters([]grpc_sdk.InboundRouterOption{
				{
					NextRouterType: util.WeightedRouterType,
					Tags:           map[string]string{"Key": "CKey"},
					InstancesName:  []string{"CKey"},
				},
			}),
		)
	}()
	wg.Wait()

	time.Sleep(1 * time.Second)
	//宕机测试
}

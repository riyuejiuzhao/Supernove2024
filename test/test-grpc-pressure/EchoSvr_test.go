package test_grpc

import (
	grpc_sdk "Supernove2024/grpc-sdk"
	"Supernove2024/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
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
		grpc_sdk.WithDefaultRouterType(grpc_sdk.KVRouterType),
		grpc_sdk.WithDefaultRouterKey(otherKey),
		grpc_sdk.WithGrpcDialOption(
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		))
	if err != nil {
		log.Fatalln(err)
	}

	client := NewEchoServiceClient(conn)
	count := 0
	for {
		_, err = client.Echo(context.Background(), &Request{Content: fmt.Sprintf("%s-%v", selfKey, count)})
		if err != nil {
			util.Error("%v", err)
		}
		count += 1
		time.Sleep(1 * time.Second)
	}
}

func ClearEtcd(address string, t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = client.Delete(ctx, "", clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
}

func TestGrpc(t *testing.T) {
	for _, addr := range []string{
		"127.0.0.1:2301",
		"127.0.0.1:2311",
		"127.0.0.1:2321",
	} {
		ClearEtcd(addr, t)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("AKey", "BKey", "127.0.0.1:20000",
			grpc_sdk.WithKVRouter([]grpc_sdk.KVRouterOption{
				{ServiceName: "EchoService", Key: "AKey"},
			}))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("BKey", "CKey", "127.0.0.1:20001",
			grpc_sdk.WithKVRouter([]grpc_sdk.KVRouterOption{
				{ServiceName: "EchoService", Key: "BKey"},
			}))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		EchoSetup("CKey", "AKey", "127.0.0.1:20002",
			grpc_sdk.WithKVRouter([]grpc_sdk.KVRouterOption{
				{ServiceName: "EchoService", Key: "CKey"},
			}))
	}()
	wg.Wait()
}

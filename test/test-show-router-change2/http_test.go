package test_show_router_change

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/metrics"
	"Supernove2024/svr"
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"testing"
	"time"
)

func SetupClient(targetService string) {
	httpClient := &http.Client{}
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		log.Fatalln(err)
	}
	var router *sdk.ProcessRouterResult = nil
	for {
		func() {
			router, err = discoveryAPI.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.KVRouterType,
				SrcInstanceName: "",
				DstService:      targetService,
				Key: map[string]string{
					"Key0": "Client",
					"Key1": "Client",
					"Key2": "Client"},
			})

			if router == nil {
				return
			}

			address := fmt.Sprintf("http://%s:%v", router.DstInstance.GetHost(), router.DstInstance.GetPort())
			req, err := http.NewRequest("GET", address, nil)
			if err != nil {
				util.Error("http get err: %v", err)
				return
			}

			response, err := httpClient.Do(req)
			if err != nil {
				util.Error("body close err: %v", err)
				return
			}
			defer response.Body.Close()
			fmt.Println("client send")
		}()
	}
}

func SetupServer(
	ctx context.Context,
	service string,
	configFile string,
	host string,
	port int32,
	name string,
	weight int32,
) {
	mt, err := metrics.Instance()
	if err != nil {
		util.Error("%v", err)
		return
	}
	//启动服务器
	address := fmt.Sprintf("%s:%v", host, port)
	config.GlobalConfigFilePath = configFile
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		log.Fatalln(err)
	}

	reply, err := registerAPI.Register(&sdk.RegisterArgv{
		ServiceName: service,
		Name:        name,
		Host:        host,
		Port:        port,
		Weight:      &weight,
	})
	if err != nil {
		log.Fatalln(err)
	} else {
		util.Info("Register %s successful", name)
	}

	defer registerAPI.Deregister(&sdk.DeregisterArgv{
		ServiceName: service,
		InstanceID:  reply.InstanceID,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mt.InstanceRequestCount.With(prometheus.Labels{"Name": name}).Inc()
		fmt.Printf("%s received\n", name)
		_, err := fmt.Fprintf(w, "Hello world")
		if err != nil {
			util.Error("http err: %v", err)
		}
	})

	srv := http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			util.Info("%v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		util.Info("%v", err)
	}
}

// 金丝雀发布
func TestHttp(t *testing.T) {
	util.ClearEtcd("127.0.0.1:2301", t)
	util.ClearEtcd("127.0.0.1:2311", t)

	go func() {
		srv, err := svr.NewConfigSvr("client-svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	config.GlobalConfigFilePath = "client.yaml"
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}

	go SetupServer(context.Background(),
		"HTTP", "client.yaml",
		"127.0.0.1", 20000,
		"A0", 30)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(context.Background(),
		"HTTP", "client.yaml",
		"127.0.0.1", 20001,
		"A1", 20)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(context.Background(),
		"HTTP", "client.yaml",
		"127.0.0.1", 20002,
		"A2", 10)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(context.Background(),
		"HTTP", "client.yaml",
		"127.0.0.1", 21000,
		"B0", 1000)
	time.Sleep(50 * time.Millisecond)

	go SetupServer(context.Background(),
		"HTTP", "client.yaml",
		"127.0.0.1", 21001,
		"B1", 50)
	time.Sleep(50 * time.Millisecond)

	ctx, _ := context.WithCancel(context.Background())

	go SetupServer(ctx,
		"HTTP", "client.yaml",
		"127.0.0.1", 21002,
		"B2", 50)
	time.Sleep(50 * time.Millisecond)

	err = registerAPI.AddTable(&sdk.AddTableArgv{
		ServiceName: "HTTP",
		Tags:        []string{"Key0"},
	})
	if err != nil {
		t.Fatal(err)
	}

	timeout := int64(10000000)
	ori0, err := registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"Key0": "Client"},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"A0", "A1", "A2"},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
		Weight:          90,
	})
	if err != nil {
		log.Fatalln(err)
	}
	rp0, err := registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"Key0": "Client"},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"B0", "B1", "B2"},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
		Weight:          10,
	})
	if err != nil {
		log.Fatalln(err)
	}

	go SetupClient("HTTP")

	time.Sleep(60 * time.Second)
	err = registerAPI.RemoveKVRouter(&sdk.RemoveKVRouterArgv{
		RouterID:       rp0.RouterID,
		DstServiceName: "HTTP",
		Dic:            map[string]string{"Key0": "Client"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	//均分
	_, err = registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"Key0": "Client"},
		DstServiceName:  "HTTP",
		DstInstanceName: []string{"B0", "B1", "B2"},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
		Weight:          90,
	})
	if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(60 * time.Second)
	err = registerAPI.RemoveKVRouter(&sdk.RemoveKVRouterArgv{
		RouterID:       ori0.RouterID,
		DstServiceName: "HTTP",
		Dic:            map[string]string{"Key0": "Client"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	for {
	}
}

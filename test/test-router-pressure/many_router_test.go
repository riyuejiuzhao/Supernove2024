package test_router_pressure_test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/svr"
	"Supernove2024/util"
	"fmt"
	"log"
	"testing"
	"time"
)

type InstanceRegisterInfo struct {
	*sdk.RegisterArgv
	InstanceID int64
}

// 设置为只有一个Cluster的情况
func TestForOneCluster(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("routers-one-svr.yaml")
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
	} {
		util.ClearEtcd(addr, t)
	}
	doTestManyRouter(t, "many_routers.yaml")
}

func TestManyRouter(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("routers-many-svr.yaml")
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
		"127.0.0.1:2331",
		"127.0.0.1:2341",
		"127.0.0.1:2351",
		"127.0.0.1:2361",
		"127.0.0.1:2371",
	} {
		util.ClearEtcd(addr, t)
	}
	doTestManyRouter(t, "many_routers.yaml")
}

// 千万级别的路由
func doTestManyRouter(t *testing.T, configFile string) {
	serviceName := "testDiscovery"
	serviceNum := 100
	config.GlobalConfigFilePath = configFile //
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SDK.Discovery.DstService = make([]string, 0)
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		cfg.SDK.Discovery.DstService = append(cfg.SDK.Discovery.DstService, nowServiceName)
	}

	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	discoveryAPI.GetInstances(&sdk.GetInstancesArgv{})

	//生成路由
	//每个服务路由个数
	serviceRouterCount := 100000
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		nowKey := "nowKey" //util.GenerateRandomString(5)
		err = registerAPI.AddTable(&sdk.AddTableArgv{
			ServiceName: nowServiceName,
			Tags:        []string{nowKey},
		})
		if err != nil {
			t.Error(err)
			continue
		}
		for j := 0; j < serviceRouterCount; j++ {
			nowVal := util.GenerateRandomString(5)
			go func() {
				timeout := int64(60 * 60 * 10)
				_, err := registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
					Dic:             map[string]string{nowKey: nowVal},
					DstServiceName:  nowServiceName,
					DstInstanceName: []string{util.GenerateRandomString(5)},
					Timeout:         &timeout,
				})
				if err != nil {
					t.Error(err)
					return
				}
			}()
			if j%5000 == 0 {
				time.Sleep(1 * time.Second)
				util.Info("i:%v j:%v", i, j)
			}
		}
	}
	util.Info("finish send")

}

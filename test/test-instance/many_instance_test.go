package test_instance

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"fmt"
	"testing"
	"time"
)

func TestManyInstance(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1:2301",
		"127.0.0.1:2311",
	} {
		util.ClearEtcd(address, t)
	}

	serviceName := "testDiscovery"
	serviceNum := 100
	instanceNum := 100
	config.GlobalConfigFilePath = "many_instance.yaml"
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

	//生成路由
	//每个服务路由个数
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < instanceNum; j++ {
			go func() {
				_, err := registerAPI.Register(&sdk.RegisterArgv{
					ServiceName: nowServiceName,
					Host:        util.RandomIP(),
					Port:        util.RandomPort(),
				})
				if err != nil {
					t.Error(err)
					return
				}
			}()
		}
		util.Info("now:%s", nowServiceName)
		time.Sleep(1 * time.Second)
	}
	util.Info("finish send")
}

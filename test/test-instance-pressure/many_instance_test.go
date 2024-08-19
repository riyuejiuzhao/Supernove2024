package test_instance_pressure

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/svr"
	"Supernove2024/util"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestManyInstance(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("many_instance-svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	for _, address := range []string{
		"127.0.0.1:2301",
		"127.0.0.1:2311",
	} {
		util.ClearEtcd(address, t)
	}

	serviceName := "testDiscovery"
	serviceNum := 10
	instanceNum := 1000
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

	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < instanceNum; j++ {
			go func() {
				mt, err := metrics.Instance()
				if err != nil {
					t.Error(err)
					return
				}
				conn, err := connMgr.NewConnManager(cfg)
				if err != nil {
					t.Error(err)
					return
				}
				registerAPI := sdk.NewRegisterAPIStandalone(cfg, conn, mt)
				_, err = registerAPI.Register(&sdk.RegisterArgv{
					ServiceName: nowServiceName,
					Host:        util.RandomIP(),
					Port:        util.RandomPort(),
				})
				if err != nil {
					t.Error(err)
					return
				}
			}()
			if j % 100 == 0 {
				time.Sleep(1*time.Second)
			}
		}
		util.Info("now:%s", nowServiceName)
		time.Sleep(1 * time.Second)
	}
	util.Info("finish send")
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}
	conn, err := connMgr.NewConnManager(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	begin := time.Now()
	dmg := dataMgr.NewServiceDataManager(cfg, conn, mt)
	util.Info("create dmg:%v", time.Now().Sub(begin).Milliseconds())
	begin = time.Now()
	dis := sdk.NewDiscoveryAPIStandalone(cfg, conn, dmg, mt)
	util.Info("create dis:%v", time.Now().Sub(begin).Milliseconds())
	begin = time.Now()
	_, err = dis.GetInstances(&sdk.GetInstancesArgv{ServiceName: ""})
	util.Info("GetInstances:%v", time.Now().Sub(begin).Milliseconds())
	time.Sleep(30 * time.Second)
}

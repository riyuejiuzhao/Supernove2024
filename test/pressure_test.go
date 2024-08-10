package test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var globalAddress = make(map[string]bool)

func RandomInstanceGlobal(serviceName string, instanceNum int) map[string]*sdk.RegisterArgv {
	testData := make(map[string]*sdk.RegisterArgv)
	for len(testData) < instanceNum {
		host := RandomIP()
		port := RandomPort()
		address := fmt.Sprintf("%s:%v", host, port)
		_, ok := globalAddress[address]
		if ok {
			continue
		}
		weight := rand.Int31n(100) + 1
		testData[address] = &sdk.RegisterArgv{
			ServiceName: serviceName,
			Host:        host,
			Port:        port,
			Weight:      &weight, //保证权重不是0
		}
		globalAddress[address] = true
	}
	return testData
}

type MiniClient struct {
	Service    string
	InstanceID string
	Host       string
	Post       int32
	configFile string
	t          *testing.T
	wg         *sync.WaitGroup
}

func (c *MiniClient) Start() {
	defer c.wg.Done()
	globalConfig, err := config.LoadConfig(c.configFile)
	if err != nil {
		c.t.Error(err)
		return
	}
	conn, err := connMgr.NewConnManager(globalConfig)
	if err != nil {
		c.t.Error(err)
		return
	}
	dmgr := dataMgr.NewDefaultServiceMgr(globalConfig, conn)

	registerAPI := sdk.NewRegisterAPIStandalone(globalConfig, conn, dmgr)
	_, err = registerAPI.Register(&sdk.RegisterArgv{
		ServiceName: c.Service,
		Host:        c.Host,
		Port:        c.Post,
	})
	if err != nil {
		c.t.Errorf("client crush: %v", err)
		return
	}

	//每过一个随机时间发送一次Get请求
	discoveryAPI := sdk.NewDiscoveryAPIStandalone(globalConfig, conn, dmgr)
	for i := 0; i < 10; i += 1 {
		dst := util.RandomItem(globalConfig.Global.Discovery.DstService)
		discoveryAPI.GetInstances(&sdk.GetInstancesArgv{ServiceName: dst})
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		runtime.GC()
	}
	util.Info("%v:%v finish", c.Service, c.InstanceID)
}

var times = 0

func newMiniClient(service string, t *testing.T, wg *sync.WaitGroup, configFile string) *MiniClient {
	times += 1
	return &MiniClient{
		Service:    service,
		InstanceID: fmt.Sprintf("client%v", times),
		Host:       RandomIP(),
		Post:       RandomPort(),
		t:          t,
		wg:         wg,
		configFile: configFile,
	}

}

func TestManyDiscovery(t *testing.T) {
	globalConfigFilePath := "many_discovery.yaml"
	serviceName := "testDiscovery"
	serviceNum := 10
	instanceNum := 1000
	testData := make([]*MiniClient, 0)
	var wg sync.WaitGroup

	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < instanceNum; j++ {
			testData = append(testData,
				newMiniClient(nowServiceName, t, &wg, globalConfigFilePath))
		}
	}

	//链接并清空数据库
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Delete(context.Background(), "", clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range testData {
		wg.Add(1)
		go func() {
			sec := time.Duration(rand.Intn(60))
			time.Sleep(sec * time.Second)
			v.Start()
		}()
	}

	wg.Wait()
}

package unit_test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
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

type MiniDiscoveryClient struct {
	Service    string
	InstanceID int64
	Host       string
	Post       int32
	cfg        config.Config
	t          *testing.T
	wg         *sync.WaitGroup
	mt         *metrics.MetricsManager
}

func (c *MiniDiscoveryClient) Start() {
	address := fmt.Sprintf("%s:%v", c.Host, c.Post)
	defer c.wg.Done()

	globalConfig := &c.cfg
	conn, err := connMgr.NewConnManager(globalConfig)
	if err != nil {
		c.t.Error(err)
		return
	}

	dmgr := dataMgr.NewDefaultServiceMgr(globalConfig, conn, c.mt)

	registerAPI := sdk.NewRegisterAPIStandalone(globalConfig, conn, dmgr, c.mt)
	reply, err := registerAPI.Register(&sdk.RegisterArgv{
		ServiceName: c.Service,
		Host:        c.Host,
		Port:        c.Post,
	})
	if err != nil {
		c.t.Errorf("client crush: %v", err)
		return
	}
	c.InstanceID = reply.InstanceID
	util.Info("cli %s Register Success", address)

	time.Sleep(60 * time.Second)
	util.Info("cli %s Start Discovery", address)

	//每过一个随机时间发送一次Get请求
	discoveryAPI := sdk.NewDiscoveryAPIStandalone(globalConfig, conn, dmgr, c.mt)
	for i := 0; i < 10; i += 1 {
		dst := util.RandomItem(globalConfig.Global.Discovery.DstService)
		reply, err := discoveryAPI.GetInstances(&sdk.GetInstancesArgv{ServiceName: dst})
		if err != nil {
			continue
		}
		util.Info("cli %s Discovery: %s %s", address, reply.ServiceName, reply.Instances)
		runtime.GC()
		time.Sleep(time.Duration(rand.Intn(20)) * time.Second)
	}

	err = registerAPI.Deregister(&sdk.DeregisterArgv{
		ServiceName: c.Service,
		InstanceID:  c.InstanceID,
	})
	if err != nil {
		c.t.Error(err)
		return
	}

	util.Info("cli %v %s finish", c.Service, address)
}

func newMiniDiscoveryClient(
	service string,
	t *testing.T,
	wg *sync.WaitGroup,
	cfg config.Config,
	mt *metrics.MetricsManager,
) *MiniDiscoveryClient {
	return &MiniDiscoveryClient{
		Service: service,
		Host:    RandomIP(),
		Post:    RandomPort(),
		t:       t,
		wg:      wg,
		cfg:     cfg,
		mt:      mt,
	}
}

func TestManyDiscovery(t *testing.T) {
	globalConfigFilePath := "many_discovery.yaml"
	config.GlobalConfigFilePath = globalConfigFilePath
	cfg, err := config.LoadConfig(globalConfigFilePath)
	if err != nil {
		t.Fatal(err)
	}
	mt := metrics.NewMetricsMgr(cfg)

	serviceName := "testDiscovery"
	serviceNum := 100
	instanceNum := 100
	testData := make([]*MiniDiscoveryClient, 0)
	var wg sync.WaitGroup

	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < instanceNum; j++ {
			testData = append(testData,
				newMiniDiscoveryClient(nowServiceName, t, &wg, *cfg, mt))
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
			sec := time.Duration(rand.Intn(120))
			time.Sleep(sec * time.Second)
			v.Start()
		}()
	}

	wg.Wait()
}

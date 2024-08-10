package test

import (
	"Supernove2024/pb"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/svr/discovery"
	"Supernove2024/svr/health"
	"Supernove2024/svr/register"
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/go-redis/redis"
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
	conn := connMgr.NewConnManager(globalConfig)
	registerAPI := sdk.NewRegisterAPIStandalone(globalConfig, conn)
	_, err = registerAPI.Register(&sdk.RegisterArgv{
		ServiceName: c.Service,
		Host:        c.Host,
		Port:        c.Post,
		InstanceID:  &c.InstanceID,
	})
	if err != nil {
		c.t.Errorf("client crush: %v", err)
		return
	}

	//每过一个随机时间发送一次Get请求
	disConn, err := conn.GetServiceConn(connMgr.Discovery)
	if err != nil {
		c.t.Error(err)
		return
	}
	disCli := pb.NewDiscoveryServiceClient(disConn)
	revisions := make(map[string]int64)
	for i := 0; i < 2; i += 1 {
		for _, dst := range globalConfig.Global.Discovery.DstService {
			revision, ok := revisions[dst]
			if !ok {
				revision = 0
			}
			_, err := disCli.GetInstances(context.Background(), &pb.GetInstancesRequest{
				ServiceName: dst,
				Revision:    revision,
			})
			if err != nil {
				c.t.Error(err)
				continue
			}
			revisions[dst] = revision
			runtime.GC()
		}
		time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
		util.Info("Get i: %v", i)
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
	ctx, cancel := context.WithCancel(context.Background())
	go register.SetupServer(ctx, "127.0.0.1:18000", "127.0.0.1:19000",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18001", "127.0.0.1:19001",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18002", "127.0.0.1:19002",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18003", "127.0.0.1:19003",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18004", "127.0.0.1:19004",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18005", "127.0.0.1:19005",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18006", "127.0.0.1:19006",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18007", "127.0.0.1:19007",
		"9.134.93.168:6380", "SDZsdz2000", 0)

	go discovery.SetupServer(ctx, "127.0.0.1:18100", "127.0.0.1:19100",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18101", "127.0.0.1:19101",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18102", "127.0.0.1:19102",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18103", "127.0.0.1:19103",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18104", "127.0.0.1:19104",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18105", "127.0.0.1:19105",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18106", "127.0.0.1:19106",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18107", "127.0.0.1:19107",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18108", "127.0.0.1:19108",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:18109", "127.0.0.1:19109",
		"9.134.93.168:6380", "SDZsdz2000", 0)

	go health.SetupServer(ctx, "127.0.0.1:18200", "127.0.0.1:19200",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	defer cancel()

	//等待服务器启动
	time.Sleep(1 * time.Second)

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
	rdb := redis.NewClient(&redis.Options{Addr: "9.134.93.168:6380", Password: "SDZsdz2000", DB: 0})
	err := rdb.FlushDB().Err()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range testData {
		wg.Add(1)
		go func() {
			sec := time.Duration(rand.Intn(30))
			time.Sleep(sec * time.Second)
			v.Start()
		}()
	}

	wg.Wait()
}

// 大量并发测试
// 10个服务
// 共计1万个实例
func TestManyInstance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go register.SetupServer(ctx, "127.0.0.1:18000", "127.0.0.1:19000",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18001", "127.0.0.1:19001",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18002", "127.0.0.1:19002",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18003", "127.0.0.1:19003",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18004", "127.0.0.1:19004",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18005", "127.0.0.1:19005",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18006", "127.0.0.1:19006",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18007", "127.0.0.1:19007",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18008", "127.0.0.1:19008",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:18009", "127.0.0.1:19009",
		"9.134.93.168:6380", "SDZsdz2000", 0)

	go discovery.SetupServer(ctx, "127.0.0.1:18100", "127.0.0.1:19100",
		"9.134.93.168:6380", "SDZsdz2000", 0)

	go health.SetupServer(ctx, "127.0.0.1:18200", "127.0.0.1:19200",
		"9.134.93.168:6380", "SDZsdz2000", 0)
	defer cancel()

	//等待服务器启动
	time.Sleep(1 * time.Second)

	serviceName := "testDiscovery"
	serviceNum := 10
	testData := make(map[string]map[string]*sdk.RegisterArgv)
	instanceNum := 1000

	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		testData[nowServiceName] = RandomInstanceGlobal(nowServiceName, instanceNum)
	}

	//链接并清空数据库
	rdb := redis.NewClient(&redis.Options{Addr: "9.134.93.168:6380", Password: "SDZsdz2000", DB: 0})
	err := rdb.FlushDB().Err()
	if err != nil {
		t.Fatal(err)
	}

	config.GlobalConfigFilePath = "many_register.yaml"
	api, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, si := range testData {
		wg.Add(1)
		go func(nowSi map[string]*sdk.RegisterArgv) {
			defer wg.Done()
			count := 0
			for _, nowIi := range nowSi {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := api.Register(nowIi)
					if err != nil {
						t.Error(err)
					}
				}()
				count += 1
				if count%500 == 0 {
					time.Sleep(time.Second)
				}
			}
		}(si)
	}
	wg.Wait()
	fmt.Println("所有服务注册结束")

	time.Sleep(1 * time.Second)

}

package test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/svr/discovery"
	"Supernove2024/svr/health"
	"Supernove2024/svr/register"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
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

// 大量并发测试
// 共计1万个实例
func TestManyService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go register.SetupServer(ctx, "127.0.0.1:8000", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8001", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8002", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8003", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8004", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8005", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8006", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8007", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8008", "9.134.93.168:6380", "SDZsdz2000", 0)
	go register.SetupServer(ctx, "127.0.0.1:8009", "9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer(ctx, "127.0.0.1:8100", "9.134.93.168:6380", "SDZsdz2000", 0)
	go health.SetupServer(ctx, "127.0.0.1:8200", "9.134.93.168:6380", "SDZsdz2000", 0)
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
				if count%200 == 0 {
					time.Sleep(time.Second)
				}
			}
		}(si)
	}
	wg.Wait()
	fmt.Println("所有服务注册结束")

	time.Sleep(1 * time.Second)

}

func TestRWMutex(t *testing.T) {

}

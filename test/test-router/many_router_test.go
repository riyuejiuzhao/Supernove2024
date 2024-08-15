package test_router_test

import (
	"Supernove2024/pb"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/util"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type InstanceRegisterInfo struct {
	*sdk.RegisterArgv
	InstanceID int64
}

func Generate(serviceName string, serviceNum int, routerCount int) (rt map[string][]*pb.KVRouterInfo) {
	rt = make(map[string][]*pb.KVRouterInfo)
	for i := 0; i < serviceNum; i++ {
		nowList := make([]*pb.KVRouterInfo, 0)
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < routerCount; j++ {
			nowKey := fmt.Sprintf("%s_%v", nowServiceName, j)
			nowVal := util.GenerateRandomString(10)
			nowList = append(nowList, &pb.KVRouterInfo{
				Key:        []string{nowKey},
				Val:        []string{nowVal},
				Timeout:    rand.Int63(),
				CreateTime: rand.Int63(),
			})
		}
		rt[nowServiceName] = nowList
	}
	return
}

func doTestMemoryForMap(generate map[string][]*pb.KVRouterInfo, t *testing.T) {
	routerBuffer := util.SyncContainer[map[string]*dataMgr.ServiceRouterBuffer]{Value: make(map[string]*dataMgr.ServiceRouterBuffer)}
	for nowServiceName, nowList := range generate {
		routerBuffer.Value[nowServiceName] = &dataMgr.ServiceRouterBuffer{
			KvRouterDic:     make(map[string]*pb.KVRouterInfo),
			TargetRouterDic: make(map[string]*pb.TargetRouterInfo),
		}
		for _, info := range nowList {
			js, err := json.Marshal(info)
			if err != nil {
				t.Fatal(err)
			}
			routerBuffer.Value[nowServiceName].KvRouterDic[string(js)] = info
		}
	}
}

func TestMemoryForMap(t *testing.T) {
	serviceName := "testDiscovery"
	serviceNum := 1
	serviceRouterCount := 100000
	result := Generate(serviceName, serviceNum, serviceRouterCount)
	doTestMemoryForMap(result, t)
}

// 设置为只有一个Cluster的情况
func TestForOneCluster(t *testing.T) {
	for _, addr := range []string{
		"127.0.0.1:2301",
		"127.0.0.1:2311",
	} {
		util.ClearEtcd(addr, t)
	}
	doTestManyRouter(t, "routers0.yaml")
}

func TestManyRouter(t *testing.T) {
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
	sdk.NewDiscoveryAPI()

	//生成路由
	//每个服务路由个数
	serviceRouterCount := 100000
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < serviceRouterCount; j++ {
			nowKey := util.GenerateRandomString(5)
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

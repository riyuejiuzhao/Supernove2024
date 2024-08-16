package test_sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"fmt"
	"log/slog"
	"math/rand"
	"testing"
)

func Generate(serviceName string, serviceNum int, routerCount int) (rt map[string][]*pb.KVRouterInfo) {
	rt = make(map[string][]*pb.KVRouterInfo)
	for i := 0; i < serviceNum; i++ {
		nowList := make([]*pb.KVRouterInfo, 0)
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < routerCount; j++ {
			nowKey := fmt.Sprintf("%s", nowServiceName)
			nowVal := util.GenerateRandomString(10)
			nowList = append(nowList, &pb.KVRouterInfo{
				Dic:             map[string]string{nowKey: nowVal},
				DstInstanceName: []string{nowServiceName},
				Timeout:         rand.Int63(),
				CreateTime:      rand.Int63(),
			})
		}
		rt[nowServiceName] = nowList
	}
	return
}

func doTestMemoryForMap(
	generate map[string][]*pb.KVRouterInfo,
	t *testing.T,
) {
	config.GlobalConfigFilePath = "sdk.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt := metrics.NewMetricsMgr(cfg)
	mgr := dataMgr.NewDefaultServiceMgr(cfg, nil, mt)
	for s, vs := range generate {
		mgr.AddRouterTable(&pb.RouterTableInfo{
			ServiceName: s,
			Tags:        []string{s},
		})
		for _, v := range vs {
			mgr.AddKVRouter(s, v)
		}
	}
}

func TestMemoryForMap(t *testing.T) {
	util.LogLevel = slog.LevelError

	serviceName := "testDiscovery"
	serviceNum := 1
	serviceRouterCount := 100000
	result := Generate(serviceName, serviceNum, serviceRouterCount)
	doTestMemoryForMap(result, t)
}

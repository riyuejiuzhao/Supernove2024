package test_sdk_pressure

import (
	"Supernove2024/pb"
	"Supernove2024/sdk"
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
			nowVal := "a" //util.GenerateRandomString(1)
			nowList = append(nowList, &pb.KVRouterInfo{
				Dic:             map[string]string{nowKey: nowVal},
				DstInstanceName: []string{nowServiceName},
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
	serviceRouterCount := 1000000
	result := Generate(serviceName, serviceNum, serviceRouterCount)
	doTestMemoryForMap(result, t)
}

var ServiceName = "testDiscovery"
var ServiceNum = 10
var InstanceNum = 1000

func GenerateInstance(dmg *dataMgr.DefaultServiceMgr, t *testing.T) {
	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		for j := 0; j < InstanceNum; j++ {
			dmg.AddInstance(nowServiceName, &pb.InstanceInfo{
				InstanceID: rand.Int63(),
				Name:       util.GenerateRandomString(10),
				Host:       util.RandomIP(),
				Port:       util.RandomPort(),
				Weight:     rand.Int31n(100),
			})
		}
	}
}

func TestRandom(t *testing.T) {
	util.LogLevel = slog.LevelError
	config.GlobalConfigFilePath = "many_instance.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}

	dmg := dataMgr.NewServiceDataManager(cfg, nil, mt)
	GenerateInstance(dmg, t)

	api := sdk.NewDiscoveryAPIStandalone(cfg, nil, dmg, mt)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		for j := 0; j < 100000; j++ {
			api.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.RandomRouterType,
				SrcInstanceName: util.GenerateRandomString(10),
				DstService:      nowServiceName,
				Key:             nil,
			})
		}
	}
}

func TestWeight(t *testing.T) {
	util.LogLevel = slog.LevelError
	config.GlobalConfigFilePath = "many_instance.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}

	dmg := dataMgr.NewServiceDataManager(cfg, nil, mt)
	GenerateInstance(dmg, t)

	api := sdk.NewDiscoveryAPIStandalone(cfg, nil, dmg, mt)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		for j := 0; j < 100000; j++ {
			api.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.WeightedRouterType,
				SrcInstanceName: util.GenerateRandomString(10),
				DstService:      nowServiceName,
				Key:             nil,
			})
		}
	}
}

func TestConsist(t *testing.T) {
	util.LogLevel = slog.LevelError
	config.GlobalConfigFilePath = "many_instance.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}

	dmg := dataMgr.NewServiceDataManager(cfg, nil, mt)
	GenerateInstance(dmg, t)

	api := sdk.NewDiscoveryAPIStandalone(cfg, nil, dmg, mt)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		for j := 0; j < 100000; j++ {
			api.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.ConsistentRouterType,
				SrcInstanceName: util.GenerateRandomString(10),
				DstService:      nowServiceName,
				Key:             nil,
			})
		}
	}
}

func GenerateTarget(t *testing.T, dmg *dataMgr.DefaultServiceMgr, api sdk.DiscoveryAPI) map[string][]string {
	rt := make(map[string][]string)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		result, err := api.GetInstances(&sdk.GetInstancesArgv{ServiceName: nowServiceName})
		if err != nil {
			t.Fatal(err)
		}
		rt[nowServiceName] = make([]string, 0)
		for j := 0; j < 1000000; j++ {
			src := util.GenerateRandomString(10)
			dmg.AddTargetRouter(nowServiceName, &pb.TargetRouterInfo{
				RouterID:        rand.Int63(),
				SrcInstanceName: src,
				DstInstanceName: util.RandomItem(result.Service.Instances).GetName(),
			})
			rt[nowServiceName] = append(rt[nowServiceName], src)
		}
	}
	return rt
}

func TestTarget(t *testing.T) {
	util.LogLevel = slog.LevelError
	config.GlobalConfigFilePath = "many_instance.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}

	dmg := dataMgr.NewServiceDataManager(cfg, nil, mt)
	GenerateInstance(dmg, t)

	api := sdk.NewDiscoveryAPIStandalone(cfg, nil, dmg, mt)
	targetDic := GenerateTarget(t, dmg, api)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		targets := targetDic[nowServiceName]
		for j := 0; j < 100000; j++ {
			_, err := api.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.TargetRouterType,
				SrcInstanceName: util.RandomItem(targets),
				DstService:      nowServiceName,
				Key:             nil,
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func GenerateKV(t *testing.T, dmg *dataMgr.DefaultServiceMgr, api sdk.DiscoveryAPI,
	table []string) map[string][]*pb.KVRouterInfo {

	rt := make(map[string][]*pb.KVRouterInfo)

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		dmg.AddRouterTable(&pb.RouterTableInfo{
			ServiceName: nowServiceName,
			Tags:        table,
		})
		result, err := api.GetInstances(&sdk.GetInstancesArgv{ServiceName: nowServiceName})
		if err != nil {
			t.Fatal(err)
		}
		rt[nowServiceName] = make([]*pb.KVRouterInfo, 0)
		for j := 0; j < 1000000; j++ {
			dic := make(map[string]string)
			for _, tag := range table {
				dic[tag] = util.GenerateRandomString(10)
			}
			router := &pb.KVRouterInfo{
				RouterID:        rand.Int63(),
				Dic:             dic,
				DstInstanceName: []string{util.RandomItem(result.GetInstance()).GetName()},
			}
			dmg.AddKVRouter(nowServiceName, router)
			rt[nowServiceName] = append(rt[nowServiceName], router)
		}
	}
	return rt
}

func TestKV(t *testing.T) {
	ServiceNum = 1
	util.LogLevel = slog.LevelError
	config.GlobalConfigFilePath = "many_instance.yaml"
	cfg, err := config.GlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	mt, err := metrics.Instance()
	if err != nil {
		t.Error(err)
		return
	}

	dmg := dataMgr.NewServiceDataManager(cfg, nil, mt)
	GenerateInstance(dmg, t)
	api := sdk.NewDiscoveryAPIStandalone(cfg, nil, dmg, mt)
	targetDic := GenerateKV(t, dmg, api, []string{"KeyA", "KeyB", "KeyC"})

	for i := 0; i < ServiceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", ServiceName, i)
		targets := targetDic[nowServiceName]
		for j := 0; j < 1000000; j++ {
			_, err := api.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:          util.KVRouterType,
				SrcInstanceName: util.GenerateRandomString(10),
				DstService:      nowServiceName,
				Key:             util.RandomItem(targets).Dic,
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

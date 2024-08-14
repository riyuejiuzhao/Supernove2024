package test_router_test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/test/unit-test"
	"Supernove2024/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

type InstanceRegisterInfo struct {
	*sdk.RegisterArgv
	InstanceID int64
}

// 千万级别的路由
func TestManyRouter(t *testing.T) {
	//链接并清空数据库
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = client.Delete(ctx, "", clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}

	serviceName := "testDiscovery"
	serviceNum := 100
	instanceNum := 100
	testData := make(map[string]map[string]*InstanceRegisterInfo)
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < instanceNum; j++ {
			nowData := unit.RandomRegisterArgv(nowServiceName, instanceNum)
			testData[nowServiceName] = util.DicMap(nowData,
				func(key string, val *sdk.RegisterArgv) (string, *InstanceRegisterInfo) {
					return key, &InstanceRegisterInfo{
						RegisterArgv: val,
						InstanceID:   0,
					}
				})
		}
	}

	config.GlobalConfigFilePath = "many_routers.yaml"
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

	count := 0
	for _, serviceV := range testData {
		for _, v := range serviceV {
			go func() {
				reply, err := registerAPI.Register(v.RegisterArgv)
				if err != nil {
					t.Error(err)
					return
				}
				v.InstanceID = reply.InstanceID
			}()
		}
		count += 1
		if count%5 == 0 {
			time.Sleep(1 * time.Second)
		}
	}
	util.Info("Register Finish")

	//生成路由
	//每个服务路由个数
	serviceRouterCount := 100000
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		nowMap, ok := testData[nowServiceName]
		if !ok {
			t.Errorf("有服务居然没有对应的请求生成 %s", nowServiceName)
			continue
		}
		for j := 0; j < serviceRouterCount; j++ {
			nowKey := fmt.Sprintf("%s_%v", nowServiceName, j)
			dst := util.RandomDicValue(nowMap)
			go func() {
				timeout := int64(60 * 60 * 10)
				_, err := registerAPI.AddKVRouter(&sdk.AddKVRouterArgv{
					Key:            nowKey,
					DstServiceName: nowServiceName,
					DstInstanceID:  dst.InstanceID,
					Timeout:        &timeout,
				})
				if err != nil {
					t.Error(err)
					return
				}
			}()
			if j%2000 == 0 {
				time.Sleep(1 * time.Second)
				util.Error("i:%v j:%v", i, j)
			}
		}
	}
	util.Error("finish send")

	//随机路由检测
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < serviceNum; i++ {
		nowServiceName := fmt.Sprintf("%s%v", serviceName, i)
		for j := 0; j < serviceRouterCount; j++ {
			nowKey := fmt.Sprintf("%s_%v", nowServiceName, j)
			_, err = discoveryAPI.ProcessRouter(&sdk.ProcessRouterArgv{
				Method:        util.KVRouterType,
				SrcInstanceID: 0,
				DstService:    &sdk.DefaultDstService{ServiceName: nowServiceName},
				Key:           nowKey,
			})
			if err != nil {
				t.Error(err)
			}
		}
	}
	time.Sleep(15 * time.Second)
}

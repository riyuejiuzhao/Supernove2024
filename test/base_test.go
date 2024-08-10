package test

import (
	"Supernove2024/pb"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"math/rand"
	"testing"
	"time"
)

// RandomServiceAddress
// 随机生成服务地址
func RandomIP() string {
	return fmt.Sprintf("%v.%v.%v.%v",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256))
}

func RandomPort() int32 {
	return rand.Int31n(65536)
}

func RandomRegisterArgv(serviceName string, instanceNum int) map[string]*sdk.RegisterArgv {
	testData := make(map[string]*sdk.RegisterArgv)
	for len(testData) < instanceNum {
		host := RandomIP()
		port := RandomPort()
		address := fmt.Sprintf("%s:%v", host, port)
		_, ok := testData[address]
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
	}
	return testData
}

func TestRouter(t *testing.T) {
	srcService := "srcService"
	dstService := "dstService"
	instanceNum := 10

	srcData := RandomRegisterArgv(srcService, instanceNum)
	dstData := RandomRegisterArgv(dstService, instanceNum)

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

	config.GlobalConfigFilePath = "router_test.yaml"
	// 注册
	registerapi, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range srcData {
		_, err := registerapi.Register(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range dstData {
		_, err := registerapi.Register(v)
		if err != nil {
			t.Fatal(err)
		}
	}

	//获取
	discoveryapi, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	srcServices, err := discoveryapi.GetInstances(&sdk.GetInstancesArgv{
		ServiceName: srcService,
	})
	if err != nil {
		t.Fatal(err)
	}
	dstServices, err := discoveryapi.GetInstances(&sdk.GetInstancesArgv{
		ServiceName: dstService,
	})
	if err != nil {
		t.Fatal(err)
	}
	//先检测其他三种路由正不正常
	processResult, err := discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:        util.ConsistentRouterType,
		SrcInstanceID: srcServices.Instances[0].InstanceID,
		DstService:    dstServices,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("一致性哈希:%s", processResult.DstInstance)
	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:        util.RandomRouterType,
		SrcInstanceID: srcServices.Instances[0].InstanceID,
		DstService:    dstServices,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("随机:%s", processResult.DstInstance)
	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:        util.WeightedRouterType,
		SrcInstanceID: srcServices.Instances[0].InstanceID,
		DstService:    dstServices,
	})
	if err != nil {
		t.Fatal(err)
	}
	maxInstance := dstServices.Instances[0]
	for _, v := range dstServices.Instances {
		if v.Weight > maxInstance.Weight {
			maxInstance = v
		}
	}
	if maxInstance.InstanceID != processResult.DstInstance.InstanceID {
		t.Fatal("权重不是最大的那个")
	}
	t.Logf("权重:%s", processResult.DstInstance)

	//注册路由
	// src0 -> dst1
	// key:key0 -> dst2
	err = registerapi.AddTargetRouter(&sdk.AddTargetRouterArgv{
		SrcInstanceID:  srcServices.Instances[0].InstanceID,
		DstServiceName: dstService,
		DstInstanceID:  dstServices.Instances[1].InstanceID,
		Timeout:        100,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = registerapi.AddKVRouter(&sdk.AddKVRouterArgv{
		Key:            "key0",
		DstServiceName: dstService,
		DstInstanceID:  dstServices.Instances[2].InstanceID,
		Timeout:        100,
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:        util.TargetRouterType,
		SrcInstanceID: srcServices.Instances[0].InstanceID,
		DstService:    dstServices,
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.InstanceID != dstServices.Instances[1].InstanceID {
		t.Fatalf("路由错误")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:        util.KVRouterType,
		SrcInstanceID: srcServices.Instances[2].InstanceID, //"src2",
		DstService:    dstServices,
		Key:           "key0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.InstanceID != dstServices.Instances[2].InstanceID {
		t.Fatalf("路由错误")
	}
}

func TestHealthSvr(t *testing.T) {
	serviceName := "testDiscovery"
	instanceNum := 10
	testData := RandomRegisterArgv(serviceName, instanceNum)
	//连接数据库
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

	config.GlobalConfigFilePath = "health_test.yaml"
	registerAPI, err := sdk.NewRegisterAPI()
	//等待一下更新
	time.Sleep(1 * time.Second)

	if err != nil {
		t.Fatal(err)
	}
	resultList := make([]*sdk.RegisterResult, 0)
	for _, v := range testData {
		result, err := registerAPI.Register(&sdk.RegisterArgv{
			ServiceName: v.ServiceName,
			Host:        v.Host,
			Port:        v.Port})
		if err != nil {
			t.Fatal(err)
		}
		resultList = append(resultList, result)
	}

	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	//睡一秒等更新
	getInstanceArgv := &sdk.GetInstancesArgv{ServiceName: serviceName}
	time.Sleep(1 * time.Second)
	getInstanceResult, err := discoveryAPI.GetInstances(getInstanceArgv)
	if err != nil {
		t.Fatal(err)
	}
	if len(getInstanceResult.GetInstance()) != len(testData) {
		t.Fatal("没有获得全部的服务")
	}
	//等待6秒让所有的服务都过期
	time.Sleep(7 * time.Second)
	getInstanceResult, err = discoveryAPI.GetInstances(getInstanceArgv)
	if err != nil {
		t.Fatal(err)
	}
	if len(getInstanceResult.GetInstance()) != 0 {
		t.Fatal("服务健康信息不正确")
	}
	//通过心跳再次激活所有的服务
	healthCli, err := sdk.NewHealthAPI()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range resultList {
		err = healthCli.HeartBeat(&sdk.HeartBeatArgv{ServiceName: serviceName, InstanceID: v.InstanceID})
		if err != nil {
			t.Error("租约没有正常超时")
		}
	}
}

func TestDiscoverSvr(t *testing.T) {
	serviceName := "testDiscovery"
	instanceNum := 10

	testData := RandomRegisterArgv(serviceName, instanceNum)
	//连接数据库
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

	config.GlobalConfigFilePath = "register_test.yaml"
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}

	resultData := make(map[int64]*sdk.RegisterArgv)
	for _, v := range testData {
		result, err := registerAPI.Register(v)
		if err != nil {
			t.Fatal(err)
		}
		resultData[result.InstanceID] = v
	}

	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	getInstanceArgv := &sdk.GetInstancesArgv{ServiceName: serviceName}
	disRt, err := discoveryAPI.GetInstances(getInstanceArgv)
	if err != nil {
		t.Fatal(err)
	}

	disDic := make(map[string]*pb.InstanceInfo)
	for _, v := range disRt.GetInstance() {
		addrV := fmt.Sprintf("%s:%v", v.Host, v.Port)
		disDic[addrV] = v
	}
	//确定双方不匹配的地方
	for addr, v := range testData {
		disV, ok := disDic[addr]
		if !ok {
			t.Errorf("discovery中缺少 Address:%s", addr)
		}
		if disV.Weight != *v.Weight ||
			disV.Host != v.Host ||
			disV.Port != v.Port {
			t.Errorf("数据不匹配，发送数据为host:%s, port:%v, weight:%v, 发现数据为host:%s, port:%v, weight:%v",
				v.Host, v.Port, v.Weight,
				disV.Host, disV.Port, disV.Weight)
		}
	}

	// 上面已经确定testData <= disRt
	// 如果相同那么就一定一样
	if len(disRt.GetInstance()) == len(testData) {
		return
	}

	for addr, v := range disDic {
		testV, ok := testData[addr]
		if !ok {
			t.Errorf("discovery中多出了Address:%s", addr)
		}
		if *testV.Weight != v.Weight ||
			testV.Host != v.Host ||
			testV.Port != v.Port {
			t.Errorf("数据不匹配，发送数据为host:%s, port:%v, weight:%v, 发现数据为host:%s, port:%v, weight:%v",
				testV.Host, testV.Port, testV.Weight,
				v.Host, v.Port, v.Weight)
		}
	}
}

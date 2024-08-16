package unit_test

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/svr"
	"Supernove2024/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"math/rand"
	"testing"
	"time"
)

func RandomRegisterArgv(serviceName string, instanceNum int) map[string]*sdk.RegisterArgv {
	testData := make(map[string]*sdk.RegisterArgv)
	for len(testData) < instanceNum {
		host := util.RandomIP()
		port := util.RandomPort()
		address := fmt.Sprintf("%s:%v", host, port)
		_, ok := testData[address]
		if ok {
			continue
		}
		weight := rand.Int31n(100) + 1
		testData[address] = &sdk.RegisterArgv{
			ServiceName: serviceName,
			Name:        address,
			Host:        host,
			Port:        port,
			Weight:      &weight, //保证权重不是0
		}
	}
	return testData
}

func TestRouter(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("router_test_svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	srcService := "srcService"
	dstService := "dstService"
	instanceNum := 10

	srcData := RandomRegisterArgv(srcService, instanceNum)
	dstData := RandomRegisterArgv(dstService, instanceNum)

	util.ClearEtcd("127.0.0.1:2311", t)
	util.ClearEtcd("127.0.0.1:2301", t)

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
		Method:          util.ConsistentRouterType,
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstService:      dstServices.GetServiceName(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("一致性哈希:%s", processResult.DstInstance)
	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.RandomRouterType,
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstService:      dstServices.GetServiceName(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("随机:%s", processResult.DstInstance)
	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.WeightedRouterType,
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstService:      dstServices.GetServiceName(),
	})
	if err != nil {
		t.Fatal(err)
	}
	maxInstance := dstServices.GetInstance()[0]
	for _, v := range dstServices.GetInstance() {
		if v.GetWeight() > maxInstance.GetWeight() {
			maxInstance = v
		}
	}
	t.Logf("权重:%s weight:%v", processResult.DstInstance, processResult.DstInstance.GetWeight())

	//注册路由
	err = registerapi.AddTable(&sdk.AddTableArgv{
		ServiceName: dstService,
		Tags:        []string{"key0", "key1", "key2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// src0 -> dst1
	// key:key0 -> dst2
	timeout := int64(100)
	targetReply, err := registerapi.AddTargetRouter(&sdk.AddTargetRouterArgv{
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstServiceName:  dstService,
		DstInstanceName: dstServices.GetInstance()[1].GetName(),
		Timeout:         &timeout,
	})
	if err != nil {
		t.Fatal(err)
	}
	kvReply, err := registerapi.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"key0": "val0"},
		DstServiceName:  dstService,
		DstInstanceName: []string{dstServices.GetInstance()[0].GetName()},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
	})
	if err != nil {
		t.Fatal(err)
	}
	kvReply, err = registerapi.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"key0": "val0", "key1": "val1"},
		DstServiceName:  dstService,
		DstInstanceName: []string{dstServices.GetInstance()[1].GetName()},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
	})
	if err != nil {
		t.Fatal(err)
	}
	kvReply, err = registerapi.AddKVRouter(&sdk.AddKVRouterArgv{
		Dic:             map[string]string{"key0": "val0", "key1": "val1", "key2": "val2"},
		DstServiceName:  dstService,
		DstInstanceName: []string{dstServices.GetInstance()[2].GetName()},
		Timeout:         &timeout,
		NextRouterType:  util.RandomRouterType,
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.TargetRouterType,
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstService:      dstServices.GetServiceName(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.GetInstanceID() != dstServices.GetInstance()[1].GetInstanceID() {
		t.Fatalf("路由错误")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.KVRouterType,
		SrcInstanceName: srcServices.GetInstance()[2].GetName(), //"src2",
		DstService:      dstServices.GetServiceName(),
		Key:             map[string]string{"key0": "val0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.GetInstanceID() != dstServices.GetInstance()[0].GetInstanceID() {
		t.Fatalf("路由错误")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.KVRouterType,
		SrcInstanceName: srcServices.GetInstance()[2].GetName(), //"src2",
		DstService:      dstServices.GetServiceName(),
		Key:             map[string]string{"key0": "val0", "key1": "val1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.GetInstanceID() != dstServices.GetInstance()[1].GetInstanceID() {
		t.Fatalf("路由错误")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.KVRouterType,
		SrcInstanceName: srcServices.GetInstance()[2].GetName(), //"src2",
		DstService:      dstServices.GetServiceName(),
		Key:             map[string]string{"key0": "val0", "key1": "val2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.GetInstanceID() != dstServices.GetInstance()[0].GetInstanceID() {
		t.Fatalf("路由错误")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.KVRouterType,
		SrcInstanceName: srcServices.GetInstance()[2].GetName(), //"src2",
		DstService:      dstServices.GetServiceName(),
		Key:             map[string]string{"key0": "val0", "key1": "val1", "key2": "val2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if processResult.DstInstance.GetInstanceID() != dstServices.GetInstance()[2].GetInstanceID() {
		t.Fatalf("路由错误")
	}

	err = registerapi.RemoveKVRouter(&sdk.RemoveKVRouterArgv{
		RouterID:       kvReply.RouterID,
		DstServiceName: dstService,
		Dic:            map[string]string{"key": "key0"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = registerapi.RemoveTargetRouter(&sdk.RemoveTargetRouterArgv{
		RouterID:    targetReply.RouterID,
		DstService:  dstService,
		SrcInstance: srcServices.GetInstance()[0].GetName(),
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.KVRouterType,
		SrcInstanceName: srcServices.GetInstance()[2].GetName(), //"src2",
		DstService:      dstServices.GetServiceName(),
		Key:             map[string]string{"key": "key0"},
	})
	if err == nil {
		t.Fatal("没有删除路由")
	}

	processResult, err = discoveryapi.ProcessRouter(&sdk.ProcessRouterArgv{
		Method:          util.TargetRouterType,
		SrcInstanceName: srcServices.GetInstance()[0].GetName(),
		DstService:      dstServices.GetServiceName(),
	})
	if err == nil {
		t.Fatal("没有删除路由")
	}

}

func TestHealthSvr(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("health_test_svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	serviceName := "testDiscovery"
	instanceNum := 10
	testData := RandomRegisterArgv(serviceName, instanceNum)
	//连接数据库
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2301"},
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
}

func TestDiscoverSvr(t *testing.T) {
	go func() {
		srv, err := svr.NewConfigSvr("register_test_svr.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		if err = srv.Serve("127.0.0.1:30000"); err != nil {
			log.Fatalln(err)
		}
	}()
	//等服务器启动
	time.Sleep(1 * time.Second)

	serviceName := "testDiscovery"
	instanceNum := 10

	testData := RandomRegisterArgv(serviceName, instanceNum)
	//连接数据库
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2301"},
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

	time.Sleep(1 * time.Second)

	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}
	getInstanceArgv := &sdk.GetInstancesArgv{ServiceName: serviceName}
	disRt, err := discoveryAPI.GetInstances(getInstanceArgv)
	if err != nil {
		t.Fatal(err)
	}

	disDic := make(map[string]util.DstInstanceInfo)
	for _, v := range disRt.GetInstance() {
		addrV := fmt.Sprintf("%s:%v", v.GetHost(), v.GetPort())
		disDic[addrV] = v
	}
	//确定双方不匹配的地方
	for addr, v := range testData {
		disV, ok := disDic[addr]
		if !ok {
			t.Errorf("discovery中缺少 Name:%s", addr)
		}
		if disV.GetWeight() != *v.Weight ||
			disV.GetHost() != v.Host ||
			disV.GetPort() != v.Port {
			t.Errorf("数据不匹配，发送数据为host:%s, port:%v, weight:%v, 发现数据为host:%s, port:%v, weight:%v",
				v.Host, v.Port, v.Weight,
				disV.GetHost(), disV.GetPort(), disV.GetWeight())
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
		if *testV.Weight != v.GetWeight() ||
			testV.Host != v.GetHost() ||
			testV.Port != v.GetPort() {
			t.Errorf("数据不匹配，发送数据为host:%s, port:%v, weight:%v, 发现数据为host:%s, port:%v, weight:%v",
				testV.Host, testV.Port, testV.Weight,
				v.GetHost(), v.GetPort(), v.GetWeight())
		}
	}
}

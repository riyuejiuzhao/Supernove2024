package test

import (
	"Supernove2024/pb"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/svr/discovery"
	"Supernove2024/svr/register"
	"Supernove2024/svr/svrutil"
	"fmt"
	"github.com/go-redis/redis"
	"google.golang.org/protobuf/proto"
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
	for i := 0; i < instanceNum; i++ {
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

func TestDiscoverSvr(t *testing.T) {
	serviceName := "testDiscovery"
	instanceNum := 10

	testData := RandomRegisterArgv(serviceName, instanceNum)
	//连接数据库
	rdb := redis.NewClient(&redis.Options{Addr: "9.134.93.168:6380", Password: "SDZsdz2000", DB: 0})
	//清空
	err := rdb.FlushDB().Err()
	if err != nil {
		t.Fatal(err)
	}

	go register.SetupServer("127.0.0.1:8080", "9.134.93.168:6380", "SDZsdz2000", 0)
	go discovery.SetupServer("127.0.0.1:9090", "9.134.93.168:6380", "SDZsdz2000", 0)

	//等待服务器启动
	time.Sleep(1 * time.Second)

	config.DefaultConfigFilePath = "register_test.yaml"
	registerAPI, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		t.Fatal(err)
	}

	resultData := make(map[string]*sdk.RegisterArgv)
	for _, v := range testData {
		result, err := registerAPI.Register(v)
		if err != nil {
			t.Fatal(err)
		}
		resultData[result.InstanceID] = v
	}

	disRt, err := discoveryAPI.GetInstances(serviceName)
	if err != nil {
		t.Fatal(err)
	}

	disDic := make(map[string]*pb.InstanceInfo)
	for _, v := range disRt.Instances {
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
	if len(disRt.Instances) == len(testData) {
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

func TestRegisterSvr(t *testing.T) {
	serviceName := "testRegister"
	instanceNum := 10

	testData := RandomRegisterArgv(serviceName, instanceNum)
	//链接并清空数据库
	rdb := redis.NewClient(&redis.Options{Addr: "9.134.93.168:6380", Password: "SDZsdz2000", DB: 0})
	err := rdb.FlushDB().Err()
	if err != nil {
		t.Fatal(err)
	}

	// 从另一个协程启动
	go register.SetupServer("127.0.0.1:8080", "9.134.93.168:6380", "SDZsdz2000", 0)

	//等待服务器启动
	time.Sleep(1 * time.Second)

	config.DefaultConfigFilePath = "register_test.yaml"
	api, err := sdk.NewRegisterAPI()
	if err != nil {
		t.Fatal(err)
	}

	resultData := make(map[string]*sdk.RegisterArgv)
	for _, v := range testData {
		result, err := api.Register(v)
		if err != nil {
			t.Fatal(err)
		}
		resultData[result.InstanceID] = v
	}

	//检查redis数据库是否全部写入了
	revision, err := rdb.HGet(svrutil.ServiceHash(serviceName), svrutil.ServiceRevisionFiled).Int64()
	if err != nil {
		t.Fatal(err)
	}
	if revision != int64(len(testData)) {
		t.Fatal("服务版本号没有正常更新")
	}
	bytes, err := rdb.HGet(svrutil.ServiceHash(serviceName), svrutil.ServiceInfoFiled).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var serviceInfo pb.ServiceInfo
	err = proto.Unmarshal(bytes, &serviceInfo)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range serviceInfo.Instances {
		dataInTest, ok := resultData[v.InstanceID]
		if !ok {
			t.Errorf("服务没有正常写入 ID:%v", v.InstanceID)
		}
		if dataInTest.Host != v.Host ||
			dataInTest.Port != v.Port ||
			*dataInTest.Weight != v.Weight {
			t.Errorf("ID:%v 发送的数据为{Host:%s,Port:%v,Weight:%v}, redis数据为{Host:%s,Port:%v,Weight:%v}",
				v.InstanceID,
				dataInTest.Host, dataInTest.Port, dataInTest.Weight,
				v.Host, v.Port, v.Weight)
		}
	}

	//逐个删除
	targetRevision := revision + int64(len(serviceInfo.Instances))
	for _, v := range serviceInfo.Instances {
		err = api.Deregister(&sdk.DeregisterArgv{
			ServiceName: serviceName,
			Host:        v.Host,
			Port:        v.Port,
			InstanceID:  v.InstanceID,
		})
		if err != nil {
			t.Error(err)
		}
	}
	revision, err = rdb.HGet(svrutil.ServiceHash(serviceName), svrutil.ServiceRevisionFiled).Int64()
	if err != nil {
		t.Fatal(err)
	}
	if revision != targetRevision {
		t.Fatal("服务版本号没有正常更新")
	}
	bytes, err = rdb.HGet(svrutil.ServiceHash(serviceName), svrutil.ServiceInfoFiled).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	err = proto.Unmarshal(bytes, &serviceInfo)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range serviceInfo.Instances {
		t.Errorf("未删除 id:%s, host:%s, port:%v", v.InstanceID, v.Host, v.Port)
	}

}

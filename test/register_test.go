package test

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
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

func TestRegisterSvr(t *testing.T) {
	serviceName := "testService1"
	instanceNum := 10

	testData := make([]*sdk.RegisterArgv, 0)
	for i := 0; i < instanceNum; i++ {
		testData = append(testData, &sdk.RegisterArgv{
			ServiceName: serviceName,
			Host:        RandomIP(),
			Port:        RandomPort(),
			Weight:      rand.Int31n(100) + 1, //保证权重不是0
		})
	}
	//链接并清空数据库
	rdb := redis.NewClient(&redis.Options{Addr: "9.134.93.168:6380", Password: "SDZsdz2000", DB: 0})
	err := rdb.FlushDB().Err()
	if err != nil {
		t.Fatal(err)
	}

	// 从另一个协程启动
	go register.SetupServer("127.0.0.1:8080", "9.134.93.168:6380", "SDZsdz2000", 0)

	//等待服务器启动
	time.Sleep(10 * time.Second)

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
	var serviceInfo miniRouterProto.ServiceInfo
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
			dataInTest.Weight != v.Weight {
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

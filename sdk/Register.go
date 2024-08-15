package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"time"
)

type RegisterCli struct {
	*APIContext
}

type RegisterArgv struct {
	ServiceName string
	Host        string
	Port        int32
	Name        string

	//optional
	Weight *int32
	TTL    *int64
}

type RegisterResult struct {
	InstanceID int64
}

type DeregisterArgv struct {
	ServiceName string
	InstanceID  int64
}

func (c *RegisterCli) Register(service *RegisterArgv) (result *RegisterResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "Register"}, err)
	}()
	var weight int32
	if service.Weight != nil {
		weight = *service.Weight
	} else {
		weight = c.Config.SDK.Register.DefaultWeight
	}

	var ttl int64
	if service.TTL != nil {
		ttl = *service.TTL
	} else {
		ttl = c.Config.SDK.Register.DefaultServiceTTL
	}

	client, err := c.ConnManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		return
	}
	resp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return
	}

	instanceInfo := &pb.InstanceInfo{
		InstanceID: int64(resp.ID),
		Name:       service.Name,
		Host:       service.Host,
		Port:       service.Port,
		Weight:     weight,
		CreateTime: time.Now().Unix(),
	}

	key := util.InstanceKey(service.ServiceName, instanceInfo.InstanceID)
	bytes, err := proto.Marshal(instanceInfo)
	if err != nil {
		return
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	result = &RegisterResult{InstanceID: instanceInfo.InstanceID}
	return
}

func (c *RegisterCli) Deregister(service *DeregisterArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "Deregister"}, err)
	}()
	client, err := c.ConnManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		return
	}
	_, err = client.Revoke(context.Background(), clientv3.LeaseID(service.InstanceID))
	if err != nil {
		return
	}
	return
}

type AddTargetRouterArgv struct {
	SrcInstanceName string
	DstServiceName  string
	DstInstanceName string
	Timeout         *int64
}

type AddKVRouterArgv struct {
	Dic             map[string]string
	DstServiceName  string
	DstInstanceName []string
	Timeout         *int64
	NextRouterType  int32
}

func (c *RegisterCli) AddTargetRouter(argv *AddTargetRouterArgv) (result *AddRouterResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "AddTargetRouter"}, err)
	}()
	client, err := c.ConnManager.GetServiceConn(connMgr.RoutersEtcd,
		util.TargetRouterHashSlotKey(argv.DstServiceName, argv.SrcInstanceName))
	if err != nil {
		return
	}
	var timeout int64
	if argv.Timeout != nil {
		timeout = *argv.Timeout
	} else {
		timeout = c.Config.SDK.Register.DefaultRouterTTL
	}
	resp, err := client.Grant(context.Background(), timeout)
	if err != nil {
		return
	}

	info := &pb.TargetRouterInfo{
		RouterID:        int64(resp.ID),
		SrcInstanceName: argv.SrcInstanceName,
		DstInstanceName: argv.DstInstanceName,
		Timeout:         *argv.Timeout,
		CreateTime:      time.Now().Unix(),
	}

	bytes, err := proto.Marshal(info)
	if err != nil {
		return
	}
	key := util.RouterTargetInfoKey(argv.DstServiceName, int64(resp.ID))
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	result = &AddRouterResult{RouterID: int64(resp.ID)}
	return
}

type AddRouterResult struct {
	RouterID int64
}

func (c *RegisterCli) AddKVRouter(argv *AddKVRouterArgv) (result *AddRouterResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "AddKVRouter"}, err)
	}()

	var timeout int64
	if argv.Timeout == nil {
		timeout = c.Config.SDK.Register.DefaultRouterTTL
	} else {
		timeout = *argv.Timeout
	}

	if argv.NextRouterType != util.RandomRouterType &&
		argv.NextRouterType != util.WeightedRouterType &&
		argv.NextRouterType != util.ConsistentRouterType {
		return nil, fmt.Errorf("不支持的NextRouterType:%v", argv.NextRouterType)
	}

	client, err := c.ConnManager.GetServiceConn(
		connMgr.RoutersEtcd,
		util.KvRouterHashSlotKey(argv.DstServiceName, argv.Dic))
	if err != nil {
		return
	}

	resp, err := client.Grant(context.Background(), timeout)
	if err != nil {
		return
	}

	keys := make([]string, 0)
	vals := make([]string, 0)
	for k, v := range argv.Dic {
		keys = append(keys, k)
		vals = append(vals, v)
	}

	info := &pb.KVRouterInfo{
		RouterID:        int64(resp.ID),
		Key:             keys,
		Val:             vals,
		DstInstanceName: argv.DstInstanceName,
		Timeout:         timeout,
		CreateTime:      time.Now().Unix(),
		RouterType:      argv.NextRouterType,
	}
	key := util.RouterKVInfoKey(argv.DstServiceName, int64(resp.ID))

	bytes, err := proto.Marshal(info)
	if err != nil {
		return
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	result = &AddRouterResult{RouterID: int64(resp.ID)}
	return
}

type RemoveKVRouterArgv struct {
	RouterID       int64
	DstServiceName string
	Dic            map[string]string
}

func (c *RegisterCli) RemoveKVRouter(argv *RemoveKVRouterArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "RemoveKVRouter"}, err)
	}()

	client, err := c.ConnManager.GetServiceConn(
		connMgr.RoutersEtcd,
		util.KvRouterHashSlotKey(argv.DstServiceName, argv.Dic))
	if err != nil {
		return
	}

	_, err = client.Revoke(context.Background(), clientv3.LeaseID(argv.RouterID))
	if err != nil {
		return
	}
	return
}

type RemoveTargetRouterArgv struct {
	RouterID    int64
	DstService  string
	SrcInstance string
}

func (c *RegisterCli) RemoveTargetRouter(argv *RemoveTargetRouterArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "RemoveKVRouter"}, err)
	}()

	client, err := c.ConnManager.GetServiceConn(
		connMgr.RoutersEtcd,
		util.TargetRouterHashSlotKey(argv.DstService, argv.SrcInstance))
	if err != nil {
		return
	}
	_, err = client.Revoke(context.Background(), clientv3.LeaseID(argv.RouterID))
	if err != nil {
		return
	}
	return
}

// RegisterAPI 功能：
// 服务注册
type RegisterAPI interface {
	Register(service *RegisterArgv) (*RegisterResult, error)
	Deregister(service *DeregisterArgv) error
	AddTargetRouter(*AddTargetRouterArgv) (result *AddRouterResult, err error)
	AddKVRouter(argv *AddKVRouterArgv) (result *AddRouterResult, err error)
	RemoveKVRouter(argv *RemoveKVRouterArgv) error
	RemoveTargetRouter(argv *RemoveTargetRouterArgv) error
}

func NewRegisterAPIStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
	dmgr dataMgr.ServiceDataManager,
	mt *metrics.MetricsManager,
) RegisterAPI {
	ctx := NewAPIContextStandalone(config, conn, dmgr, mt)
	return &RegisterCli{
		APIContext: ctx,
	}
}

func NewRegisterAPI() (RegisterAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	return &RegisterCli{ctx}, nil
}

package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"context"
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
	key := util.InstanceKey(service.ServiceName, service.Host, service.Port)
	var weight int32
	if service.Weight != nil {
		weight = *service.Weight
	} else {
		weight = c.Config.Global.Register.DefaultWeight
	}

	var ttl int64
	if service.TTL != nil {
		ttl = *service.TTL
	} else {
		ttl = c.Config.Global.Register.DefaultTTL
	}

	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return
	}
	resp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return
	}

	instanceInfo := &pb.InstanceInfo{
		InstanceID: int64(resp.ID),
		Host:       service.Host,
		Port:       service.Port,
		Weight:     weight,
		CreateTime: time.Now().Unix(),
	}

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
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
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
	SrcInstanceID  int64
	DstServiceName string
	DstInstanceID  int64
	Timeout        int64
}

type AddKVRouterArgv struct {
	Key            string
	DstServiceName string
	DstInstanceID  int64
	Timeout        int64
}

func (c *RegisterCli) AddTargetRouter(argv *AddTargetRouterArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "AddTargetRouter"}, err)
	}()
	key := util.RouterTargetInfoKey(argv.DstServiceName, argv.SrcInstanceID)
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return
	}
	resp, err := client.Grant(context.Background(), argv.Timeout)
	if err != nil {
		return
	}

	info := &pb.TargetRouterInfo{
		RouterID:      int64(resp.ID),
		SrcInstanceID: argv.SrcInstanceID,
		DstInstanceID: argv.DstInstanceID,
		Timeout:       argv.Timeout,
		CreateTime:    time.Now().Unix(),
	}

	bytes, err := proto.Marshal(info)
	if err != nil {
		return
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	return
}

func (c *RegisterCli) AddKVRouter(argv *AddKVRouterArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "AddKVRouter"}, err)
	}()
	key := util.RouterKVInfoKey(argv.DstServiceName, argv.Key)
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return
	}
	resp, err := client.Grant(context.Background(), argv.Timeout)
	if err != nil {
		return
	}

	info := &pb.KVRouterInfo{
		RouterID:      int64(resp.ID),
		Key:           argv.Key,
		DstInstanceID: argv.DstInstanceID,
		Timeout:       argv.Timeout,
		CreateTime:    time.Now().Unix(),
	}

	bytes, err := proto.Marshal(info)
	if err != nil {
		return
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
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
	AddTargetRouter(*AddTargetRouterArgv) error
	AddKVRouter(argv *AddKVRouterArgv) error
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

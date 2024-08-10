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

var registerLabel = prometheus.Labels{"Method": "Register"}

func (c *RegisterCli) Register(service *RegisterArgv) (*RegisterResult, error) {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(registerLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(registerLabel).Inc()
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
		return nil, err
	}
	resp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}
	return &RegisterResult{InstanceID: instanceInfo.InstanceID}, nil
}

var deregisterLabel = prometheus.Labels{"Method": "Deregister"}

func (c *RegisterCli) Deregister(service *DeregisterArgv) error {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(deregisterLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(deregisterLabel).Inc()
	}()
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return err
	}
	_, err = client.Revoke(context.Background(), clientv3.LeaseID(service.InstanceID))
	if err != nil {
		return err
	}
	return nil
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

var addTargetRouterLabel = prometheus.Labels{"Method": "AddTargetRouter"}

func (c *RegisterCli) AddTargetRouter(argv *AddTargetRouterArgv) error {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(addTargetRouterLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(addTargetRouterLabel).Inc()
	}()
	key := util.RouterTargetInfoKey(argv.DstServiceName, argv.SrcInstanceID)
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return err
	}
	resp, err := client.Grant(context.Background(), argv.Timeout)
	if err != nil {
		return err
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
		return err
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	return nil
}

var addKVRouterLabel = prometheus.Labels{"Method": "AddKVRouter"}

func (c *RegisterCli) AddKVRouter(argv *AddKVRouterArgv) error {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(addKVRouterLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(addKVRouterLabel).Inc()
	}()
	key := util.RouterKVInfoKey(argv.DstServiceName, argv.Key)
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return err
	}
	resp, err := client.Grant(context.Background(), argv.Timeout)
	if err != nil {
		return err
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
		return err
	}
	_, err = client.Put(context.Background(), key, string(bytes), clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	return nil
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

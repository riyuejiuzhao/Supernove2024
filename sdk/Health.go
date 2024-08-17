package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/metrics"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type HealthCli struct {
	*APIContext
}

type HeartBeatArgv struct {
	ServiceName string
	InstanceID  int64
}

func (c *HealthCli) HeartBeat(argv *HeartBeatArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "HeartBeat"}, err)
	}()

	client, err := c.ConnManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		return
	}
	ch, err := client.KeepAlive(context.Background(), clientv3.LeaseID(argv.InstanceID))
	if err != nil {
		return
	}
	go func() {
		for _ = range ch {
		}
	}()
	return
}

func (c *HealthCli) HeartBeatOnce(argv *HeartBeatArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "HeartBeatOnce"}, err)
	}()

	client, err := c.ConnManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		return
	}
	_, err = client.KeepAliveOnce(context.Background(), clientv3.LeaseID(argv.InstanceID))
	if err != nil {
		return
	}
	return
}

type KeepRouterAliveArgv struct {
	RouterID int64
}

func (c *HealthCli) KeepRouterAlive(argv *KeepRouterAliveArgv) (err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "KeepRouterAlive"}, err)
	}()

	client, err := c.ConnManager.GetServiceConn(connMgr.InstancesEtcd, "")
	if err != nil {
		return
	}
	ch, err := client.KeepAlive(context.Background(), clientv3.LeaseID(argv.RouterID))
	if err != nil {
		return
	}
	go func() {
		for _ = range ch {
		}
	}()
	return
}

// HealthAPI 功能：
// 发送心跳
type HealthAPI interface {
	HeartBeatOnce(*HeartBeatArgv) error
	HeartBeat(argv *HeartBeatArgv) error
	KeepRouterAlive(argv *KeepRouterAliveArgv) error
}

func NewHealthAPIStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
	mt *metrics.MetricsManager,
) HealthAPI {
	ctx := NewAPIContextStandalone(config, conn, mt)
	return &HealthCli{
		APIContext: ctx,
	}
}

func NewHealthAPI() (HealthAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	return &HealthCli{ctx}, nil
}

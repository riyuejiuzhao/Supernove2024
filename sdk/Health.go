package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
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
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return
	}
	_, err = client.KeepAliveOnce(context.Background(), clientv3.LeaseID(argv.InstanceID))
	if err != nil {
		return
	}
	return
}

// HealthAPI 功能：
// 发送心跳
type HealthAPI interface {
	HeartBeat(*HeartBeatArgv) error
}

func NewHealthAPIStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
	dmgr dataMgr.ServiceDataManager,
	mt *metrics.MetricsManager,
) HealthAPI {
	ctx := NewAPIContextStandalone(config, conn, dmgr, mt)
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

package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type HealthCli struct {
	*APIContext
}

type HeartBeatArgv struct {
	ServiceName string
	InstanceID  int64
}

func (c *HealthCli) HeartBeat(argv *HeartBeatArgv) error {
	client, err := c.ConnManager.GetServiceConn(connMgr.Etcd)
	if err != nil {
		return err
	}
	ch, err := client.KeepAlive(context.Background(), clientv3.LeaseID(argv.InstanceID))
	if err != nil {
		return err
	}

	ka, ok := <-ch
	if !ok || ka == nil {
		return errors.New("实例已经超时")
	}
	return nil
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
) HealthAPI {
	ctx := NewAPIContextStandalone(config, conn, dmgr)
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

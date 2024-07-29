package sdk

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/connMgr"
	"context"
)

type HealthCli struct {
	APIContext
}

type HeartBeatArgv struct {
	ServiceName string
	InstanceID  string
}

func (c *HealthCli) HeartBeat(argv *HeartBeatArgv) error {
	conn, err := c.ConnManager.GetServiceConn(connMgr.HealthCheck)
	if err != nil {
		return err
	}
	defer conn.Close()
	rpcCli := miniRouterProto.NewHealthServiceClient(conn.Value())
	_, err = rpcCli.HeartBeat(context.Background(),
		&miniRouterProto.HeartBeatRequest{
			ServiceName: argv.ServiceName,
			InstanceID:  argv.InstanceID,
		})
	if err != nil {
		return err
	}
	return nil
}

// HealthAPI 功能：
// 发送心跳
type HealthAPI interface {
	HeartBeat(*HeartBeatArgv) error
}

func NewHealthAPI() (HealthAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	return &HealthCli{*ctx}, nil
}

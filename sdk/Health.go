package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"context"
)

type HealthCli struct {
	*APIContext
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
	rpcCli := pb.NewHealthServiceClient(conn.Value())
	_, err = rpcCli.HeartBeat(context.Background(),
		&pb.HeartBeatRequest{
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

func NewHealthAPIStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
) HealthAPI {
	ctx := NewAPIContextStandalone(config, conn)
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

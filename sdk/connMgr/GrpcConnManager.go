package connMgr

import (
	"Supernove2024/sdk"
	"github.com/shimingyah/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

const (
	// DialTimeout the timeout of create connection
	DialTimeout = 5 * time.Second

	// KeepAliveTime is the duration of time after which if the client doesn't see
	// any activity it pings the server to see if the transport is still alive.
	KeepAliveTime = time.Duration(10) * time.Second

	// KeepAliveTimeout is the duration of time for which the client waits after having
	// pinged for keepalive check and if no activity is seen even after that the connection
	// is closed.
	KeepAliveTimeout = time.Duration(3) * time.Second

	// InitialWindowSize we set it 1GB is to provide system's throughput.
	InitialWindowSize = 1 << 30

	// InitialConnWindowSize we set it 1GB is to provide system's throughput.
	InitialConnWindowSize = 1 << 30

	// MaxSendMsgSize set max gRPC request message size sent to server.
	// If any request message size is larger than current value, an error will be reported from gRPC.
	MaxSendMsgSize = 4 << 30

	// MaxRecvMsgSize set max gRPC receive message size received from server.
	// If any message size is larger than current value, an error will be reported from gRPC.
	MaxRecvMsgSize = 4 << 30

	// RetryPolicy 重传配置
	RetryPolicy = `{
        "methodConfig": [{
            "name": [
				{"service": "RegisterService"},
				{"service": "DiscoveryService"}
			],
            "retryPolicy": {
                "MaxAttempts": 4,
                "InitialBackoff": "0.1s",
                "MaxBackoff": "1s",
                "BackoffMultiplier": 2,
                "RetryableStatusCodes": [ "UNAVAILABLE" ]
            }
        }]
    }`
)

var DefaultOptions = pool.Options{
	Dial:                 Dial,
	MaxIdle:              8,
	MaxActive:            64,
	MaxConcurrentStreams: 64,
	Reuse:                true,
}

func Dial(address string) (*grpc.ClientConn, error) {
	return grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   120 * time.Second},
			MinConnectTimeout: DialTimeout}),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(RetryPolicy))
}

// GrpcConnManager 管理GRPC链接
type GrpcConnManager struct {
	poolDic map[string]pool.Pool
}

func ConnManagerInstance()

func newConnManager(services []sdk.ServiceConfig) *GrpcConnManager {
	poolDic := make(map[string]pool.Pool)
	for _, service := range services {
		address := service.String()
		p, err := pool.New(service.String(), DefaultOptions)
		if err != nil {
			sdk.Error("创建连接池失败，地址：%s", address)
			continue
		}
		sdk.Info("创建连接池，地址: %s", address)
		poolDic[address] = p
	}
	sdk.Info("初始化GrpcConnManager成功")
	return &GrpcConnManager{poolDic: poolDic}
}

func (m *GrpcConnManager) GetConn(address string) (pool.Conn, error) {
	p, ok := m.poolDic[address]
	if !ok {
		newPool, err := pool.New(address, DefaultOptions)
		if err != nil {
			return nil, err
		}
		sdk.Info("创建连接池 %s", address)
		m.poolDic[address] = newPool
		p = newPool
	}
	conn, err := p.Get()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

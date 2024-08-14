package grpc_sdk

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"errors"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type KVRouterOption struct {
	ServiceName string
	Key         string
}

type serverOptions struct {
	GrpcOption []grpc.ServerOption

	KVRouterOption []KVRouterOption
	Weight         *int32
	TTL            *int64
}

type ServerOption func(*serverOptions)

// WithKVRouter 标识哪些Key会被发送到本服务器
func WithKVRouter(routers []KVRouterOption) ServerOption {
	return func(options *serverOptions) {
		options.KVRouterOption = routers
	}
}

func WithWeight(weight int32) ServerOption {
	return func(options *serverOptions) {
		options.Weight = &weight
	}
}

func WithTTL(ttl int64) ServerOption {
	return func(options *serverOptions) {
		options.TTL = &ttl
	}
}

func WithGrpcOption(grpcOption []grpc.ServerOption) ServerOption {
	return func(options *serverOptions) {
		options.GrpcOption = grpcOption
	}
}

type Server struct {
	*grpc.Server

	GlobalConfig *config.Config
	DiscoveryAPI sdk.DiscoveryAPI
	RegisterAPI  sdk.RegisterAPI
	HealthAPI    sdk.HealthAPI

	options       *serverOptions
	InstanceIdDic map[string]int64
}

func NewServer(opts ...ServerOption) (srv *Server, err error) {
	cfg, err := config.GlobalConfig()
	if err != nil {
		return
	}
	dis, err := sdk.NewDiscoveryAPI()
	if err != nil {
		return
	}
	reg, err := sdk.NewRegisterAPI()
	if err != nil {
		return
	}
	health, err := sdk.NewHealthAPI()
	if err != nil {
		return
	}
	options := &serverOptions{}
	for _, opt := range opts {
		opt(options)
	}
	grpcSrv := grpc.NewServer(options.GrpcOption...)
	srv = &Server{
		Server: grpcSrv,

		GlobalConfig: cfg,
		DiscoveryAPI: dis,
		RegisterAPI:  reg,
		HealthAPI:    health,

		options: options,

		InstanceIdDic: make(map[string]int64),
	}

	return
}

func (srv *Server) doRegister(lis net.Listener) (err error) {
	serviceInfoDic := srv.Server.GetServiceInfo()
	var host string
	var port int32

	switch v := lis.Addr().(type) {
	case *net.TCPAddr:
		host = v.IP.String()
		port = int32(v.Port)
	case *net.UDPAddr:
		host = v.IP.String()
		port = int32(v.Port)
	default:
		return errors.New("不支持的地址类型")
	}

	for name := range serviceInfoDic {
		var registerResult *sdk.RegisterResult
		registerResult, err = srv.RegisterAPI.Register(&sdk.RegisterArgv{
			ServiceName: name,
			Host:        host,
			Port:        port,

			Weight: srv.options.Weight,
			TTL:    srv.options.TTL,
		})
		if err != nil {
			util.Error("register err: %v", err)
			continue
		}
		srv.InstanceIdDic[name] = registerResult.InstanceID
		err = srv.HealthAPI.HeartBeat(&sdk.HeartBeatArgv{
			ServiceName: name,
			InstanceID:  registerResult.InstanceID,
		})
		if err != nil {
			util.Error("register err: %v", err)
			continue
		}
	}
	return
}

func (srv *Server) routerRegister() {
	for _, router := range srv.options.KVRouterOption {
		instanceID, ok := srv.InstanceIdDic[router.ServiceName]
		if !ok {
			continue
		}
		result, err := srv.RegisterAPI.AddKVRouter(&sdk.AddKVRouterArgv{
			Key:            router.Key,
			DstServiceName: router.ServiceName,
			DstInstanceID:  instanceID,
		})
		if err != nil {
			util.Error("grpc注册路由失败：%v", err)
			continue
		}
		err = srv.HealthAPI.KeepRouterAlive(&sdk.KeepRouterAliveArgv{RouterID: result.RouterID})
		if err != nil {
			util.Error("路由保活失败: %v", err)
			continue
		}
	}

}

func (srv *Server) Serve(lis net.Listener) (err error) {
	err = srv.doRegister(lis)
	if err != nil {
		return err
	}
	srv.routerRegister()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGSEGV, syscall.SIGINT, syscall.SIGTERM)
		signal.Stop(c)
	}()

	return srv.Server.Serve(lis)
}

func (srv *Server) doDeregister() (err error) {
	for name := range srv.Server.GetServiceInfo() {
		instanceID, ok := srv.InstanceIdDic[name]
		if !ok {
			continue
		}
		err = srv.RegisterAPI.Deregister(&sdk.DeregisterArgv{
			ServiceName: name,
			InstanceID:  instanceID,
		})
		if err != nil {
			return
		}
		delete(srv.InstanceIdDic, name)
	}
	return
}

func (srv *Server) Stop() {
	err := srv.doDeregister()
	if err != nil {
		util.Error("stop err: %v", err)
	}
	srv.Server.Stop()
}

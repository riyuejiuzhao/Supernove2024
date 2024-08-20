package grpc_sdk

import (
	"Supernove2024/sdk"
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"context"
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

type InboundRouterOption struct {
	NextRouterType int32
	Tags           map[string]string
	InstancesName  []string
}

type serverOptions struct {
	GrpcOption []grpc.ServerOption

	InstanceName string
	Weight       *int32
	TTL          *int64

	Tables        []string
	InboundRouter []InboundRouterOption
}

type ServerOption func(*serverOptions)

func WithInstanceName(name string) ServerOption {
	return func(options *serverOptions) {
		options.InstanceName = name
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

func WithRouterTables(tables []string) ServerOption {
	return func(options *serverOptions) {
		options.Tables = tables
	}
}

func WithInboundRouters(inbounds []InboundRouterOption) ServerOption {
	return func(options *serverOptions) {
		options.InboundRouter = inbounds
	}
}

type Server struct {
	*grpc.Server

	GlobalConfig *config.Config
	RegisterAPI  sdk.RegisterAPI
	HealthAPI    sdk.HealthAPI

	options       *serverOptions
	InstanceIdDic map[string]int64
	cancelHeart   []context.CancelFunc
}

func NewServer(opts ...ServerOption) (srv *Server, err error) {
	cfg, err := config.GlobalConfig()
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

	if options.InstanceName == "" {
		util.Warn("缺少Instance Name，将会随机生成")
		options.InstanceName = util.GenerateRandomString(10)
	}

	grpcSrv := grpc.NewServer(options.GrpcOption...)
	srv = &Server{
		Server: grpcSrv,

		GlobalConfig: cfg,
		RegisterAPI:  reg,
		HealthAPI:    health,

		options:       options,
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
		//注册路由表
		if srv.options.Tables != nil &&
			len(srv.options.Tables) != 0 {
			err = srv.RegisterAPI.AddTable(&sdk.AddTableArgv{
				ServiceName: name,
				Tags:        srv.options.Tables,
			})
			if err != nil {
				util.Error("Grpc服务器%s 添加路由表失败 err:%v", name, err)
			}
		}
		//注册服务
		registerResult, err = srv.RegisterAPI.Register(&sdk.RegisterArgv{
			ServiceName: name,
			Host:        host,
			Port:        port,
			Name:        srv.options.InstanceName,

			Weight: srv.options.Weight,
			TTL:    srv.options.TTL,
		})
		if err != nil {
			util.Error("register err: %v", err)
			continue
		}
		srv.InstanceIdDic[name] = registerResult.InstanceID
		//添加心跳
		cancel, err := srv.HealthAPI.HeartBeat(&sdk.HeartBeatArgv{
			ServiceName: name,
			InstanceID:  registerResult.InstanceID,
		})
		srv.cancelHeart = append(srv.cancelHeart, cancel)
		if err != nil {
			util.Error("register heart beat err: %v", err)
			continue
		}
		// 注册路由
		for _, router := range srv.options.InboundRouter {
			result, err := srv.RegisterAPI.AddKVRouter(&sdk.AddKVRouterArgv{
				Dic:             router.Tags,
				DstServiceName:  name,
				DstInstanceName: router.InstancesName,
				NextRouterType:  router.NextRouterType,
			})
			if err != nil {
				util.Error("注册路由失败 err:%v", err)
				continue
			}
			cancel, err := srv.HealthAPI.KeepRouterAlive(&sdk.KeepRouterAliveArgv{RouterID: result.RouterID})
			srv.cancelHeart = append(srv.cancelHeart, cancel)
			if err != nil {
				util.Error("路由保活失败")
				continue
			}
		}
	}
	return
}

func (srv *Server) Serve(lis net.Listener) (err error) {
	err = srv.doRegister(lis)
	if err != nil {
		return err
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGSEGV, syscall.SIGINT, syscall.SIGTERM)
		signal.Stop(c)
	}()

	return srv.Server.Serve(lis)
}

// 不删除已经注册的路由了
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
	for _, cancel := range srv.cancelHeart {
		cancel()
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

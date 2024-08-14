package grpc_sdk

import (
	"Supernove2024/sdk"
	"Supernove2024/util"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// MiniRouterResolver 通过负载均衡器实现路由
type MiniRouterResolver struct {
	discoveryAPI sdk.DiscoveryAPI

	target resolver.Target
	cc     resolver.ClientConn
}

func (r *MiniRouterResolver) start() {
	service := r.target.URL.Host
	ch, err := r.discoveryAPI.WatchService(&sdk.WatchServiceArgv{ServiceName: service})
	if err != nil {
		util.Error("resolver启动失败, err: %v", err)
		return
	}
	go func() {
		for msg := range ch {
			err := r.doResolveOnce(service, msg.Instances)
			if err != nil {
				util.Error("resolver更新地址失败, err: %v", err)
			}
			return
		}
	}()
}

func (r *MiniRouterResolver) doResolveOnce(service string, instances []util.DstInstanceInfo) (err error) {
	addresses := util.SliceMap(instances, func(t util.DstInstanceInfo) resolver.Address {
		return resolver.Address{
			Addr:       util.Address(t.GetHost(), t.GetPort()),
			Attributes: attributes.New("Info", t),
		}
	})
	err = r.cc.UpdateState(
		resolver.State{
			Addresses:  addresses,
			Attributes: attributes.New("Name", service),
		})
	return
}

func (r *MiniRouterResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	service := r.target.URL.Host
	result, err := r.discoveryAPI.GetInstances(&sdk.GetInstancesArgv{ServiceName: service})
	if err != nil {
		util.Error("更新RPC地址失败, err: %v", err)
		return
	}
	err = r.doResolveOnce(service, result.Service.Instances)
	if err != nil {
		util.Error("更新RPC地址失败, err: %v", err)
		return
	}
}

func (*MiniRouterResolver) Close() {}

type miniRouterResolverBuilder struct{}

func (*miniRouterResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	discoveryAPI, err := sdk.NewDiscoveryAPI()
	if err != nil {
		return nil, err
	}
	r := &MiniRouterResolver{
		discoveryAPI: discoveryAPI,
		target:       target,
		cc:           cc,
	}
	r.start()
	return r, nil
}

func (*miniRouterResolverBuilder) Scheme() string {
	return Scheme
}

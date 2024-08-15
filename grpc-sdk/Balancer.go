package grpc_sdk

import (
	"Supernove2024/sdk"
	"Supernove2024/util"
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

type miniRouterBalancerBuilder struct {
}

func (bb *miniRouterBalancerBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	discoveryAPI, err := sdk.NewDiscoveryAPI()

	if err != nil {
		util.Error("balancer获取discoveryAPI出现错误 err: %v", err)
	}

	return &MiniRouterBalancer{
		discoveryAPI: discoveryAPI,
		cc:           cc,
	}
}

func (bb *miniRouterBalancerBuilder) Name() string {
	return Scheme
}

type BalancerInstanceInfo struct {
	SubConn  balancer.SubConn
	Instance util.DstInstanceInfo
}

func (r *BalancerInstanceInfo) GetName() string {
	return r.Instance.GetName()
}
func (r *BalancerInstanceInfo) GetInstanceID() int64 {
	return r.Instance.GetInstanceID()
}
func (r *BalancerInstanceInfo) GetWeight() int32 {
	return r.Instance.GetWeight()
}
func (r *BalancerInstanceInfo) GetHost() string {
	return r.Instance.GetHost()
}
func (r *BalancerInstanceInfo) GetPort() int32 {
	return r.Instance.GetPort()
}

type MiniRouterBalancer struct {
	discoveryAPI sdk.DiscoveryAPI
	cc           balancer.ClientConn

	nowPicker util.SyncContainer[*MiniRouterPicker]
}

func (b *MiniRouterBalancer) Build(cc balancer.ClientConn, _ balancer.BuildOptions) balancer.Balancer {
	dis, err := sdk.NewDiscoveryAPI()
	if err != nil {
		util.Error("Balancer创建Discovery错误 err: %v", err)
	}
	return &MiniRouterBalancer{
		discoveryAPI: dis,
		cc:           cc,
	}
}

func (b *MiniRouterBalancer) ResolverError(err error) {
	util.Error("Resolver error: %v\n", err)
}

func (b *MiniRouterBalancer) UpdateClientConnState(s balancer.ClientConnState) error {
	serviceName := s.ResolverState.Attributes.Value("Name")
	routerInfos := make([]util.DstInstanceInfo, 0, len(s.ResolverState.Addresses))
	states := make(map[balancer.SubConn]connectivity.State)

	for _, addr := range s.ResolverState.Addresses {
		instanceInfo := addr.Attributes.Value("Info").(util.DstInstanceInfo)
		sc, err := b.cc.NewSubConn([]resolver.Address{addr}, balancer.NewSubConnOptions{})
		if err != nil {
			return err
		}
		routerInfos = append(routerInfos, &BalancerInstanceInfo{
			SubConn:  sc,
			Instance: instanceInfo,
		})
		states[sc] = connectivity.Idle
		sc.Connect()
	}

	picker := &MiniRouterPicker{
		discoveryAPI: b.discoveryAPI,
		serviceInfo: &util.ServiceInfo{
			Name:      serviceName.(string),
			Instances: routerInfos,
		},
		states: states,
	}

	func() {
		b.nowPicker.Mutex.Lock()
		defer b.nowPicker.Mutex.Unlock()
		b.nowPicker.Value = picker
	}()

	b.cc.UpdateState(balancer.State{ConnectivityState: connectivity.Ready, Picker: picker})
	return nil
}

func (b *MiniRouterBalancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
	b.nowPicker.Mutex.Lock()
	defer b.nowPicker.Mutex.Unlock()
	b.nowPicker.Value.states[sc] = state.ConnectivityState
	b.cc.UpdateState(balancer.State{ConnectivityState: connectivity.Ready, Picker: b.nowPicker.Value})
	return
}

func (b *MiniRouterBalancer) Close() {}

type MiniRouterPicker struct {
	discoveryAPI sdk.DiscoveryAPI
	serviceInfo  util.DstService

	states map[balancer.SubConn]connectivity.State
}

func (p *MiniRouterPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	md, ok := metadata.FromOutgoingContext(info.Ctx)
	if !ok {
		return balancer.PickResult{}, fmt.Errorf("缺少metadata")
	}
	routerTypes, ok := md[RouterTypeHeader]
	if !ok {
		bInfo := util.RandomItem(p.serviceInfo.GetInstance()).(*BalancerInstanceInfo)
		return balancer.PickResult{SubConn: bInfo.SubConn}, nil
	}
	if len(routerTypes) > 1 {
		util.Warn("路由方法指定了多次，将使用首次指定的结果")
	}
	argv := &sdk.ProcessRouterArgv{
		Method:          0,
		SrcInstanceName: "",
		DstService:      p.serviceInfo,
		Key:             "",
	}
	switch routerTypes[0] {
	case ConsistentRouterType:
		argv.Method = util.ConsistentRouterType
	case RandomRouterType:
		argv.Method = util.RandomRouterType
	case WeightedRouterType:
		argv.Method = util.WeightedRouterType
	case TargetRouterType:
		argv.Method = util.TargetRouterType
	case KVRouterType:
		argv.Method = util.KVRouterType
		routerKey, ok := md[RouterKeyHeader]
		if !ok {
			bInfo := util.RandomItem(p.serviceInfo.GetInstance()).(*BalancerInstanceInfo)
			return balancer.PickResult{SubConn: bInfo.SubConn}, fmt.Errorf("报文头缺少Key")
		}
		if len(routerKey) > 1 {
			util.Warn("Key指定了多个，将使用首个")
		}
		argv.Key = routerKey[0]
	}
	result, err := p.discoveryAPI.ProcessRouter(argv)
	if err != nil {
		return balancer.PickResult{}, err
	}
	bInfo := result.DstInstance.(*BalancerInstanceInfo)
	if p.states[bInfo.SubConn] != connectivity.Ready {
		return balancer.PickResult{}, fmt.Errorf("对映实例尚未连接完成")
	}
	return balancer.PickResult{SubConn: bInfo.SubConn}, nil
}

package grpc_sdk

import (
	"Supernove2024/sdk"
	"Supernove2024/util"
	"encoding/json"
	"fmt"
	"github.com/sony/gobreaker"
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
	serviceName := s.ResolverState.Attributes.Value("Name").(string)
	states := make(map[balancer.SubConn]connectivity.State)
	infos := make(map[int64]balancer.SubConn)

	for _, addr := range s.ResolverState.Addresses {
		instanceInfo := addr.Attributes.Value("Info").(util.DstInstanceInfo)
		sc, err := b.cc.NewSubConn([]resolver.Address{addr}, balancer.NewSubConnOptions{})
		if err != nil {
			return err
		}
		states[sc] = connectivity.Idle
		infos[instanceInfo.GetInstanceID()] = sc
		sc.Connect()
	}

	picker := &MiniRouterPicker{
		discoveryAPI: b.discoveryAPI,
		serviceName:  serviceName,
		states:       states,
		infos:        infos,
		breakers:     make(map[int64]*gobreaker.CircuitBreaker),
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
	serviceName  string //util.DstService

	breakers map[int64]*gobreaker.CircuitBreaker
	infos    map[int64]balancer.SubConn
	states   map[balancer.SubConn]connectivity.State
}

func (p *MiniRouterPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	md, ok := metadata.FromOutgoingContext(info.Ctx)
	if !ok {
		return balancer.PickResult{}, fmt.Errorf("缺少metadata")
	}
	routerTypes, ok := md[RouterTypeHeader]
	if !ok {
		bInfo := util.RandomDicValue(p.infos)
		return balancer.PickResult{SubConn: bInfo}, nil
	}
	if len(routerTypes) > 1 {
		util.Warn("路由方法指定了多次，将使用首次指定的结果")
	}
	argv := &sdk.ProcessRouterArgv{
		Method:          0,
		SrcInstanceName: "",
		DstService:      p.serviceName,
		Key:             make(map[string]string),
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
			return balancer.PickResult{}, fmt.Errorf("报文头缺少Key")
		}
		if len(routerKey) > 1 {
			util.Warn("Key指定了多个，将使用首个")
		}
		var dic map[string]string
		err := json.Unmarshal([]byte(routerKey[0]), &dic)
		if err != nil {
			return balancer.PickResult{}, fmt.Errorf("报文头缺少Key")
		}
		argv.Key = dic
	}
	result, err := p.discoveryAPI.ProcessRouter(argv)
	if err != nil {
		return balancer.PickResult{}, err
	}
	bInfo := p.infos[result.DstInstance.GetInstanceID()] //result.SrcInstance.(*BalancerInstanceInfo)
	if p.states[bInfo] != connectivity.Ready {
		return balancer.PickResult{}, fmt.Errorf("对映实例尚未连接完成")
	}
	return balancer.PickResult{SubConn: bInfo, Done: func(doneInfo balancer.DoneInfo) {
		if doneInfo.Err != nil {
			result.DstInstance.GetBreaker().Fail()
		} else {
			result.DstInstance.GetBreaker().Success()
		}
	}}, nil
}

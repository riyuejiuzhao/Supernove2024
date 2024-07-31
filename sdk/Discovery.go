package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/util"
	"errors"
	"math/rand"
	"stathat.com/c/consistent"
)

const (
	// Consistent 一致性哈希
	Consistent = int32(iota)
	// Random 随机路由
	Random
	// Weighted 基于权重
	Weighted
	// Target 特定路由
	Target
	// KeyValue 键值对路由
	KeyValue
)

type GetInstancesResult struct {
	serviceInfo *pb.ServiceInfo
}

func (r *GetInstancesResult) GetServiceName() string {
	return r.serviceInfo.ServiceName
}

func (r *GetInstancesResult) GetInstance() []*pb.InstanceInfo {
	return r.serviceInfo.Instances
}

type DiscoveryCli struct {
	APIContext
	dataMgr dataMgr.ServiceDataManager
}

func (c *DiscoveryCli) GetInstances(serviceName string) (*GetInstancesResult, error) {
	//在缓存数据中查找
	service, ok := c.dataMgr.GetServiceInstance(serviceName)
	if !ok {
		return nil, errors.New("不存在该服务")
	}
	return &GetInstancesResult{serviceInfo: service}, nil
}

func (c *DiscoveryCli) processConsistentRouter(srcInstanceID string, dstInstances []*pb.InstanceInfo) (*ProcessRouterResult, error) {
	hash := consistent.New()
	dict := make(map[string]*pb.InstanceInfo)
	for _, v := range dstInstances {
		dict[v.InstanceID] = v
		hash.Add(v.InstanceID)
	}
	dstInstanceID, err := hash.Get(srcInstanceID)
	if err != nil {
		return nil, err
	}
	return &ProcessRouterResult{DstInstance: dict[dstInstanceID]}, err
}

func (c *DiscoveryCli) processRandomRouter(dstInstances []*pb.InstanceInfo) (*ProcessRouterResult, error) {
	instance := util.RandomItem(dstInstances)
	return &ProcessRouterResult{DstInstance: instance}, nil
}

func (c *DiscoveryCli) processWeightRouter(dstInstances []*pb.InstanceInfo) (*ProcessRouterResult, error) {
	totalWeight := int32(0)
	for _, v := range dstInstances {
		totalWeight += v.Weight
	}
	targetWeight := rand.Int31n(totalWeight)
	for _, v := range dstInstances {
		if targetWeight <= v.Weight {
			return &ProcessRouterResult{DstInstance: v}, nil
		}
		targetWeight -= v.Weight
	}
	return nil, errors.New("权重超过上限")
}

func (c *DiscoveryCli) processTargetRouter(srcInstanceID string, dstService string) (*ProcessRouterResult, error) {
}

func (c *DiscoveryCli) processKeyValueRouter(srcInstanceID string, dstService string) (*ProcessRouterResult, error) {
}

// ProcessRouter 如果没有提供可选的Instance，那么自动从缓冲中获取
func (c *DiscoveryCli) ProcessRouter(argv *ProcessRouterArgv) (*ProcessRouterResult, error) {
	instances := argv.DstService.GetInstance()
	if instances != nil || len(instances) == 0 {
		service, ok := c.dataMgr.GetServiceInstance(argv.DstService.GetServiceName())
		if !ok {
			return nil, errors.New("不存在目标服务")
		}
		instances = service.Instances
	}

	switch argv.Method {
	case Target:
		return c.processTargetRouter(argv.SrcInstanceID, argv.DstService.GetServiceName())
	case KeyValue:
		//TODO
		//return c.processKeyValueRouter(argv)
	case Consistent:
		return c.processConsistentRouter(argv.SrcInstanceID, instances)
	case Random:
		return c.processRandomRouter(instances)
	case Weighted:
		return c.processWeightRouter(instances)
	}
	return nil, errors.New("不支持的路由算法类型")
}

func (c *DiscoveryCli) AddTargetRouter(*AddTargetRouterArgv) (*AddTargetRouterResult, error) {

}

type DefaultDstService struct {
	serviceName string
	instances   []*pb.InstanceInfo
}

func (d *DefaultDstService) GetServiceName() string {
	return d.serviceName
}
func (d *DefaultDstService) GetInstance() []*pb.InstanceInfo {
	return d.instances
}

type DstService interface {
	GetServiceName() string
	GetInstance() []*pb.InstanceInfo
}

type ProcessRouterArgv struct {
	Method        int32
	SrcInstanceID string
	DstService    DstService
}

type ProcessRouterResult struct {
	DstInstance *pb.InstanceInfo
}

type AddTargetRouterArgv struct {
}

type AddTargetRouterResult struct {
}

type DiscoveryAPI interface {
	GetInstances(serviceName string) (*GetInstancesResult, error)
	ProcessRouter(*ProcessRouterArgv) (*ProcessRouterResult, error)
	AddTargetRouter(*AddTargetRouterArgv) (*AddTargetRouterResult, error)
}

func NewDiscoveryAPI() (DiscoveryAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	dataManager, err := dataMgr.Instance()
	if err != nil {
		return nil, err
	}
	return &DiscoveryCli{*ctx, dataManager}, nil
}

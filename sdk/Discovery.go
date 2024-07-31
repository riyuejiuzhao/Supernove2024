package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/util"
	"errors"
	"math/rand"
	"stathat.com/c/consistent"
	"time"
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

type GetInstancesArgv struct {
	ServiceName      string
	SkipHealthFilter bool
}

type GetInstancesResult struct {
	ServiceName string
	Instances   []*pb.InstanceInfo
}

func (r *GetInstancesResult) GetServiceName() string {
	return r.ServiceName
}

func (r *GetInstancesResult) GetInstance() []*pb.InstanceInfo {
	return r.Instances
}

type DiscoveryCli struct {
	APIContext
	dataMgr dataMgr.ServiceDataManager
}

func (c *DiscoveryCli) GetInstances(argv *GetInstancesArgv) (*GetInstancesResult, error) {
	//在缓存数据中查找
	service, ok := c.dataMgr.GetServiceInfo(argv.ServiceName)
	nowTime := time.Now().Unix()
	result := &GetInstancesResult{ServiceName: argv.ServiceName,
		Instances: make([]*pb.InstanceInfo, 0, len(service.Instances))}
	//健康信息过滤
	if argv.SkipHealthFilter {
		result.Instances = service.Instances
	} else {
		for _, instance := range service.Instances {
			healthInfo, ok := c.dataMgr.GetHealthInfo(argv.ServiceName, instance.InstanceID)
			if !ok {
				continue
			}
			if healthInfo.LastHeartBeat+3*healthInfo.TTL < nowTime {
				continue
			}
			result.Instances = append(result.Instances, instance)
		}
	}

	if !ok {
		return nil, errors.New("不存在该服务")
	}
	return result, nil
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
	target, ok := c.dataMgr.GetTargetRouter(dstService, srcInstanceID)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	instance, ok := c.dataMgr.GetInstanceInfo(dstService, target.DstInstanceID)
	if !ok {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: instance}, nil
}

func (c *DiscoveryCli) processKeyValueRouter(key string, dstService string) (*ProcessRouterResult, error) {
	v, ok := c.dataMgr.GetKVRouter(dstService, key)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	instance, ok := c.dataMgr.GetInstanceInfo(dstService, v.DstInstanceID)
	if !ok {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: instance}, nil
}

// ProcessRouter 如果没有提供可选的Instance，那么自动从缓冲中获取
func (c *DiscoveryCli) ProcessRouter(argv *ProcessRouterArgv) (*ProcessRouterResult, error) {
	instances := argv.DstService.GetInstance()
	if instances != nil || len(instances) == 0 {
		service, ok := c.dataMgr.GetServiceInfo(argv.DstService.GetServiceName())
		if !ok {
			return nil, errors.New("不存在目标服务")
		}
		instances = service.Instances
	}

	switch argv.Method {
	case Target:
		return c.processTargetRouter(argv.SrcInstanceID, argv.DstService.GetServiceName())
	case KeyValue:
		if argv.Key == nil {
			return nil, errors.New("使用键值对路由但是缺少Key")
		}
		return c.processKeyValueRouter(*argv.Key, argv.DstService.GetServiceName())
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
	return nil, errors.New("未实现")
}

func (c *DiscoveryCli) AddKVRouter(*AddTargetRouterArgv) (*AddTargetRouterResult, error) {
	return nil, errors.New("未实现")
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

	Key *string
}

type ProcessRouterResult struct {
	DstInstance *pb.InstanceInfo
}

type AddTargetRouterArgv struct {
}

type AddTargetRouterResult struct {
}

type DiscoveryAPI interface {
	GetInstances(argv *GetInstancesArgv) (*GetInstancesResult, error)
	ProcessRouter(*ProcessRouterArgv) (*ProcessRouterResult, error)
	AddTargetRouter(*AddTargetRouterArgv) (*AddTargetRouterResult, error)
	AddKVRouter(*AddTargetRouterArgv) (*AddTargetRouterResult, error)
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

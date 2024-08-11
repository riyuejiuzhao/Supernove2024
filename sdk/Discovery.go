package sdk

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"stathat.com/c/consistent"
	"strconv"
	"time"
)

type GetInstancesArgv struct {
	ServiceName string
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
	*APIContext
}

var getInstancesLabel = prometheus.Labels{"Method": "GetInstances"}

func (c *DiscoveryCli) GetInstances(argv *GetInstancesArgv) (*GetInstancesResult, error) {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(getInstancesLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(getInstancesLabel).Inc()
	}()
	//在缓存数据中查找
	service, ok := c.DataMgr.GetServiceInfo(argv.ServiceName)
	if !ok {
		return nil, errors.New("不存在该服务")
	}
	result := &GetInstancesResult{ServiceName: argv.ServiceName,
		Instances: service.Instances}
	return result, nil
}

func (c *DiscoveryCli) processConsistentRouter(srcInstanceID int64, dstInstances []*pb.InstanceInfo) (*ProcessRouterResult, error) {
	hash := consistent.New()
	dict := make(map[string]*pb.InstanceInfo)
	for _, v := range dstInstances {
		address := util.Address(v.Host, v.Port)
		dict[address] = v
		hash.Add(address)
	}
	dstInstanceID, err := hash.Get(strconv.FormatInt(srcInstanceID, 10))
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
	maxWeight := dstInstances[0]
	for _, v := range dstInstances {
		if v.Weight > maxWeight.Weight {
			maxWeight = v
		}
	}
	return &ProcessRouterResult{DstInstance: maxWeight}, nil
}

func (c *DiscoveryCli) processTargetRouter(srcInstanceID int64, dstService string) (*ProcessRouterResult, error) {
	target, ok := c.DataMgr.GetTargetRouter(dstService, srcInstanceID)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	instance, ok := c.DataMgr.GetInstanceInfo(dstService, target.DstInstanceID)
	if !ok {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: instance}, nil
}

func (c *DiscoveryCli) processKeyValueRouter(key string, dstService string) (*ProcessRouterResult, error) {
	v, ok := c.DataMgr.GetKVRouter(dstService, key)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	instance, ok := c.DataMgr.GetInstanceInfo(dstService, v.DstInstanceID)
	if !ok {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: instance}, nil
}

var processRouterLabel = prometheus.Labels{"Method": "ProcessRouter"}

// ProcessRouter 如果没有提供可选的Instance，那么自动从缓冲中获取
func (c *DiscoveryCli) ProcessRouter(argv *ProcessRouterArgv) (*ProcessRouterResult, error) {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin).Milliseconds()
		c.Metrics.MethodTime.With(processRouterLabel).Observe(float64(duration))
		c.Metrics.MethodCounter.With(processRouterLabel).Inc()
	}()
	instances := argv.DstService.GetInstance()
	if instances == nil || len(instances) == 0 {
		service, ok := c.DataMgr.GetServiceInfo(argv.DstService.GetServiceName())
		if !ok {
			return nil, errors.New("不存在目标服务")
		}
		instances = service.Instances
	}

	switch argv.Method {
	case util.TargetRouterType:
		return c.processTargetRouter(argv.SrcInstanceID, argv.DstService.GetServiceName())
	case util.KVRouterType:
		if argv.Key == "" {
			return nil, errors.New("使用键值对路由但是缺少Key")
		}
		return c.processKeyValueRouter(argv.Key, argv.DstService.GetServiceName())
	case util.ConsistentRouterType:
		return c.processConsistentRouter(argv.SrcInstanceID, instances)
	case util.RandomRouterType:
		return c.processRandomRouter(instances)
	case util.WeightedRouterType:
		return c.processWeightRouter(instances)
	}
	return nil, errors.New("不支持的路由算法类型")
}

type DefaultDstService struct {
	ServiceName string
	Instances   []*pb.InstanceInfo
}

func (d *DefaultDstService) GetServiceName() string {
	return d.ServiceName
}
func (d *DefaultDstService) GetInstance() []*pb.InstanceInfo {
	return d.Instances
}

type DstService interface {
	GetServiceName() string
	GetInstance() []*pb.InstanceInfo
}

type ProcessRouterArgv struct {
	Method        int32
	SrcInstanceID int64
	DstService    DstService
	//可选的，只有kv需要
	Key string
}

type ProcessRouterResult struct {
	DstInstance *pb.InstanceInfo
}

type DiscoveryAPI interface {
	GetInstances(argv *GetInstancesArgv) (*GetInstancesResult, error)
	ProcessRouter(*ProcessRouterArgv) (*ProcessRouterResult, error)
}

func NewDiscoveryAPIStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
	dmgr dataMgr.ServiceDataManager,
	mt *metrics.MetricsManager,
) DiscoveryAPI {
	ctx := NewAPIContextStandalone(config, conn, dmgr, mt)
	return &DiscoveryCli{
		APIContext: ctx,
	}
}

func NewDiscoveryAPI() (DiscoveryAPI, error) {
	ctx, err := NewAPIContext()
	if err != nil {
		return nil, err
	}
	return &DiscoveryCli{ctx}, nil
}

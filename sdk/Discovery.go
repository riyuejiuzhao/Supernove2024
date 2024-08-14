package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"stathat.com/c/consistent"
	"strconv"
	"time"
)

type GetInstancesArgv struct {
	ServiceName string
}

type GetInstancesResult struct {
	Service *util.ServiceInfo
}

func (r *GetInstancesResult) GetServiceName() string {
	return r.Service.Name
}

func (r *GetInstancesResult) GetInstance() []util.DstInstanceInfo {
	return r.Service.Instances
}

type DiscoveryCli struct {
	*APIContext
	DataMgr dataMgr.ServiceDataManager
}

func (c *DiscoveryCli) GetInstances(argv *GetInstancesArgv) (result *GetInstancesResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "GetInstances"}, err)
	}()
	//在缓存数据中查找
	service, ok := c.DataMgr.GetServiceInfo(argv.ServiceName)
	if !ok {
		result = nil
		err = errors.New("不存在该服务")
		return
	}
	result = &GetInstancesResult{
		Service: &util.ServiceInfo{
			Name:      argv.ServiceName,
			Instances: service.Instances,
		},
	}
	return
}

func (c *DiscoveryCli) processConsistentRouter(srcInstanceID int64, dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	hash := consistent.New()
	dict := make(map[string]util.DstInstanceInfo)
	for _, v := range dstInstances {
		address := util.Address(v.GetHost(), v.GetPort())
		dict[address] = v
		hash.Add(address)
	}
	dstInstanceID, err := hash.Get(strconv.FormatInt(srcInstanceID, 10))
	if err != nil {
		return nil, err
	}
	return &ProcessRouterResult{DstInstance: dict[dstInstanceID]}, err
}

func (c *DiscoveryCli) processRandomRouter(dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	instance := util.RandomItem(dstInstances)
	return &ProcessRouterResult{DstInstance: instance}, nil
}

func (c *DiscoveryCli) processWeightRouter(dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	maxWeight := dstInstances[0]
	for _, v := range dstInstances {
		if v.GetWeight() > maxWeight.GetWeight() {
			maxWeight = v
		}
	}
	return &ProcessRouterResult{DstInstance: maxWeight}, nil
}

func (c *DiscoveryCli) processTargetRouter(srcInstanceName string, dstService string, instance []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	target, ok := c.DataMgr.GetTargetRouter(dstService, srcInstanceName)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	for _, i := range instance {
		if i.GetName() != target.GetDstInstanceName() {
			continue
		}
		return &ProcessRouterResult{DstInstance: i}, nil
	}
	return nil, errors.New("没有目标实例")
}

func (c *DiscoveryCli) processKeyValueRouter(key string, dstService string, instances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	v, ok := c.DataMgr.GetKVRouter(dstService, key)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	for _, i := range instances {
		if i.GetName() != v.DstInstanceName {
			continue
		}
		return &ProcessRouterResult{DstInstance: i}, nil
	}
	return nil, errors.New("没有目标实例")
}

// ProcessRouter 如果没有提供可选的Instance，那么自动从缓冲中获取
func (c *DiscoveryCli) ProcessRouter(argv *ProcessRouterArgv) (result *ProcessRouterResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "ProcessRouter"}, err)
	}()
	instances := argv.DstService.GetInstance()
	if instances == nil || len(instances) == 0 {
		service, ok := c.DataMgr.GetServiceInfo(argv.DstService.GetServiceName())
		if !ok {
			err = errors.New(fmt.Sprintf(
				"不存在目标服务 %s",
				argv.DstService.GetServiceName()))
			return
		}
		instances = service.Instances
	}

	switch argv.Method {
	case util.TargetRouterType:
		return c.processTargetRouter(argv.SrcInstanceName, argv.DstService.GetServiceName(), instances)
	case util.KVRouterType:
		if argv.Key == "" {
			return nil, errors.New("使用键值对路由但是缺少Key")
		}
		return c.processKeyValueRouter(argv.Key, argv.DstService.GetServiceName(), instances)
	case util.ConsistentRouterType:
		return c.processConsistentRouter(argv.SrcInstanceName, instances)
	case util.RandomRouterType:
		return c.processRandomRouter(instances)
	case util.WeightedRouterType:
		return c.processWeightRouter(instances)
	}
	return nil, errors.New("不支持的路由算法类型")
}

func (c *DiscoveryCli) WatchService(argv *WatchServiceArgv) (<-chan *util.ServiceInfo, error) {
	return c.DataMgr.WatchServiceInfo(argv.ServiceName)
}

type DefaultDstService struct {
	ServiceName string
	Instances   []util.DstInstanceInfo
}

func (d *DefaultDstService) GetServiceName() string {
	return d.ServiceName
}
func (d *DefaultDstService) GetInstance() []util.DstInstanceInfo {
	return d.Instances
}

type ProcessRouterArgv struct {
	Method          int32
	SrcInstanceName string

	DstService util.DstService
	//可选的，只有kv需要
	Key string
}

type ProcessRouterResult struct {
	DstInstance util.DstInstanceInfo
}

type WatchServiceArgv struct {
	ServiceName string
}

type DiscoveryAPI interface {
	GetInstances(argv *GetInstancesArgv) (*GetInstancesResult, error)
	ProcessRouter(*ProcessRouterArgv) (*ProcessRouterResult, error)
	WatchService(argv *WatchServiceArgv) (<-chan *util.ServiceInfo, error)
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
	mgr, err := dataMgr.Instance()
	if err != nil {
		return nil, err
	}

	return &DiscoveryCli{ctx, mgr}, nil
}

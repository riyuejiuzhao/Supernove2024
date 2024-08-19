package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/dgryski/go-jump"
	"github.com/prometheus/client_golang/prometheus"
	"hash/fnv"
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
		err = fmt.Errorf("不存在该服务%s", argv.ServiceName)
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

func (c *DiscoveryCli) doProcessConsistentRouter(
	srcInstanceName string,
	dstInstances []util.DstInstanceInfo,
) (*ProcessRouterResult, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(srcInstanceName))
	if err != nil {
		return nil, err
	}
	keyHash := h.Sum64()
	bucket := jump.Hash(keyHash, len(dstInstances))
	return &ProcessRouterResult{DstInstance: dstInstances[bucket]}, err
}

func (c *DiscoveryCli) processConsistentRouter(
	srcInstanceName string,
	dstService string,
) (*ProcessRouterResult, error) {
	instances, ok := c.DataMgr.GetServiceInfo(dstService)
	if !ok || len(instances.Instances) == 0 {
		return nil, errors.New("没有对应实例")
	}
	readyInstance := make([]util.DstInstanceInfo, 0, len(instances.Instances))
	for _, i := range instances.Instances {
		breakInfo, ok := i.(*dataMgr.InstanceInfo)
		if !ok || !breakInfo.Breaker.Ready() {
			continue
		}
		readyInstance = append(readyInstance, i)
	}
	return c.doProcessConsistentRouter(srcInstanceName, readyInstance)
}

func (c *DiscoveryCli) doProcessRandomRouter(dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	instance := util.RandomItem(dstInstances)
	return &ProcessRouterResult{DstInstance: instance}, nil
}

func (c *DiscoveryCli) processRandomRouter(dstService string) (*ProcessRouterResult, error) {
	serviceInfo, ok := c.DataMgr.GetServiceInfo(dstService)
	if !ok {
		return nil, fmt.Errorf("不存在服务%s", dstService)
	}
	readyInstance := make([]util.DstInstanceInfo, 0, len(serviceInfo.Instances))
	for _, i := range serviceInfo.Instances {
		breakInfo, ok := i.(*dataMgr.InstanceInfo)
		if !ok || !breakInfo.Breaker.Ready() {
			continue
		}
		readyInstance = append(readyInstance, i)
	}
	return c.doProcessRandomRouter(readyInstance)
}

func (c *DiscoveryCli) doProcessWeightRouter(dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	v := util.RandomWeightItem(dstInstances)
	return &ProcessRouterResult{DstInstance: v}, nil
}

func (c *DiscoveryCli) processWeightRouter(dstInstances string) (*ProcessRouterResult, error) {
	service, ok := c.DataMgr.GetServiceInfo(dstInstances)
	if !ok {
		return nil, fmt.Errorf("没有服务%s", dstInstances)
	}
	readyInstance := make([]util.DstInstanceInfo, 0, len(service.Instances))
	for _, i := range service.Instances {
		breakInfo, ok := i.(*dataMgr.InstanceInfo)
		if !ok || !breakInfo.Breaker.Ready() {
			continue
		}
		readyInstance = append(readyInstance, i)
	}
	return c.doProcessWeightRouter(readyInstance)
}

func (c *DiscoveryCli) processTargetRouter(srcInstanceName string, dstService string) (*ProcessRouterResult, error) {
	target, ok := c.DataMgr.GetTargetRouter(dstService, srcInstanceName)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	i, ok := c.DataMgr.GetInstanceInfoByName(dstService, target.DstInstanceName)
	if !ok || !i.Breaker.Ready() {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: i}, nil

}

func (c *DiscoveryCli) processKeyValueRouter(
	srcInstanceName string,
	key map[string]string,
	dstService string,
) (*ProcessRouterResult, error) {
	vs, ok := c.DataMgr.GetKVRouter(dstService, key)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	v := util.RandomWeightItem(vs)
	allInstance := make([]*dataMgr.InstanceInfo, 0, len(v.DstInstanceName))
	for _, name := range v.DstInstanceName {
		nowInstance, ok := c.DataMgr.GetInstanceInfoByName(dstService, name)
		if !ok {
			continue
		}
		allInstance = append(allInstance, nowInstance)
	}
	if len(allInstance) == 0 {
		return nil, errors.New("没有目标实例")
	} else if len(allInstance) == 1 {
		if allInstance[0].Breaker.Ready() {
			return &ProcessRouterResult{DstInstance: allInstance[0]}, nil
		}
		return nil, errors.New("没有目标实例")
	}
	readyInstance := make([]util.DstInstanceInfo, 0, len(allInstance))
	for _, i := range allInstance {
		if !i.Breaker.Ready() {
			continue
		}
		readyInstance = append(readyInstance, i)
	}
	if len(readyInstance) == 0 {
		return nil, errors.New("没有目标实例")
	}
	switch v.RouterType {
	case util.ConsistentRouterType:
		return c.doProcessConsistentRouter(srcInstanceName, readyInstance)
	case util.RandomRouterType:
		return c.doProcessRandomRouter(readyInstance)
	case util.WeightedRouterType:
		return c.doProcessWeightRouter(readyInstance)
	default:
		return nil, fmt.Errorf("不支持进一步路由格式%v", v.RouterType)
	}
}

// ProcessRouter 如果没有提供可选的Instance，那么自动从缓冲中获取
func (c *DiscoveryCli) ProcessRouter(argv *ProcessRouterArgv) (result *ProcessRouterResult, err error) {
	begin := time.Now()
	defer func() {
		c.Metrics.MetricsUpload(begin, prometheus.Labels{"Method": "ProcessRouter"}, err)
	}()
	switch argv.Method {
	case util.TargetRouterType:
		return c.processTargetRouter(argv.SrcInstanceName, argv.DstService)
	case util.KVRouterType:
		if len(argv.Key) == 0 {
			return nil, errors.New("使用键值对路由但是缺少Key")
		}
		return c.processKeyValueRouter(argv.SrcInstanceName, argv.Key, argv.DstService)
	case util.ConsistentRouterType:
		return c.processConsistentRouter(argv.SrcInstanceName, argv.DstService)
	case util.RandomRouterType:
		return c.processRandomRouter(argv.DstService)
	case util.WeightedRouterType:
		return c.processWeightRouter(argv.DstService)
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
	DstService      string
	//可选的，只有kv需要
	Key map[string]string
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
	ctx := NewAPIContextStandalone(config, conn, mt)
	return &DiscoveryCli{
		APIContext: ctx,
		DataMgr:    dmgr,
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

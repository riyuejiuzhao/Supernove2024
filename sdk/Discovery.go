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
	"math/rand"
	"stathat.com/c/consistent"
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
	hash := consistent.New()
	dict := make(map[string]util.DstInstanceInfo)
	for _, v := range dstInstances {
		address := util.Address(v.GetHost(), v.GetPort())
		dict[address] = v
		hash.Add(address)
	}
	dstInstanceID, err := hash.Get(srcInstanceName)
	if err != nil {
		return nil, err
	}
	return &ProcessRouterResult{DstInstance: dict[dstInstanceID]}, err
}

func (c *DiscoveryCli) processConsistentRouter(
	srcInstanceName string,
	dstService string,
) (*ProcessRouterResult, error) {
	instances, ok := c.DataMgr.GetServiceInfo(dstService)
	if !ok || len(instances.Instances) == 0 {
		return nil, errors.New("没有对应实例")
	}
	return c.doProcessConsistentRouter(srcInstanceName, instances.Instances)
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
	return c.doProcessRandomRouter(serviceInfo.Instances)
}

func (c *DiscoveryCli) doProcessWeightRouter(dstInstances []util.DstInstanceInfo) (*ProcessRouterResult, error) {
	allWeight := int32(0) //dstInstances[0].GetWeight()
	for _, v := range dstInstances {
		allWeight += v.GetWeight()
	}
	chosen := rand.Int31n(allWeight)
	for _, v := range dstInstances {
		if chosen > v.GetWeight() {
			chosen -= v.GetWeight()
			continue
		}
		return &ProcessRouterResult{DstInstance: v}, nil
	}
	return nil, fmt.Errorf("没有随机到结果")
}

func (c *DiscoveryCli) processWeightRouter(dstInstances string) (*ProcessRouterResult, error) {
	service, ok := c.DataMgr.GetServiceInfo(dstInstances)
	if !ok {
		return nil, fmt.Errorf("没有服务%s", dstInstances)
	}
	return c.doProcessWeightRouter(service.Instances)
}

func (c *DiscoveryCli) processTargetRouter(srcInstanceName string, dstService string) (*ProcessRouterResult, error) {
	target, ok := c.DataMgr.GetTargetRouter(dstService, srcInstanceName)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	i, ok := c.DataMgr.GetInstanceInfoByName(dstService, target.DstInstanceName)
	if !ok {
		return nil, errors.New("没有目标实例")
	}
	return &ProcessRouterResult{DstInstance: i}, nil

}

func (c *DiscoveryCli) processKeyValueRouter(
	srcInstanceName string,
	key map[string]string,
	dstService string,
) (*ProcessRouterResult, error) {
	v, ok := c.DataMgr.GetKVRouter(dstService, key)
	if !ok {
		return nil, errors.New("没有目标路由")
	}
	allInstance := make([]util.DstInstanceInfo, 0, len(v.DstInstanceName))
	for _, name := range v.DstInstanceName {
		nowInstance, ok := c.DataMgr.GetInstanceInfoByName(dstService, name)
		if !ok {
			continue
		}
		allInstance = append(allInstance, nowInstance)
	}
	if len(allInstance) == 0 {
		return nil, errors.New("没有目标路由")
	} else if len(allInstance) == 1 {
		return &ProcessRouterResult{DstInstance: allInstance[0]}, nil
	}
	switch v.RouterType {
	case util.ConsistentRouterType:
		return c.doProcessConsistentRouter(srcInstanceName, allInstance)
	case util.RandomRouterType:
		return c.doProcessRandomRouter(allInstance)
	case util.WeightedRouterType:
		return c.doProcessWeightRouter(allInstance)
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

package sdk

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/dataMgr"
)

type GetInstancesResult struct {
	Instances []*miniRouterProto.InstanceInfo
}

type DiscoveryCli struct {
	APIContext
	dataMgr dataMgr.ServiceDataManager
}

func (c *DiscoveryCli) GetInstances(serviceName string) (*GetInstancesResult, error) {
	//在缓存数据中查找
	service := c.dataMgr.GetServiceInstance(serviceName)
	return &GetInstancesResult{Instances: service.Instances}, nil
}

type DiscoveryAPI interface {
	GetInstances(serviceName string) (*GetInstancesResult, error)
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

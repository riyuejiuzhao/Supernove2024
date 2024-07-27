package sdk

import (
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/util"
)

type GetInstancesResult struct {
	Instances []util.InstanceBaseInfo
}

type DiscoveryCli struct {
	util.APIContext
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
	ctx, err := util.NewAPIContext()
	if err != nil {
		return nil, err
	}
	dataManager, err := dataMgr.Instance()
	if err != nil {
		return nil, err
	}
	return &DiscoveryCli{*ctx, dataManager}, nil
}

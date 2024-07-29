package dataMgr

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
)

type ServiceDataManager interface {
	GetHealthInfo(serviceName string, instanceID string) (*miniRouterProto.InstanceHealthInfo, bool)
	GetServiceInstance(serviceName string) *miniRouterProto.ServiceInfo
	FlushService(serviceName string)
}

var (
	serviceDataMgr        ServiceDataManager = nil
	NewServiceDataManager                    = NewDefaultServiceMgr
)

func Instance() (ServiceDataManager, error) {
	if serviceDataMgr == nil {
		cfg, err := config.GlobalConfig()
		if err != nil {
			return nil, err
		}
		conn, err := connMgr.Instance()
		if err != nil {
			return nil, err
		}
		serviceDataMgr = NewServiceDataManager(cfg, conn)
	}
	return serviceDataMgr, nil
}

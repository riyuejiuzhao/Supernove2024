package dataMgr

import "Supernove2024/miniRouterProto"

type ServiceDataManager interface {
	GetServiceInstance(serviceName string) miniRouterProto.ServiceInfo
	GetRevision() int64
	Flush()
}

var (
	ServiceDataMgr        ServiceDataManager = nil
	NewServiceDataManager                    = NewDefaultServiceMgr
)

func Instance() (ServiceDataManager, error) {
	if ServiceDataMgr == nil {

	}
	return ServiceDataMgr, nil
}

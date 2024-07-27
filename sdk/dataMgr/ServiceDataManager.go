package dataMgr

import "Supernove2024/sdk/util"

type ServiceDataManager interface {
	GetServiceInstance(serviceName string) util.ServiceInfo
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

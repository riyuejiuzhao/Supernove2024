package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
)

type ServiceDataManager interface {
	GetHealthInfo(serviceName string, instanceID string) (*pb.InstanceHealthInfo, bool)
	GetServiceInfo(serviceName string) (*pb.ServiceInfo, bool)
	GetInstanceInfo(serviceName string, instanceID string) (*pb.InstanceInfo, bool)
	GetTargetRouter(ServiceName string, SrcInstanceID string, skipTimeCheck bool) (*pb.TargetRouterInfo, bool)
	GetKVRouter(ServiceName string, Key string, skipTimeCheck bool) (*pb.KVRouterInfo, bool)
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

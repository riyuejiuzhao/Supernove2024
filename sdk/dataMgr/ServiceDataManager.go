package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/metrics"
)

type ServiceDataManager interface {
	GetServiceInfo(serviceName string) (*pb.ServiceInfo, bool)
	GetInstanceInfo(serviceName string, instanceID int64) (*pb.InstanceInfo, bool)
	GetTargetRouter(ServiceName string, SrcInstanceID int64) (*pb.TargetRouterInfo, bool)
	GetKVRouter(ServiceName string, Key string) (*pb.KVRouterInfo, bool)
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
		mt, err := metrics.Instance()
		if err != nil {
			return nil, err
		}
		serviceDataMgr = NewServiceDataManager(cfg, conn, mt)
	}
	return serviceDataMgr, nil
}

package dataMgr

import (
	"Supernove2024/pb"
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	circuit "github.com/rubyist/circuitbreaker"
	"sync"
)

type InstanceInfo struct {
	*pb.InstanceInfo
	Breaker *circuit.Breaker
}

func (m *InstanceInfo) GetBreaker() *circuit.Breaker {
	return m.Breaker
}

type ServiceDataManager interface {
	GetServiceInfo(serviceName string) (*util.ServiceInfo, bool)
	WatchServiceInfo(serviceName string) (<-chan *util.ServiceInfo, error)
	GetInstanceInfo(serviceName string, instanceID int64) (*InstanceInfo, bool)
	GetInstanceInfoByName(serviceName string, name string) (*InstanceInfo, bool)
	GetTargetRouter(ServiceName string, SrcInstanceName string) (*pb.TargetRouterInfo, bool)
	GetKVRouter(ServiceName string, Key map[string]string) ([]*pb.KVRouterInfo, bool)
}

var (
	serviceDataMgr        ServiceDataManager = nil
	NewServiceDataManager                    = NewDefaultServiceMgr
)

var dataMgrMutex sync.Mutex

func Instance() (ServiceDataManager, error) {
	dataMgrMutex.Lock()
	defer dataMgrMutex.Unlock()
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

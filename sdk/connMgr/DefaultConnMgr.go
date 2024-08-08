package connMgr

import (
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DefaultConnManager struct {
	config  *config.Config
	poolDic map[ServiceType]map[string]*grpc.ClientConn
}

func newAddressPoolDic(serviceConfig []config.InstanceConfig) map[string]*grpc.ClientConn {
	dic := make(map[string]*grpc.ClientConn)
	for _, service := range serviceConfig {
		address := service.String()
		client, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			util.Error("创建链接 err: %v", err)
		}
		dic[address] = client
	}
	return dic
}

func newDefaultConnManager(config *config.Config) ConnManager {
	dic := make(map[ServiceType]map[string]*grpc.ClientConn)
	dic[Discovery] = newAddressPoolDic(
		config.Global.DiscoverService)
	dic[Register] = newAddressPoolDic(
		config.Global.RegisterService)
	dic[HealthCheck] = newAddressPoolDic(
		config.Global.HealthService)
	return &DefaultConnManager{poolDic: dic}
}

// GetServiceConn 指定服务的链接
func (m *DefaultConnManager) GetServiceConn(service ServiceType) (*grpc.ClientConn, error) {
	if service >= ServiceTypeCount {
		return nil, errors.New("指定的服务不存在")
	}
	servicesDic := m.poolDic[service]
	p := util.RandomDicValue(servicesDic)
	if p == nil {
		return nil, errors.New("指定的服务不存在")
	}
	return p, nil
}

// GetConn 指定地址的链接
func (m *DefaultConnManager) GetConn(service ServiceType, address string) (*grpc.ClientConn, error) {
	conn, ok := m.poolDic[service][address]
	if !ok {
		newConn, err := grpc.NewClient(address)
		if err != nil {
			return nil, err
		}
		m.poolDic[service][address] = newConn
		conn = newConn
	}
	return conn, nil
}

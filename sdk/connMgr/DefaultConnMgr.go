package connMgr

import (
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type DefaultConnManager struct {
	config  *config.Config
	poolDic map[ServiceType]*clientv3.Client
}

func newAddressPoolDic(serviceConfig []config.InstanceConfig) (*clientv3.Client, error) {
	addresses := util.Map(serviceConfig, func(t config.InstanceConfig) string {
		return util.Address(t.Host, t.Port)
	})
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   addresses,
		DialTimeout: 5 * time.Second,
	})
	return client, err
}

func newDefaultConnManager(config *config.Config) (ConnManager, error) {
	dic := make(map[ServiceType]*clientv3.Client)
	p, err := newAddressPoolDic(config.Global.EtcdService)
	if err != nil {
		return nil, err
	}
	dic[Etcd] = p
	return &DefaultConnManager{poolDic: dic}, nil
}

// GetServiceConn 指定服务的链接
func (m *DefaultConnManager) GetServiceConn(service ServiceType) (*clientv3.Client, error) {
	if service >= ServiceTypeCount {
		return nil, errors.New("指定的服务不存在")
	}
	p := m.poolDic[service]
	return p, nil
}

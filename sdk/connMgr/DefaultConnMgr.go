package connMgr

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/howeyc/crc16"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type ClientInfo struct {
	*clientv3.Client
	Address string
}

type DefaultConnManager struct {
	mt      *metrics.MetricsManager
	poolDic map[ServiceType][]*ClientInfo
}

const HashSlots = 16384

func getHashSlot(key string) int {
	hashValue := crc16.ChecksumIBM([]byte(key))
	return int(hashValue % HashSlots)
}

func newAddressPoolDic(serviceConfig []config.InstanceConfig) (cli []*ClientInfo, err error) {
	cli = make([]*ClientInfo, 0)
	for _, cfg := range serviceConfig {
		var client *clientv3.Client
		address := util.Address(cfg.Host, cfg.Port)
		client, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{address},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return
		}
		cli = append(cli, &ClientInfo{
			Client:  client,
			Address: address,
		})
	}
	return
}

func newDefaultConnManager(cfg *config.Config) (ConnManager, error) {
	mt, err := metrics.Instance()
	if err != nil {
		return nil, err
	}

	dic := make(map[ServiceType][]*ClientInfo)

	p, err := newAddressPoolDic(cfg.SDK.InstancesEtcd)
	if err != nil {
		return nil, err
	}
	dic[InstancesEtcd] = p

	p, err = newAddressPoolDic(cfg.SDK.RouterEtcd)
	if err != nil {
		return nil, err
	}
	dic[RoutersEtcd] = p
	return &DefaultConnManager{poolDic: dic, mt: mt}, nil
}

func (m *DefaultConnManager) GetAllServiceConn(service ServiceType) []*clientv3.Client {
	rt := util.SliceMap(m.poolDic[service], func(t *ClientInfo) *clientv3.Client {
		return t.Client
	})
	return rt
}

// GetServiceConn 指定服务的链接
// key用来一致性哈希
func (m *DefaultConnManager) GetServiceConn(service ServiceType, key string) (cli *clientv3.Client, err error) {
	var connAddress string
	defer func() {
		m.mt.ConnCount.With(prometheus.Labels{"Address": connAddress}).Inc()
	}()
	if service >= ServiceTypeCount {
		err = errors.New("指定的服务不存在")
		return
	}
	p := m.poolDic[service]
	switch service {
	case InstancesEtcd:
		rt := util.RandomItem(p)
		cli = rt.Client
		connAddress = rt.Address
		return //util.RandomItem(p).Client, nil
	case RoutersEtcd:
		slot := getHashSlot(key)
		cliInfo := hashSlot2Client(slot, p)
		connAddress = cliInfo.Address
		cli = cliInfo.Client
		return
	default:
		err = fmt.Errorf("不支持的连接类型")
		return
	}
}

// 我们假定slot在不同的集群上分布是均匀的
func hashSlot2Client(slot int, infos []*ClientInfo) *ClientInfo {
	return infos[slot%len(infos)]
}

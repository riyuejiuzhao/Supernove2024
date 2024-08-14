package connMgr

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/metrics"
	"Supernove2024/util"
	"errors"
	"github.com/howeyc/crc16"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type ClientInfo struct {
	*clientv3.Client
	Name string
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

func newAddressPoolDic(serviceConfig []config.ClusterConfig) (cli []*ClientInfo, err error) {
	cli = make([]*ClientInfo, 0, len(serviceConfig))
	for _, cfg := range serviceConfig {
		var client *clientv3.Client
		address := util.SliceMap(cfg.Pod, func(t config.PodConfig) string {
			return util.Address(t.Host, t.Port)
		})
		client, err = clientv3.New(clientv3.Config{
			Endpoints:   address,
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return
		}
		cli = append(cli, &ClientInfo{
			Client: client,
			Name:   cfg.Name,
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
	var connName string
	defer func() {
		m.mt.ConnCount.With(prometheus.Labels{"Name": connName}).Inc()
	}()
	if service >= ServiceTypeCount {
		err = errors.New("指定的服务不存在")
		return
	}
	p := m.poolDic[service]
	slot := getHashSlot(key)
	cliInfo := hashSlot2Client(slot, p)
	connName = cliInfo.Name
	cli = cliInfo.Client
	return
}

// 我们假定slot在不同的集群上分布是均匀的
func hashSlot2Client(slot int, infos []*ClientInfo) *ClientInfo {
	return infos[slot%len(infos)]
}

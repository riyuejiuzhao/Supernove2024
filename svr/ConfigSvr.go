package svr

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"time"
)

type ClusterConfig struct {
	Name string       `yaml:"Name"`
	Pod  []*PodConfig `yaml:"Pod"`
}

type PodConfig struct {
	Host string `yaml:"Host"`
	Port int32  `yaml:"Port"`
}

type Config struct {
	InstancesEtcd []*ClusterConfig `yaml:"InstancesEtcd"`
	RouterEtcd    []*ClusterConfig `yaml:"RouterEtcd"`
}

type ConfigSvr struct {
	pb.UnimplementedConfigServiceServer
	etcdConfig *Config
}

func NewConfigSvr(configFile string) (srv *ConfigSvr, err error) {
	configYaml, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return nil, err
	}

	srv = &ConfigSvr{etcdConfig: &config}
	return
}

func (s *ConfigSvr) configToPb(cs []*ClusterConfig) []*pb.ClusterInfo {
	return util.SliceMap(cs, func(t *ClusterConfig) *pb.ClusterInfo {
		address := make([]string, 0, len(t.Pod))
		for _, p := range t.Pod {
			address = append(address, util.Address(p.Host, p.Port))
		}
		return &pb.ClusterInfo{
			Name:    t.Name,
			Address: address,
		}
	})
}

func (s *ConfigSvr) Init(context.Context, *pb.InitArgv) (*pb.InitResult, error) {
	return &pb.InitResult{
		Instances: s.configToPb(s.etcdConfig.InstancesEtcd),
		Routers:   s.configToPb(s.etcdConfig.RouterEtcd),
	}, nil
}

func checkEtcd(clusters []*ClusterConfig) {
	for _, cluster := range clusters {
		address := util.SliceMap(cluster.Pod, func(t *PodConfig) string {
			return util.Address(t.Host, t.Port)
		})
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   address,
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			util.Error("check etcd %s err: %v", cluster.Name, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err = client.Get(ctx, "Valid")
		if err != nil {
			util.Error("check etcd %s err: %v", cluster.Name, err)
		}
		cancel()
	}
}

func (s *ConfigSvr) Serve(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	//检测配置中的etcd是否都能相应
	checkEtcd(s.etcdConfig.InstancesEtcd)
	checkEtcd(s.etcdConfig.RouterEtcd)

	grpcSrv := grpc.NewServer()
	pb.RegisterConfigServiceServer(grpcSrv, s)
	return grpcSrv.Serve(lis)
}

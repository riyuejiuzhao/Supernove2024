package svr

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"context"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"net"
	"os"
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
	return util.SliceMap(s.etcdConfig.InstancesEtcd, func(t *ClusterConfig) *pb.ClusterInfo {
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

func (s *ConfigSvr) Serve(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterConfigServiceServer(grpcSrv, s)
	return grpcSrv.Serve(lis)
}

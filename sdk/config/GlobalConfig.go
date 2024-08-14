package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type InstanceConfig struct {
	Host string `yaml:"Host"`
	Port int32  `yaml:"Port"`
}

func (s InstanceConfig) String() string {
	return fmt.Sprintf("%s:%v", s.Host, s.Port)
}

type Config struct {
	SDK struct {
		InstancesEtcd []InstanceConfig `yaml:"InstancesEtcd"`
		RouterEtcd    []InstanceConfig `yaml:"RouterEtcd"`
		Register      struct {
			DefaultWeight     int32 `yaml:"DefaultWeight"`
			DefaultServiceTTL int64 `yaml:"DefaultServiceTTL"`
			DefaultRouterTTL  int64 `yaml:"DefaultRouterTTL"`
		} `yaml:"Register"`
		Discovery struct {
			DstService []string `yaml:"DstService"`
		} `yaml:"Discovery"`
		Metrics string `yaml:"Metrics"`
	} `yaml:"SDK"`
}

var globalConfig *Config = nil
var GlobalConfigFilePath = "mini-router.yaml"

var configMutex sync.Mutex

func GlobalConfig() (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()
	if globalConfig != nil {
		return globalConfig, nil
	}
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	globalConfig = config
	return globalConfig, nil
}

// LoadConfig 从文件中读取配置
func LoadConfig(configFileOpts ...string) (*Config, error) {
	configFile := GlobalConfigFilePath
	if len(configFileOpts) == 1 {
		configFile = configFileOpts[0]
	} else if len(configFileOpts) > 1 {
		return nil, errors.New("配置文件路径数量超过1")
	}
	configYaml, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

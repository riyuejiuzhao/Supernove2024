package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type InstanceConfig struct {
	Host string `yaml:"Host"`
	Port int32  `yaml:"Port"`
}

func (s InstanceConfig) String() string {
	return fmt.Sprintf("%s:%v", s.Host, s.Port)
}

type Config struct {
	Global struct {
		DiscoverService []InstanceConfig `yaml:"DiscoverService"`
		RegisterService []InstanceConfig `yaml:"RegisterService"`
		HealthService   []InstanceConfig `yaml:"HealthService"`
		Register        struct {
			DefaultWeight int32 `yaml:"DefaultWeight"`
			DefaultTTL    int64 `yaml:"DefaultTTL"`
		} `yaml:"Register"`
		Discovery struct {
			//获取远程服务的时间间隔
			DefaultTimeout int32 `yaml:"DefaultTimeout"`
			//Service
			DstService []string `yaml:"DstService"`
		} `yaml:"Discovery"`
	} `yaml:"Global"`
}

var globalConfig *Config = nil
var GlobalConfigFilePath = ""

func GlobalConfig() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	globalConfig = config
	return globalConfig, nil
}

// 从文件中读取配置
func loadConfig(configFileOpts ...string) (*Config, error) {
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

package sdk

import (
	config2 "Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
)

type APIContext struct {
	//配置
	Config *config2.Config
	//连接管理
	ConnManager connMgr.ConnManager
	//缓存管理
}

func NewAPIContext() (*APIContext, error) {
	globalConfig, err := config2.GlobalConfig()
	if err != nil {
		return nil, err
	}
	//随机选择一个配置中的DiscoverySvr
	connManager, err := connMgr.Instance()
	if err != nil {
		return nil, err
	}
	api := APIContext{ConnManager: connManager,
		Config: globalConfig}
	return &api, nil
}

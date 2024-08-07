package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
)

type APIContext struct {
	//配置
	Config *config.Config
	//连接管理
	ConnManager connMgr.ConnManager
}

func NewAPIContextStandalone(config *config.Config, conn connMgr.ConnManager) *APIContext {
	return &APIContext{
		Config:      config,
		ConnManager: conn,
	}
}

func NewAPIContext() (*APIContext, error) {
	globalConfig, err := config.GlobalConfig()
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

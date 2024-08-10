package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
)

type APIContext struct {
	//配置
	Config *config.Config
	//连接管理
	ConnManager connMgr.ConnManager
	//缓存
	DataMgr dataMgr.ServiceDataManager
}

func NewAPIContextStandalone(config *config.Config, conn connMgr.ConnManager, dataManger dataMgr.ServiceDataManager) *APIContext {
	return &APIContext{
		Config:      config,
		ConnManager: conn,
		DataMgr:     dataManger,
	}
}

func NewAPIContext() (*APIContext, error) {
	globalConfig, err := config.GlobalConfig()
	if err != nil {
		return nil, err
	}
	connManager, err := connMgr.Instance()
	if err != nil {
		return nil, err
	}
	mgr, err := dataMgr.Instance()
	if err != nil {
		return nil, err
	}
	api := APIContext{ConnManager: connManager,
		Config: globalConfig, DataMgr: mgr}
	return &api, nil
}

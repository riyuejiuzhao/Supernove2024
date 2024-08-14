package sdk

import (
	"Supernove2024/sdk/config"
	"Supernove2024/sdk/connMgr"
	"Supernove2024/sdk/dataMgr"
	"Supernove2024/sdk/metrics"
)

type APIContext struct {
	//配置
	Config *config.Config
	//连接管理
	ConnManager connMgr.ConnManager

	Metrics *metrics.MetricsManager
}

func NewAPIContextStandalone(
	config *config.Config,
	conn connMgr.ConnManager,
	dataManger dataMgr.ServiceDataManager,
	mt *metrics.MetricsManager,
) *APIContext {
	return &APIContext{
		Config:      config,
		ConnManager: conn,
		//DataMgr:     dataManger,
		Metrics: mt,
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

	metricsMgr, err := metrics.Instance()
	if err != nil {
		return nil, err
	}

	api := APIContext{
		ConnManager: connManager,
		Config:      globalConfig,
		//DataMgr:     mgr,
		Metrics: metricsMgr,
	}

	return &api, nil
}

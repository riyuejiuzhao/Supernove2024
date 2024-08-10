package metrics

import (
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type MetricsManager struct {
	//Metrics
	MethodCounter *prometheus.CounterVec
	MethodTime    *prometheus.SummaryVec
}

var (
	metricsMgr *MetricsManager
)

func NewMetricsMgr(cfg *config.Config) *MetricsManager {
	mMgr := &MetricsManager{
		MethodCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "MethodCounter_count",
			},
			[]string{"Method"},
		),
		MethodTime: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "MethodTime_sum",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"Method"},
		),
	}

	go func() {
		if cfg.Global.Metrics == "" {
			return
		}
		prometheus.MustRegister(mMgr.MethodCounter, mMgr.MethodTime)
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.Global.Metrics, nil)
		if err != nil {
			util.Error("%v", err)
		}
	}()
	return mMgr
}

func Instance() (*MetricsManager, error) {
	if metricsMgr == nil {
		cfg, err := config.GlobalConfig()
		if err != nil {
			return nil, err
		}
		metricsMgr = NewMetricsMgr(cfg)
	}
	return metricsMgr, nil
}

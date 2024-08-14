package metrics

import (
	"Supernove2024/sdk/config"
	"Supernove2024/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
	"time"
)

type MetricsManager struct {
	//Metrics
	MethodCounter *prometheus.CounterVec
	MethodTime    *prometheus.SummaryVec
	MethodFailed  *prometheus.CounterVec
	// 统计每个链接连接次数
	ConnCount *prometheus.CounterVec
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
		MethodFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "MethodFailed_count",
			},
			[]string{"Method"},
		),
		ConnCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "Conn_count",
			},
			[]string{"Address"},
		),
	}

	go func() {
		if cfg.SDK.Metrics == "" {
			return
		}
		prometheus.MustRegister(
			mMgr.MethodCounter,
			mMgr.MethodTime,
			mMgr.MethodFailed,
			mMgr.ConnCount,
		)
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.SDK.Metrics, nil)
		if err != nil {
			util.Error("%v", err)
		}
	}()
	return mMgr
}

var metricsMutex sync.Mutex

func Instance() (*MetricsManager, error) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	if metricsMgr == nil {
		cfg, err := config.GlobalConfig()
		if err != nil {
			return nil, err
		}
		metricsMgr = NewMetricsMgr(cfg)
	}
	return metricsMgr, nil
}

func (mt *MetricsManager) MetricsUpload(begin time.Time, label prometheus.Labels, err error) {
	duration := time.Since(begin).Milliseconds()
	mt.MethodTime.With(label).Observe(float64(duration))
	mt.MethodCounter.With(label).Inc()
	if err != nil {
		mt.MethodFailed.With(label).Inc()
	}
}

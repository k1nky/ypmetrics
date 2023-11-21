package poller

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// хранилище метрик
type metricStorage interface {
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	Snapshot(ctx context.Context, metrics *metric.Metrics) error
	UpdateMetrics(ctx context.Context, metrics metric.Metrics) error
}

// логер
type logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

// отправитель метрик на сервер
type sender interface {
	PushCounter(name string, value int64) error
	PushGauge(name string, value float64) error
	PushMetrics(metrics metric.Metrics) error
}

// сборщик метрик
type Collector interface {
	Collect(ctx context.Context) (metric.Metrics, error)
	Init() error
}

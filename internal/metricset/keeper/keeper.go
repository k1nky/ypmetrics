package keeper

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
)

type logger interface {
	Info(template string, args ...interface{})
}

// Keeper представляет собой набор метрик. В текущей реализации практически ничем
// не отличается от metricset.Set, сделан для будущего использования.
type Keeper struct {
	metricset.Set
	logger logger
}

type metricStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	SetCounter(*metric.Counter)
	SetGauge(*metric.Gauge)
	Snapshot(*metric.Metrics)
}

func New(storage metricStorage, log logger) *Keeper {
	metricSet := metricset.NewSet(storage)
	if metricSet == nil {
		return nil
	}
	return &Keeper{
		Set:    *metricSet,
		logger: log,
	}
}

package keeper

import (
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type metricStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	Snapshot(*metric.Metrics)
}

// Keeper представляет собой набор метрик. В текущей реализации представляет
// функционал storage.Storage.
type Keeper struct {
	metricStorage
}

func New(store metricStorage) *Keeper {
	return &Keeper{
		metricStorage: store,
	}
}

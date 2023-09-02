// Пакет storage реализует хранилище метрик
package storage

import "github.com/k1nky/ypmetrics/internal/entities/metric"

type storageLogger interface {
	Error(template string, args ...interface{})
}

type Storage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	Snapshot(*metric.Metrics)
	Close() error
}

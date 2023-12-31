// Пакет storage реализует хранилище метрик
package storage

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type storageLogger interface {
	Errorf(template string, args ...interface{})
}

type storageRetrier interface {
	Init(func(error) bool)
	Next(error) bool
}

//go:generate mockgen -source=contract.go -destination=mock/storage.go -package=mock Storage
type Storage interface {
	Open(cfg Config) error
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateMetrics(ctx context.Context, metrics metric.Metrics) error
	Snapshot(ctx context.Context, metrics *metric.Metrics) error
	Close() error
}

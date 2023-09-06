// Пакет storage реализует хранилище метрик
package storage

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type storageLogger interface {
	Error(template string, args ...interface{})
}

//go:generate mockgen -source=contract.go -destination=mock/storage.go -package=mock Storage
type Storage interface {
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	UpdateCounter(ctx context.Context, name string, value int64)
	UpdateGauge(ctx context.Context, name string, value float64)
	Snapshot(ctx context.Context, metrics *metric.Metrics)
	Close() error
}

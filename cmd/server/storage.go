package main

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type metricStorage interface {
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateMetrics(ctx context.Context, metrics metric.Metrics) error
	Snapshot(ctx context.Context, metrics *metric.Metrics) error
	Close() error
}

func openStorage(cfg config.KeeperConfig, log *logger.Logger) (metricStorage, error) {
	switch {
	case len(cfg.DatabaseDSN) > 0:
		s := storage.NewDBStorage(log)
		return s, s.Open(cfg.DatabaseDSN)
	case len(cfg.FileStoragePath) > 0:
		if cfg.StoreIntervalInSec == 0 {
			s := storage.NewSyncFileStorage(log)
			return s, s.Open(cfg.FileStoragePath, cfg.Restore)
		} else {
			s := storage.NewAsyncFileStorage(log)
			return s, s.Open(cfg.FileStoragePath, cfg.Restore, cfg.StorageInterval())
		}
	default:
		return storage.NewMemStorage(), nil
	}
}

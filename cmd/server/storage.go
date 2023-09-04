package main

import (
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type metricStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	Snapshot(*metric.Metrics)
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

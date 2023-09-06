package keeper

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type metricStorage interface {
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	UpdateCounter(ctx context.Context, name string, value int64)
	UpdateGauge(ctx context.Context, name string, value float64)
	Snapshot(ctx context.Context, metrics *metric.Metrics)
}

type logger interface {
	Error(template string, args ...interface{})
}

// Keeper представляет собой набор метрик. В текущей реализации представляет
// функционал storage.Storage.
type Keeper struct {
	metricStorage
	config config.KeeperConfig
	logger logger
}

func New(store metricStorage, cfg config.KeeperConfig, logger logger) *Keeper {
	return &Keeper{
		metricStorage: store,
		config:        cfg,
		logger:        logger,
	}
}

// Ping проверяет подключение к базе данных.
func (k *Keeper) Ping(ctx context.Context) error {
	db := storage.NewDBStorage(k.logger)
	if err := db.Open(k.config.DatabaseDSN); err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

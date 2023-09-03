package keeper

import (
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
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
	config config.KeeperConfig
}

func New(store metricStorage, cfg config.KeeperConfig) *Keeper {
	return &Keeper{
		metricStorage: store,
		config:        cfg,
	}
}

// Ping проверяет подключение к базе данных.
func (k *Keeper) Ping() error {
	db := storage.NewDBStorage()
	if err := db.Open("pgx", k.config.DatabaseDSN); err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

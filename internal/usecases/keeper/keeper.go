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
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateMetrics(ctx context.Context, metrics metric.Metrics) error
	Snapshot(ctx context.Context, metrics *metric.Metrics) error
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

func uniqueCounters(counters []*metric.Counter) []*metric.Counter {
	// карта соответствия имени метрики - индексу это метрики в результирующем массиве
	names := make(map[string]int, 0)
	result := make([]*metric.Counter, 0, len(counters))
	for _, m := range counters {
		if dx, ok := names[m.Name]; !ok {
			// такой метрики еще не было
			// запоминаем ее индекс в итоговом массиве
			names[m.Name] = len(result)
			// добавляем в итоговый массив
			result = append(result, m)
		} else {
			// метрика уже была, обновляем ее значение
			result[dx].Update(m.Value)
		}
	}
	return result
}

func uniqueGauge(gauges []*metric.Gauge) []*metric.Gauge {
	names := make(map[string]int, 0)
	result := make([]*metric.Gauge, 0, len(gauges))
	for _, m := range gauges {
		if dx, ok := names[m.Name]; !ok {
			names[m.Name] = len(result)
			result = append(result, m)
		} else {
			result[dx].Update(m.Value)
		}
	}
	return result
}

func (k *Keeper) UpdateMetrics(ctx context.Context, metrics metric.Metrics) error {
	// оставляем только уникальные метрики
	// обновления из дублирующих метрик применяются последовательно
	metrics.Counters = uniqueCounters(metrics.Counters)
	metrics.Gauges = uniqueGauge(metrics.Gauges)
	return k.metricStorage.UpdateMetrics(ctx, metrics)
}

// Ping проверяет подключение к базе данных.
func (k *Keeper) Ping(ctx context.Context) error {
	cfg := storage.Config{
		DSN: k.config.DatabaseDSN,
	}
	db := storage.NewDBStorage(k.logger, nil)
	if err := db.Open(cfg); err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

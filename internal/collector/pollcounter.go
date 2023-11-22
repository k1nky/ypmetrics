package collector

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// PollCounter сборщик, который всегда отдает счетчик со значением 1.
// Таким образом можно посчитать количество опросов.
type PollCounter struct{}

// Init инициализирует сборщика.
func (pc *PollCounter) Init() error {
	return nil
}

// Collect возвращает метрики, собранные сборщиком.
func (pc *PollCounter) Collect(ctx context.Context) (metric.Metrics, error) {
	return metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("PollCount", 1)},
	}, nil
}

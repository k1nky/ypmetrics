package collector

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Сборщик PollCounter всегда отдает счетчик со значением 1
// таким образом можно посчитать количество опросов
type PollCounter struct{}

func (pc *PollCounter) Init() error {
	return nil
}

func (pc *PollCounter) Collect(ctx context.Context) (metric.Metrics, error) {
	return metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("PollCount", 1)},
	}, nil
}

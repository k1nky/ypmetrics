package collector

import (
	"github.com/k1nky/ypmetrics/internal/metric"
)

type PollCounter struct{}

func (pc PollCounter) Collect() (metric.Metrics, error) {
	return metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("PollCount", 1)},
	}, nil
}

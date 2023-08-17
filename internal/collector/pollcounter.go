package collector

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
)

type PollCounter struct{}

func (pc PollCounter) Collect() (metricset.Snapshot, error) {
	return metricset.Snapshot{
		Counters: []*metric.Counter{metric.NewCounter("PollCounter", 1)},
	}, nil
}

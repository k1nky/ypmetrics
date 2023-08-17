package collector

import (
	"math/rand"
	"time"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
)

type RandomCollector struct{}

func (rc RandomCollector) Collect() (metricset.Snapshot, error) {
	return metricset.Snapshot{
		Gauges: []*metric.Gauge{metric.NewGauge("RandomValue", randomFloat())},
	}, nil
}

func randomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.NormFloat64()
}

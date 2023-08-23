package collector

import (
	"math/rand"
	"time"

	"github.com/k1nky/ypmetrics/internal/metric"
)

type Random struct{}

func (rc Random) Collect() (metric.Metrics, error) {
	return metric.Metrics{
		Gauges: []*metric.Gauge{metric.NewGauge("RandomValue", randomFloat())},
	}, nil
}

func randomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.NormFloat64()
}

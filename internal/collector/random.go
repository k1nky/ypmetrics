package collector

import (
	"context"
	"math/rand"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Сборщик произвольного значения для метрики RandomValue
type Random struct{}

func (rc *Random) Collect(ctx context.Context) (metric.Metrics, error) {
	return metric.Metrics{
		Gauges: []*metric.Gauge{metric.NewGauge("RandomValue", randomFloat())},
	}, nil
}

func randomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.NormFloat64()
}

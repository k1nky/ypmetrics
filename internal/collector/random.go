package collector

import (
	"context"
	"math/rand"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Сборщик произвольного значения для метрики RandomValue
type Random struct {
	r *rand.Rand
}

func (rc *Random) Init() error {
	rc.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

func (rc *Random) Collect(ctx context.Context) (metric.Metrics, error) {
	return metric.Metrics{
		Gauges: []*metric.Gauge{metric.NewGauge("RandomValue", rc.r.NormFloat64())},
	}, nil
}

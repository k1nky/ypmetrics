package collector

import (
	"context"
	"math/rand"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Random cборщик случайного значения для метрики типа Gauge.
// Используемый генератор случайных чисел инициализируется текущим значением времени.
type Random struct {
	r *rand.Rand
}

// Init инициализирует сборщика.
func (c *Random) Init() error {
	c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

// Collect возвращает метрики, собранные сборщиком.
func (c *Random) Collect(ctx context.Context) (metric.Metrics, error) {
	return metric.Metrics{
		Gauges: []*metric.Gauge{metric.NewGauge("RandomValue", c.r.NormFloat64())},
	}, nil
}

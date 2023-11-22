package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Gops cборщик метрик на основе пакета gopsutil.
type Gops struct{}

// Init инициализирует сборщика.
func (c *Gops) Init() error {
	return nil
}

// Collect возвращает метрики, собранные сборщиком.
func (c *Gops) Collect(ctx context.Context) (metric.Metrics, error) {
	metrics := metric.NewMetrics()
	memstat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return *metrics, err
	}
	cpustat, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return *metrics, err
	}
	metrics.Gauges = append(metrics.Gauges, metric.NewGauge("TotalMemory", float64(memstat.Total)), metric.NewGauge("FreeMemory", float64(memstat.Free)))
	for i, v := range cpustat {
		metrics.Gauges = append(metrics.Gauges, metric.NewGauge(fmt.Sprintf("CPUutilization%d", i+1), v))
	}
	return *metrics, nil
}

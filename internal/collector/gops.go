package collector

import (
	"fmt"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Gops struct{}

func (c Gops) Collect() (metric.Metrics, error) {
	metrics := metric.NewMetrics()
	memstat, err := mem.VirtualMemory()
	if err != nil {
		return *metrics, err
	}
	cpustat, err := cpu.Percent(0, true)
	if err != nil {
		return *metrics, err
	}
	metrics.Gauges = append(metrics.Gauges, metric.NewGauge("TotalMemory", float64(memstat.Total)), metric.NewGauge("FreeMemory", float64(memstat.Free)))
	for i, v := range cpustat {
		metrics.Gauges = append(metrics.Gauges, metric.NewGauge(fmt.Sprintf("CPUutilization%d", i+1), v))
	}
	return *metrics, nil
}

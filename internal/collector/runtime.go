package collector

import (
	"context"
	"runtime"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Сборщик метрик из пакета runtime
type Runtime struct{}

func (c *Runtime) Init() error {
	return nil
}

func (rc *Runtime) Collect(ctx context.Context) (metric.Metrics, error) {
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	return metric.Metrics{
		Gauges: []*metric.Gauge{
			metric.NewGauge("Alloc", float64(memStat.Alloc)),
			metric.NewGauge("BuckHashSys", float64(memStat.BuckHashSys)),
			metric.NewGauge("Frees", float64(memStat.Frees)),
			metric.NewGauge("GCCPUFraction", float64(memStat.GCCPUFraction)),
			metric.NewGauge("GCSys", float64(memStat.GCSys)),
			metric.NewGauge("HeapAlloc", float64(memStat.HeapAlloc)),
			metric.NewGauge("HeapIdle", float64(memStat.HeapIdle)),
			metric.NewGauge("HeapInuse", float64(memStat.HeapInuse)),
			metric.NewGauge("HeapObjects", float64(memStat.HeapObjects)),
			metric.NewGauge("HeapReleased", float64(memStat.HeapReleased)),
			metric.NewGauge("HeapSys", float64(memStat.HeapSys)),
			metric.NewGauge("LastGC", float64(memStat.LastGC)),
			metric.NewGauge("Lookups", float64(memStat.Lookups)),
			metric.NewGauge("MCacheInuse", float64(memStat.MCacheInuse)),
			metric.NewGauge("MCacheSys", float64(memStat.MCacheSys)),
			metric.NewGauge("MSpanInuse", float64(memStat.MSpanInuse)),
			metric.NewGauge("MSpanSys", float64(memStat.MSpanSys)),
			metric.NewGauge("Mallocs", float64(memStat.Mallocs)),
			metric.NewGauge("NextGC", float64(memStat.NextGC)),
			metric.NewGauge("NumForcedGC", float64(memStat.NumForcedGC)),
			metric.NewGauge("NumGC", float64(memStat.NumGC)),
			metric.NewGauge("OtherSys", float64(memStat.OtherSys)),
			metric.NewGauge("PauseTotalNs", float64(memStat.PauseTotalNs)),
			metric.NewGauge("StackInuse", float64(memStat.StackInuse)),
			metric.NewGauge("StackSys", float64(memStat.StackSys)),
			metric.NewGauge("Sys", float64(memStat.Sys)),
			metric.NewGauge("TotalAlloc", float64(memStat.TotalAlloc)),
		},
	}, nil
}

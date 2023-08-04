// Пакет agent реализует агента сбора метрик
package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/apiclient"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type Agent struct {
	metricSet *metric.Set
	client    *apiclient.Client
	Config    config.AgentConfig
}

// список метрик, которые берутся из пакета runtime
var runtimeMetricsList = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

// New возвращает нового агента сбора метрик. По умолчанию в качестве хранилища используется MemStorage.
func New(cfg config.AgentConfig) *Agent {

	a := &Agent{
		client: apiclient.New(cfg.Address.String()),
		Config: cfg,
	}

	a.metricSet = metric.NewSet(storage.NewMemStorage())

	return a
}

func (a *Agent) setupPredefinedMetrics() {
	for _, v := range runtimeMetricsList {
		a.metricSet.UpdateGauge(v, 0)
	}
	a.metricSet.UpdateCounter("PollCounter", 0)
	a.metricSet.UpdateGauge("RandomValue", 0)
}

// Run запускает агента
func (a Agent) Run() {

	var wg sync.WaitGroup

	a.setupPredefinedMetrics()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := a.report(); err != nil {
				fmt.Printf("report error: %s\n", err)
			}
			time.Sleep(time.Duration(a.Config.ReportIntervalInSec) * time.Second)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			a.pollRuntime()
			a.metricSet.UpdateCounter("PollCounter", 1)
			a.metricSet.UpdateGauge("RandomValue", randomFloat())
			fmt.Println("start polling")
			time.Sleep(time.Duration(a.Config.PollIntervalInSec) * time.Second)
		}
	}()
	wg.Wait()
}

func (a Agent) report() error {
	snap := a.metricSet.GetMetrics()
	for _, m := range snap.Counters {
		if err := a.client.PushMetric("counter", m.Name, m.String()); err != nil {
			return err
		}
	}
	for _, m := range snap.Gauges {
		if err := a.client.PushMetric("gauge", m.Name, m.String()); err != nil {
			return err
		}
	}
	return nil
}

func (a Agent) pollRuntime() {
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	a.metricSet.UpdateGauge("Alloc", float64(memStat.Alloc))
	a.metricSet.UpdateGauge("BuckHashSys", float64(memStat.BuckHashSys))
	a.metricSet.UpdateGauge("Frees", float64(memStat.Frees))
	a.metricSet.UpdateGauge("GCCPUFraction", float64(memStat.GCCPUFraction))
	a.metricSet.UpdateGauge("GCSys", float64(memStat.GCSys))
	a.metricSet.UpdateGauge("HeapAlloc", float64(memStat.HeapAlloc))
	a.metricSet.UpdateGauge("HeapIdle", float64(memStat.HeapIdle))
	a.metricSet.UpdateGauge("HeapInuse", float64(memStat.HeapInuse))
	a.metricSet.UpdateGauge("HeapObjects", float64(memStat.HeapObjects))
	a.metricSet.UpdateGauge("HeapReleased", float64(memStat.HeapReleased))
	a.metricSet.UpdateGauge("HeapSys", float64(memStat.HeapSys))
	a.metricSet.UpdateGauge("LastGC", float64(memStat.LastGC))
	a.metricSet.UpdateGauge("Lookups", float64(memStat.Lookups))
	a.metricSet.UpdateGauge("MCacheInuse", float64(memStat.MCacheInuse))
	a.metricSet.UpdateGauge("MCacheSys", float64(memStat.MCacheSys))
	a.metricSet.UpdateGauge("MSpanInuse", float64(memStat.MSpanInuse))
	a.metricSet.UpdateGauge("MSpanSys", float64(memStat.MSpanSys))
	a.metricSet.UpdateGauge("Mallocs", float64(memStat.Mallocs))
	a.metricSet.UpdateGauge("NextGC", float64(memStat.NextGC))
	a.metricSet.UpdateGauge("NumForcedGC", float64(memStat.NumForcedGC))
	a.metricSet.UpdateGauge("NumGC", float64(memStat.NumGC))
	a.metricSet.UpdateGauge("OtherSys", float64(memStat.OtherSys))
	a.metricSet.UpdateGauge("PauseTotalNs", float64(memStat.PauseTotalNs))
	a.metricSet.UpdateGauge("StackInuse", float64(memStat.StackInuse))
	a.metricSet.UpdateGauge("StackSys", float64(memStat.StackSys))
	a.metricSet.UpdateGauge("Sys", float64(memStat.Sys))
	a.metricSet.UpdateGauge("TotalAlloc", float64(memStat.TotalAlloc))
}

func randomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.NormFloat64()
}

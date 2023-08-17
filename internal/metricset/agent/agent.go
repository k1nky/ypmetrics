// Пакет agent реализует агента сбора метрик
package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
)

type Collector interface {
	Collect() (metricset.Snapshot, error)
}

type agentStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	SetCounter(*metric.Counter)
	SetGauge(*metric.Gauge)
	Snapshot(*metricset.Snapshot)
}

type agentLogger interface {
	Info(template string, args ...interface{})
	Error(template string, args ...interface{})
}

type sender interface {
	PushMetric(typ, name, value string) (err error)
}

type Agent struct {
	metricset.Set
	collectors []Collector
	logger     agentLogger
	client     sender
	Config     config.AgentConfig
}

// New возвращает нового агента сбора метрик. По умолчанию в качестве хранилища используется MemStorage.
func New(cfg config.AgentConfig, storage agentStorage, l agentLogger, client sender) *Agent {
	metricSet := metricset.NewSet(storage)
	if metricSet == nil {
		return nil
	}
	return &Agent{
		client: client,
		logger: l,
		Set:    *metricSet,
		Config: cfg,
	}
}

func (a *Agent) AddCollector(c Collector) {
	a.collectors = append(a.collectors, c)
}

// Run запускает агента
func (a Agent) Run() {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := a.report(); err != nil {
				fmt.Printf("report error: %s\n", err)
			}
			time.Sleep(a.Config.ReportInterval())
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			fmt.Println("start polling")
			for _, collector := range a.collectors {
				snap, err := collector.Collect()
				if err != nil {
					a.logger.Error("collector: %s", err)
					continue
				}
				if len(snap.Counters) != 0 {
					for _, c := range snap.Counters {
						a.UpdateCounter(c.Name, c.Value)
					}
				}
				if len(snap.Gauges) != 0 {
					for _, g := range snap.Gauges {
						a.UpdateGauge(g.Name, g.Value)
					}
				}
			}
			time.Sleep(a.Config.ReportInterval())
		}
	}()
	wg.Wait()
}

func (a Agent) report() error {
	snap := a.GetMetrics()
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

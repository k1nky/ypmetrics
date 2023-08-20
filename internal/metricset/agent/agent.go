package agent

import (
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
	Debug(template string, args ...interface{})
	Info(template string, args ...interface{})
	Error(template string, args ...interface{})
}

type sender interface {
	PushCounter(name string, value int64) (err error)
	PushGauge(name string, value float64) (err error)
}

// Agent представляет собой набор метрик с расширенным функционалом.
// Расширенный функционал позволяет наполнять набор метриками, полученными из сборщиков (Collector),
// и периодически отправлять обновления метрик.
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

// AddCollector добавляет совместимый сборщик для получения метрик.
func (a *Agent) AddCollector(collectors ...Collector) {
	a.collectors = append(a.collectors, collectors...)
}

// Run запускает агента
func (a Agent) Run() {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			a.logger.Debug("sending updates")
			if err := a.report(); err != nil {
				a.logger.Error("report error: %s", err)
			}
			time.Sleep(a.Config.ReportInterval())
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			for _, collector := range a.collectors {
				a.logger.Debug("start polling %T", collector)
				m, err := collector.Collect()
				if err != nil {
					a.logger.Error("collector %T: %s", collector, err)
					continue
				}
				if len(m.Counters) != 0 {
					for _, c := range m.Counters {
						a.UpdateCounter(c.Name, c.Value)
					}
				}
				if len(m.Gauges) != 0 {
					for _, g := range m.Gauges {
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
		if err := a.client.PushCounter(m.Name, m.Value); err != nil {
			return err
		}
	}
	for _, m := range snap.Gauges {
		if err := a.client.PushGauge(m.Name, m.Value); err != nil {
			return err
		}
	}
	return nil
}

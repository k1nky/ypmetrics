package poller

import (
	"context"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type Collector interface {
	Collect() (metric.Metrics, error)
}

type metricStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	Snapshot(*metric.Metrics)
}

type logger interface {
	Debug(template string, args ...interface{})
	Info(template string, args ...interface{})
	Error(template string, args ...interface{})
}

type sender interface {
	PushCounter(name string, value int64) (err error)
	PushGauge(name string, value float64) (err error)
}

// Poller представляет собой набор метрик с расширенным функционалом. Он опрашивает сборщиков (Collector)
// и периодически отправляет обновления метрик в единый набор метрик.
type Poller struct {
	storage    metricStorage
	collectors []Collector
	logger     logger
	client     sender
	Config     config.PollerConfig
}

// New возвращает нового Poller для сбора метрик. По умолчанию в качестве хранилища используется MemStorage.
func New(cfg config.PollerConfig, store metricStorage, log logger, client sender) *Poller {
	return &Poller{
		client:  client,
		logger:  log,
		storage: store,
		Config:  cfg,
	}
}

// AddCollector добавляет совместимый сборщик для получения метрик.
func (a *Poller) AddCollector(collectors ...Collector) {
	a.collectors = append(a.collectors, collectors...)
}

// Run запускает Poller
func (a Poller) Run(ctx context.Context) {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(a.Config.ReportInterval())
		for {
			select {
			case <-ctx.Done():
				a.logger.Debug("stop reporting")
				return
			case <-t.C:
				a.logger.Debug("sending updates")
				if err := a.report(); err != nil {
					a.logger.Error("report error: %s", err)
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(a.Config.PollInterval())
		for {
			select {
			case <-ctx.Done():
				a.logger.Debug("stop polling")
				return
			case <-t.C:
				for _, collector := range a.collectors {
					a.logger.Debug("start polling %T", collector)
					m, err := collector.Collect()
					if err != nil {
						a.logger.Error("collector %T: %s", collector, err)
						continue
					}
					if len(m.Counters) != 0 {
						for _, c := range m.Counters {
							a.storage.UpdateCounter(c.Name, c.Value)
						}
					}
					if len(m.Gauges) != 0 {
						for _, g := range m.Gauges {
							a.storage.UpdateGauge(g.Name, g.Value)
						}
					}
				}
			}
		}
	}()
	wg.Wait()
}

func (a Poller) report() error {
	snap := &metric.Metrics{}
	a.storage.Snapshot(snap)
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

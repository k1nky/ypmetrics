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
	GetCounter(ctx context.Context, name string) *metric.Counter
	GetGauge(ctx context.Context, name string) *metric.Gauge
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	Snapshot(ctx context.Context, metrics *metric.Metrics) error
}

type logger interface {
	Debug(template string, args ...interface{})
	Info(template string, args ...interface{})
	Error(template string, args ...interface{})
}

type sender interface {
	PushCounter(name string, value int64) error
	PushGauge(name string, value float64) error
	PushMetrics(metrics metric.Metrics) error
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
				if err := a.sendReport(); err != nil {
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
				a.poll(ctx)
			}
		}
	}()

	wg.Wait()
}

func (a Poller) poll(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(a.collectors))
	for _, collector := range a.collectors {
		go func(collector Collector) {
			defer wg.Done()
			a.logger.Debug("polling %T", collector)
			m, err := collector.Collect()
			if err != nil {
				a.logger.Error("collector %T: %s", collector, err)
				return
			}
			if len(m.Counters) != 0 {
				for _, c := range m.Counters {
					a.storage.UpdateCounter(ctx, c.Name, c.Value)
				}
			}
			if len(m.Gauges) != 0 {
				for _, g := range m.Gauges {
					a.storage.UpdateGauge(ctx, g.Name, g.Value)
				}
			}
		}(collector)
	}
	wg.Wait()
}

func (a Poller) sendReport() error {
	snapshot := &metric.Metrics{}
	ctx := context.Background()
	if err := a.storage.Snapshot(ctx, snapshot); err != nil {
		return err
	}
	return a.client.PushMetrics(*snapshot)
}

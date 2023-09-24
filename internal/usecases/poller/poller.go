package poller

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

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
		client:     client,
		logger:     log,
		storage:    store,
		Config:     cfg,
		collectors: make([]Collector, 0),
	}
}

func (p *Poller) AddCollector(c ...Collector) {
	p.collectors = append(p.collectors, c...)
}

// Run запускает Poller
func (p Poller) Run(ctx context.Context) {
	p.report(ctx, 3)
	metrics := p.poll(ctx, 3)
	p.storeWorker(ctx, metrics)
	<-ctx.Done()
}

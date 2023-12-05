package poller

import (
	"context"
	"time"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Poller представляет собой набор метрик с расширенным функционалом. Он опрашивает сборщиков (Collector)
// и периодически отправляет обновления метрик в единый набор метрик.
type Poller struct {
	storage    metricStorage
	collectors []Collector
	logger     logger
	client     sender
	Config     config.Poller
}

// Тип для ключа контекста
type contextKey int

const (
	keyWorkerID contextKey = iota
)

const (
	//
	NoLimitToReport = 0
)

// Количество воркеров по умолчанию
const (
	// воркеры сбора метрик
	MaxPollWorkers = 2
	// воркеры отправки метрик
	MaxReportWorkers = 2
)

// New возвращает нового Poller для сбора метрик. По умолчанию в качестве хранилища используется MemStorage.
func New(cfg config.Poller, store metricStorage, log logger, client sender) *Poller {
	return &Poller{
		client:     client,
		logger:     log,
		storage:    store,
		Config:     cfg,
		collectors: make([]Collector, 0),
	}
}

// Добавляет сборщика для опроса
func (p *Poller) AddCollector(c ...Collector) {
	for _, collector := range c {
		if err := collector.Init(); err != nil {
			p.logger.Errorf("failed initializing collector %T: %s", collector, err)
			continue
		}
		p.collectors = append(p.collectors, collector)
	}
}

// Run запускает Poller
func (p Poller) Run(ctx context.Context) {
	// получаем метрики со сборщиков
	metrics := p.poll(ctx, MaxPollWorkers)
	// сохраняем их по мере поступления
	p.storeWorker(ctx, metrics)
	// отправляем метрик на сервер по таймеру
	p.report(ctx, MaxReportWorkers)
}

// Создает и запускает maxWorkers воркеров для опроса сборщиков метрик.
// Собранные метрики передаются через возвращаемый канал по мере поступления.
func (p Poller) poll(ctx context.Context, maxWorkers int) <-chan metric.Metrics {
	// возвращаемый канал с получаемыми метриками от сборщиков
	result := make(chan metric.Metrics, len(p.collectors))
	// задания для воркеров сбора метрик
	jobs := make(chan Collector, len(p.collectors))

	for i := 1; i <= maxWorkers; i++ {
		go func(id int) {
			// запускаем очередной воркер сбора метрик
			// у каждого воркера свой канал, в который отправляются собранные метрики
			ch := p.pollWorker(context.WithValue(ctx, keyWorkerID, id), jobs)
			for {
				select {
				case <-ctx.Done():
					p.logger.Debugf("poll worker #%d: done", id)
					return
				case m, ok := <-ch:
					if !ok {
						return
					}
					// собираем метрики из воркеров в один результирующий канал
					result <- m
				}
			}
		}(i)
	}
	go func() {
		t := time.NewTicker(p.Config.PollInterval())
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				time.Sleep(100 * time.Millisecond)
				close(result)
				return
			case <-t.C:
				// кладем в канал сборщиков, которых должны опросить воркеры
				for _, c := range p.collectors {
					jobs <- c
				}
			}
		}
	}()

	return result
}

// Воркер опроса сборщиков метрик
func (p Poller) pollWorker(ctx context.Context, jobs <-chan Collector) <-chan metric.Metrics {
	result := make(chan metric.Metrics)
	go func() {
		defer close(result)
		id := ctx.Value(keyWorkerID).(int)
		for job := range jobs {
			p.logger.Debugf("poll worker #%d: poll %T", id, job)
			m, err := job.Collect(ctx)
			if err != nil {
				p.logger.Errorf("poll worker #%d: %s", id, err)
			} else {
				result <- m
			}
		}
	}()
	return result
}

// Воркер сохранения метрик из канала
func (p Poller) storeWorker(ctx context.Context, metrics <-chan metric.Metrics) {
	go func() {
		for m := range metrics {
			if err := p.storage.UpdateMetrics(ctx, m); err != nil {
				p.logger.Errorf("store worker: %s", err)
			}
		}
	}()
}

// Создает и запускает maxWorkers воркеров для отправки метрик на сервер.
func (p Poller) report(ctx context.Context, maxWorkers int) {
	metrics := make(chan metric.Metrics, p.Config.RateLimit)

	for i := 1; i <= maxWorkers; i++ {
		// запускаем воркеры для отправки метрик
		p.reportWorker(context.WithValue(ctx, keyWorkerID, i), metrics)
	}
	go func() {
		t := time.NewTicker(p.Config.ReportInterval())
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				close(metrics)
				return
			case <-t.C:
				// делаем снапшот метрик из хранилища, которые будем отправлять
				snapshot := &metric.Metrics{}
				if err := p.storage.Snapshot(ctx, snapshot); err != nil {
					p.logger.Errorf("report: %s", err)
					continue
				}
				if p.Config.RateLimit == NoLimitToReport {
					// ограничения на отправку нет - отправляем все метрики одним большим запросов
					metrics <- *snapshot
					continue
				}
				// ограничение на количество запросов установлен
				// отправляем по одной метрике (иначе не совсем понятно ограничение на отправку)
				for _, m := range snapshot.Counters {
					metrics <- metric.Metrics{
						Counters: []*metric.Counter{m},
					}
				}
				for _, m := range snapshot.Gauges {
					metrics <- metric.Metrics{
						Gauges: []*metric.Gauge{m},
					}
				}
			}
		}
	}()
}

func (p Poller) reportWorker(ctx context.Context, metrics <-chan metric.Metrics) {
	go func() {
		id := ctx.Value(keyWorkerID).(int)
		for {
			select {
			case <-ctx.Done():
				p.logger.Debugf("report worker #%d: done", id)
				return
			case m, ok := <-metrics:
				if !ok {
					return
				}
				if err := p.client.PushMetrics(m); err != nil {
					p.logger.Errorf("report worker #%d: %s", id, err)
				}
			}
		}
	}()
}

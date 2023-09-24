package poller

import (
	"context"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type Collector interface {
	Collect() (metric.Metrics, error)
}

func (p Poller) poll(ctx context.Context, maxWorkers int) chan metric.Metrics {
	result := make(chan metric.Metrics, len(p.collectors))
	jobs := make(chan Collector, len(p.collectors))

	for i := 1; i <= maxWorkers; i++ {
		go p.pollWorker(i, jobs, result)
	}
	go func() {
		<-ctx.Done()
		close(result)
	}()
	go func() {
		t := time.NewTicker(p.Config.PollInterval())
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				for _, c := range p.collectors {
					jobs <- c
				}
			}
		}
	}()

	return result
}

func (p Poller) pollWorker(id int, jobs <-chan Collector, result chan<- metric.Metrics) {
	for {
		job, ok := <-jobs
		if !ok {
			break
		}
		p.logger.Debug("worker #%d starts polling %T", id, job)
		m, err := job.Collect()
		if err != nil {
			p.logger.Error("poll worker #%d: %s", id, err)
		} else {
			result <- m
		}
	}
}

func (p Poller) storeWorker(ctx context.Context, metrics chan metric.Metrics) {
	go func() {
		for {
			m, ok := <-metrics
			if !ok {
				return
			}
			if len(m.Counters) != 0 {
				for _, c := range m.Counters {
					p.storage.UpdateCounter(ctx, c.Name, c.Value)
				}
			}
			if len(m.Gauges) != 0 {
				for _, g := range m.Gauges {
					p.storage.UpdateGauge(ctx, g.Name, g.Value)
				}
			}
		}
	}()
}

func (p Poller) report(ctx context.Context, maxWorkers int) {
	metrics := make(chan metric.Metrics, p.Config.RateLimit)

	for i := 1; i <= maxWorkers; i++ {
		go p.reportWorker(ctx, i, metrics)
	}
	go func() {
		t := time.NewTicker(p.Config.ReportInterval())
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				snapshot := &metric.Metrics{}
				if err := p.storage.Snapshot(ctx, snapshot); err != nil {
					p.logger.Error("report: %s", err)
					continue
				}
				if p.Config.RateLimit == 0 {
					metrics <- *snapshot
					continue
				}
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

func (p Poller) reportWorker(ctx context.Context, id int, metrics <-chan metric.Metrics) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-metrics:
				if !ok {
					return
				}
				if err := p.client.PushMetrics(m); err != nil {
					p.logger.Error("report worker #%d: %s", id, err)
				}
			}
		}
	}()
}

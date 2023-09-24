package poller

import (
	"context"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type CollectWorkerPool struct{}

func (cp CollectWorkerPool) Run(ctx context.Context, maxWorkers int, jobs <-chan Collector) (chan<- metric.Metrics, chan error) {
	result := make(chan metric.Metrics)
	errors := make(chan error)

	for i := 0; i < maxWorkers; i++ {
		go collectWorker(i, jobs, result, errors)
	}
	go func() {
		<-ctx.Done()
		close(result)
		close(errors)
	}()
	return result, errors
}

func collectWorker(id int, jobs <-chan Collector, result chan<- metric.Metrics, errors chan<- error) {
	for {
		job, ok := <-jobs
		if !ok {
			break
		}
		m, err := job.Collect()
		if err != nil {
			result <- m
		} else {
			errors <- err
		}
	}
}

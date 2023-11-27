package poller

import (
	"context"
	"testing"
	"time"

	"github.com/k1nky/ypmetrics/internal/collector"
	"github.com/k1nky/ypmetrics/internal/config"
	log "github.com/k1nky/ypmetrics/internal/logger"
)

func BenchmarkPoll(b *testing.B) {
	// Проверка опроса коллекторов
	tests := []struct {
		name string
		// количество опросов на тест
		times int
		// количество воркеров для опроса сборщиков
		maxWorkers int
		// количество сборщиков, которые будем опрашивать
		numCollectors int
	}{
		{name: "Single worker with 100 collectors", times: 10, maxWorkers: 1, numCollectors: 100},
		{name: "Two workers with 100 collectors", times: 10, maxWorkers: 2, numCollectors: 100},
		{name: "Single worker with 1000 collectors", times: 10, maxWorkers: 1, numCollectors: 1000},
		{name: "Two workers with 1000 collectors", times: 10, maxWorkers: 2, numCollectors: 1000},
	}
	poll := func(p *Poller, maxWorkers int, d time.Duration) int {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		ch := p.poll(ctx, maxWorkers)
		count := 0
		for {
			select {
			case <-ctx.Done():
				return count
			case <-ch:
				count++
			}
		}
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.StopTimer()
			p := New(config.Poller{
				PollIntervalInSec: 1,
			}, nil, &log.Blackhole{}, nil)

			for i := 0; i < tt.numCollectors; i++ {
				p.AddCollector(
					// пока это самый ресурсоемкий сборщик
					&collector.Gops{},
				)
			}
			b.StartTimer()
			count := poll(p, tt.maxWorkers, time.Duration(tt.times*int(time.Second)))
			// дополнительно считаем, сколько смогли опросить сборщиков за один интервал опроса (фактически 1 сек)
			b.ReportMetric(float64(count)/float64(tt.times-int(p.Config.PollIntervalInSec)), "collectors/s")
		})
	}
}

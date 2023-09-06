package storage

import (
	"context"
	"testing"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
)

func TestDBStorageUpdate(t *testing.T) {
	t.SkipNow()
	blackholeLogger := &logger.Blackhole{}
	db := NewDBStorage(blackholeLogger)
	if err := db.Open("postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"); err != nil {
		t.Log(err)
		return
	}
	ctx := context.TODO()
	db.UpdateCounter(ctx, "c0", 1)
	db.UpdateCounter(ctx, "c0", 10)
	m := db.GetCounter(ctx, "c0")
	t.Log(m)
}

func TestDBStorageUpdateMetrics(t *testing.T) {
	// t.SkipNow()
	blackholeLogger := &logger.Blackhole{}
	db := NewDBStorage(blackholeLogger)
	if err := db.Open("postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"); err != nil {
		t.Log(err)
		return
	}
	ctx := context.TODO()
	m := metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("c0", 1), metric.NewCounter("c1", 10)},
		Gauges:   []*metric.Gauge{metric.NewGauge("g0", 2.2)},
	}
	db.UpdateMetrics(ctx, m)
}

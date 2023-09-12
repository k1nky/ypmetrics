package storage

import (
	"context"
	"sync"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// MemStorage хранилище метрик в памяти
type MemStorage struct {
	counters     map[string]*metric.Counter
	gauges       map[string]*metric.Gauge
	countersLock sync.RWMutex
	gaugesLock   sync.RWMutex
}

// NewMemStorage возвращает новое хранилище в памяти
func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]*metric.Counter),
		gauges:   make(map[string]*metric.Gauge),
	}
}

func (ms *MemStorage) Open(cfg Config) error {
	return nil
}

// GetCounter возвращает метрику Counter по имени name.
// Будет возвращен nil, если метрика не найдена
func (ms *MemStorage) GetCounter(ctx context.Context, name string) *metric.Counter {
	ms.countersLock.RLock()
	defer ms.countersLock.RUnlock()

	if m, ok := ms.counters[name]; ok {
		return m
	}
	return nil
}

// GetGauge возвращает метрику Gauge по имени name.
// Будет возвращен nil, если метрика не найдена
func (ms *MemStorage) GetGauge(ctx context.Context, name string) *metric.Gauge {
	ms.gaugesLock.RLock()
	defer ms.gaugesLock.RUnlock()

	if m, ok := ms.gauges[name]; ok {
		return m
	}
	return nil
}

// UpdateCounter сохраняет метрику Counter в хранилище
func (ms *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	c := ms.GetCounter(ctx, name)

	ms.countersLock.Lock()
	defer ms.countersLock.Unlock()

	if c == nil {
		c = metric.NewCounter(name, value)
	} else {
		c.Update(value)
	}

	ms.counters[name] = c
	return nil
}

// UpdateMetrics сохраняет метрики в хранилище
func (ms *MemStorage) UpdateMetrics(ctx context.Context, metrics metric.Metrics) error {
	for _, m := range metrics.Counters {
		ms.UpdateCounter(ctx, m.Name, m.Value)
	}
	for _, m := range metrics.Gauges {
		ms.UpdateGauge(ctx, m.Name, m.Value)
	}
	return nil
}

// UpdateGauge сохраняет метрику Gauge в хранилище
func (ms *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	g := ms.GetGauge(ctx, name)

	ms.gaugesLock.Lock()
	defer ms.gaugesLock.Unlock()

	if g == nil {
		g = metric.NewGauge(name, value)
	} else {
		g.Update(value)
	}

	ms.gauges[name] = g

	return nil
}

// Snapshot создает снимок метрик из хранилища
func (ms *MemStorage) Snapshot(ctx context.Context, snap *metric.Metrics) error {

	if snap == nil {
		return nil
	}

	snap.Counters = make([]*metric.Counter, 0, len(ms.counters))
	snap.Gauges = make([]*metric.Gauge, 0, len(ms.gauges))

	ms.countersLock.Lock()
	defer ms.countersLock.Unlock()
	for _, v := range ms.counters {
		snap.Counters = append(snap.Counters, metric.NewCounter(v.Name, v.Value))
	}

	ms.gaugesLock.Lock()
	defer ms.gaugesLock.Unlock()
	for _, v := range ms.gauges {
		snap.Gauges = append(snap.Gauges, metric.NewGauge(v.Name, v.Value))
	}

	return nil
}

func (ms *MemStorage) Close() error {
	return nil
}

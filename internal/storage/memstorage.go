package storage

import (
	"sync"

	"github.com/k1nky/ypmetrics/internal/metric"
)

// MemStorage хранилище метрик в памяти
type MemStorage struct {
	counters     map[string]*metric.Counter
	gauges       map[string]*metric.Gauge
	countersLock sync.Mutex
	gaugesLock   sync.Mutex
}

// NewMemStorage возвращает новое хранилище в памяти
func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]*metric.Counter),
		gauges:   make(map[string]*metric.Gauge),
	}
}

// GetCounter возвращает метрику Counter по имени name.
// Будет возвращен nil, если метрика не найдена
func (ms *MemStorage) GetCounter(name string) *metric.Counter {
	ms.countersLock.Lock()
	defer ms.countersLock.Unlock()

	if m, ok := ms.counters[name]; ok {
		return m
	}
	return nil
}

// GetGauge возвращает метрику Gauge по имени name.
// Будет возвращен nil, если метрика не найдена
func (ms *MemStorage) GetGauge(name string) *metric.Gauge {
	ms.gaugesLock.Lock()
	defer ms.gaugesLock.Unlock()

	if m, ok := ms.gauges[name]; ok {
		return m
	}
	return nil
}

// SetCounter сохраняет метрику Counter в хранилище
func (ms *MemStorage) SetCounter(m *metric.Counter) {
	if m == nil {
		return
	}

	ms.countersLock.Lock()
	defer ms.countersLock.Unlock()

	ms.counters[m.GetName()] = m
}

// SetGauge сохраняет метрику Gauge в хранилище
func (ms *MemStorage) SetGauge(m *metric.Gauge) {
	if m == nil {
		return
	}

	ms.gaugesLock.Lock()
	defer ms.gaugesLock.Unlock()

	ms.gauges[m.GetName()] = m
}

// Snapshot создает снимок метрик из хранилища
func (ms *MemStorage) Snapshot(snap *metric.Metrics) {

	if snap == nil {
		return
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
}

func (ms *MemStorage) Close() error {
	return nil
}

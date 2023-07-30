package storage

import (
	"sync"

	"github.com/k1nky/ypmetrics/internal/metric"
)

// MemStorage хранилище метрик в памяти
type MemStorage struct {
	values map[string]metric.Measure
	lock   sync.Mutex
}

// Storage хранилище метрик
type Storage interface {
	// Get возвращает метрику по имени. Если запращиваемой метрики нет - возвращает nil
	Get(name string) metric.Measure
	// GetNames возвращает имена всех метрик, имеющихся в хранилище
	GetNames() []string
	// Set добавляет метрику в хранилище. Если метрика с таким именем уже есть, то метрика перезапищется
	Set(value metric.Measure)
	// UpSet добавляет/обновляет метрику в хранилище. Если такая метрика уже есть, то обновляет ее значение
	// новым значением из переданной метрики
	UpSet(metric metric.Measure)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		values: make(map[string]metric.Measure),
	}
}

func (ms *MemStorage) Get(name string) metric.Measure {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	if m, ok := ms.values[name]; ok {
		return m
	}
	return nil
}

func (ms *MemStorage) Set(value metric.Measure) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	ms.values[value.GetName()] = value
}

func (ms *MemStorage) UpSet(metric metric.Measure) {
	if m := ms.Get(metric.GetName()); m != nil {
		m.Update(metric.GetValue())
	} else {
		ms.Set(metric)
	}
}

func (ms *MemStorage) GetNames() []string {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	result := make([]string, 0)
	for name := range ms.values {
		result = append(result, name)
	}
	return result
}

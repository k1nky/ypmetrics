package storage

import "github.com/k1nky/ypmetrics/internal/metric"

type MemStorage struct {
	values map[string]metric.Measure
}

type Storage interface {
	Get(name string) metric.Measure
	GetNames() []string
	Set(value metric.Measure)
	UpSet(metric metric.Measure, value interface{}) error
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		values: make(map[string]metric.Measure),
	}
}

func (ms *MemStorage) Get(name string) metric.Measure {
	if m, ok := ms.values[name]; ok {
		return m
	}
	return nil
}

func (ms *MemStorage) Set(value metric.Measure) {
	ms.values[value.GetName()] = value
}

func (ms *MemStorage) UpSet(metric metric.Measure, value interface{}) error {
	if m, ok := ms.values[metric.GetName()]; ok {
		return m.Update(value)
	} else {
		if err := metric.Update(value); err != nil {
			return err
		}
		ms.Set(metric)
	}
	return nil
}

func (ms *MemStorage) GetNames() []string {
	result := make([]string, 0)
	for name := range ms.values {
		result = append(result, name)
	}
	return result
}

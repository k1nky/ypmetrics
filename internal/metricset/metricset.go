package metricset

import "github.com/k1nky/ypmetrics/internal/metric"

type metricSetStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	SetCounter(*metric.Counter)
	SetGauge(*metric.Gauge)
	Snapshot(*metric.Metrics)
}

// Set представляет собой обобщенный набор метрик, основной вункционал которого
// получение и обновление метрик.
type Set struct {
	storage metricSetStorage
}

// NewSet возвращает новый набор метрик для хранения, которых будет использоваться storage.
func NewSet(storage metricSetStorage) *Set {
	s := &Set{
		storage: storage,
	}

	return s
}

// GetCounter возвращает счетчик с именем name из набора.
// Если счетчик не найден будет возвращен nil.
func (s Set) GetCounter(name string) *metric.Counter {
	return s.storage.GetCounter(name)
}

// GetOrCreateCounter возвращает  из набора счетчик с именем name.
// Если метрика не найдена, то будет возвращена и зарегистрирована новая метрика в наборе.
func (s Set) GetOrCreateCounter(name string) *metric.Counter {
	m := s.GetCounter(name)
	if m != nil {
		return m
	}
	m = metric.NewCounter(name, 0)
	s.storage.SetCounter(m)
	return m
}

// GetOrCreateGauge возвращает из набора метрику типа Gauge с именем name.
// Если метрика не найдена, то будет возвращена и зарегистрирована новая метрика в наборе.
func (s Set) GetOrCreateGauge(name string) *metric.Gauge {
	m := s.GetGauge(name)
	if m != nil {
		return m
	}
	m = metric.NewGauge(name, 0)
	s.storage.SetGauge(m)
	return m
}

// GetCounter возвращает метрику типа Gauge с именем name из набора.
// Если метрика не найдена будет возвращен nil.
func (s Set) GetGauge(name string) *metric.Gauge {
	return s.storage.GetGauge(name)
}

// GetMetrics возвращает список метрик, имеющихся в хранилище сервера
func (s Set) GetMetrics() metric.Metrics {
	snap := &metric.Metrics{}
	s.storage.Snapshot(snap)
	return *snap
}

// UpdateCounter обновляет значение счетчика с именем name новым значением value.
// Если метрика еще не существует в наборе, она будет добавлена.
func (s Set) UpdateCounter(name string, value int64) {
	c := s.GetOrCreateCounter(name)
	c.Update(value)
	s.storage.SetCounter(c)
}

// UpdateGauge обновляет значение метрика типа Gauge с именем name новым значением value.
// Если метрика еще не существует в наборе, она будет добавлена.
func (s Set) UpdateGauge(name string, value float64) {
	g := s.GetOrCreateGauge(name)
	g.Update(value)
	s.storage.SetGauge(g)
}

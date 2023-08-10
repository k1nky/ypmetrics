package metric

type metricStorage interface {
	GetCounter(name string) *Counter
	GetGauge(name string) *Gauge
	SetCounter(*Counter)
	SetGauge(*Gauge)
	Snapshot(*Snapshot)
}

// Set представляет собой набор метрик, формат хранения которых определяется storage.
type Set struct {
	storage metricStorage
}

// Snapshot срез текущих метрик в наборе
type Snapshot struct {
	Counters []*Counter
	Gauges   []*Gauge
}

// NewSet возвращает новый набор метрик для хранения, которых будет использоваться storage.
func NewSet(storage metricStorage) *Set {
	s := &Set{
		storage: storage,
	}

	return s
}

// GetCounter возвращает счетчик с именем name из набора.
// Если счетчик не найден будет возвращен nil.
func (s Set) GetCounter(name string) *Counter {
	return s.storage.GetCounter(name)
}

// GetOrCreateCounter возвращает  из набора счетчик с именем name.
// Если метрика не найдена, то будет возвращена и зарегистрирована новая метрика в наборе.
func (s Set) GetOrCreateCounter(name string) *Counter {
	m := s.GetCounter(name)
	if m != nil {
		return m
	}
	m = NewCounter(name, 0)
	s.storage.SetCounter(m)
	return m
}

// GetOrCreateGauge возвращает из набора метрику типа Gauge с именем name.
// Если метрика не найдена, то будет возвращена и зарегистрирована новая метрика в наборе.
func (s Set) GetOrCreateGauge(name string) *Gauge {
	m := s.GetGauge(name)
	if m != nil {
		return m
	}
	m = NewGauge(name, 0)
	s.storage.SetGauge(m)
	return m
}

// GetCounter возвращает метрику типа Gauge с именем name из набора.
// Если метрика не найдена будет возвращен nil.
func (s Set) GetGauge(name string) *Gauge {
	return s.storage.GetGauge(name)
}

// GetMetrics возвращает список метрик, имеющихся в хранилище сервера
func (s Set) GetMetrics() Snapshot {
	snap := &Snapshot{}
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

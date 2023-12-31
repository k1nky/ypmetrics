// Пакет metric реализует метрики Gauge и Counter.
package metric

import (
	"fmt"
)

type namedMetric struct {
	Name string
}

// Counter метрика "счетчик". При обновлении к старому значению добавляется новое.
type Counter struct {
	namedMetric
	Value int64
}

// Gauge метрика "измеритель". При обновлении новое значение замещает старое.
type Gauge struct {
	namedMetric
	Value float64
}

// Metrics набор метрик.
type Metrics struct {
	// Counters метрики типа Counter.
	Counters []*Counter
	// Gauges метрики типа Gauges.
	Gauges []*Gauge
}

// NewCounter возвращает новый счетчик с именем name и значением initValue.
func NewCounter(name string, initValue int64) *Counter {
	return &Counter{
		namedMetric: namedMetric{Name: name},
		Value:       initValue,
	}
}

// NewGauge возвращает новый "измеритель" с именем name и значением initValue.
func NewGauge(name string, initValue float64) *Gauge {
	return &Gauge{
		namedMetric: namedMetric{Name: name},
		Value:       initValue,
	}
}

// NewMetrics возвращает новый пустой набор метрик.
func NewMetrics() *Metrics {
	return &Metrics{
		Counters: make([]*Counter, 0),
		Gauges:   make([]*Gauge, 0),
	}
}

// GetName возвращает имя метрики.
func (nm namedMetric) GetName() string {
	return nm.Name
}

// String возвращает строковое значение счетчика.
func (c *Counter) String() string {
	return fmt.Sprintf("%d", c.Value)
}

// String возвращает строковое предствление "измерителя".
func (g *Gauge) String() string {
	return fmt.Sprintf("%g", g.Value)
}

// Update обовляет значение счетчика. К старому значению добавляется новое.
func (c *Counter) Update(value int64) {
	c.Value += value
}

// Update обовляет значение измерителя. Старое значение замещается новым
func (g *Gauge) Update(value float64) {
	g.Value = value
}

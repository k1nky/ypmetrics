package handler

// metricType тип метрики
type metricType string

const (
	TypeGauge   = metricType("gauge")
	TypeCounter = metricType("counter")
)

// IsValid возвращает true, если тип метрики имеет допустимое значение.
func (t metricType) IsValid() bool {
	switch t {
	case TypeGauge, TypeCounter:
		return true
	default:
		return false
	}
}

func isValidMetricParams(mtype metricType, name string) bool {
	return mtype.IsValid() && len(name) > 0
}

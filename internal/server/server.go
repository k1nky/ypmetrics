// Пакет server реализует сервер сбора метрик
package server

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

// Server сбора метрик
type Server struct {
	storage storage.Storage
}

// Option опция конфигурации сервера сбора метрик.
// Используются при создании нового сервера через функцию New.
type Option func(*Server)

// WithStorage задает опцию, определяющее какое хранилище использовать для метрик
func WithStorage(storage storage.Storage) Option {
	return func(s *Server) {
		s.storage = storage
	}
}

// New возвращает новый сервер. По умолчанию в качестве хранилища используется MemStorage
func New(options ...Option) *Server {
	s := &Server{}

	for _, opt := range options {
		opt(s)
	}
	if s.storage == nil {
		s.storage = storage.NewMemStorage()
	}

	return s
}

// GetAllMetrics возвращает список метрик, имеющихся в хранилище сервера
func (s Server) GetAllMetrics() []metric.Measure {
	names := s.storage.GetNames()
	metrics := make([]metric.Measure, 0, len(names))
	for _, name := range names {
		metrics = append(metrics, s.storage.Get(name))
	}
	return metrics
}

// GetMetric ищет метрику в хранилище по типу и имени.
// Если тип имеет не верное значение, функция вернет ошибку.
// Если запращиваемой метрики не найдено, то будет возвращен nil.
// Метод провеяет соответствие указанного типа и типа метрики.
func (s Server) GetMetric(typ metric.Type, name string) (metric.Measure, error) {
	if !typ.IsValid() {
		return nil, metric.ErrInvalidType
	}
	m := s.storage.Get(name)
	if m != nil {
		if m.GetType() != typ {
			return nil, nil
		}
	}
	return m, nil
}

// UpdateMetric обновляет значение метрики или добавляет ее в хранилище
// если ее еще нет.
func (s Server) UpdateMetric(metric metric.Measure) {
	s.storage.UpSet(metric)
}

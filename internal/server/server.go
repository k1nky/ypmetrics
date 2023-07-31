package server

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type Server struct {
	storage storage.Storage
}

type Option func(*Server)

// WithStorage возвращает опцию для указания хранилища при создании нового сервера
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

// GetAllMetrics возвращает массив метрик, имеющихся в хранилище сервера
func (s *Server) GetAllMetrics() []metric.Measure {
	metrics := make([]metric.Measure, 0)
	names := s.storage.GetNames()
	for _, name := range names {
		metrics = append(metrics, s.storage.Get(name))
	}
	return metrics
}

// GetMetric ищет метрику в хранилище по типу и имени.
// Если тип имеет не верное значение, функция вернет ошибку.
// Если запращиваемой метрики не найдено, то будет возвращен nil.
// Метод провеяет соответствие указанного типа и типа метрики.
func (s *Server) GetMetric(typ metric.Type, name string) (metric.Measure, error) {
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
func (s *Server) UpdateMetric(metric metric.Measure) {
	s.storage.UpSet(metric)
}

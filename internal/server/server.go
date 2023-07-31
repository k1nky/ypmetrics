package server

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type Server struct {
	storage storage.Storage
}

type Option func(*Server)

func WithStorage(storage storage.Storage) Option {
	return func(s *Server) {
		s.storage = storage
	}
}

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

func (s *Server) UpdateMetric(metric metric.Measure) {
	s.storage.UpSet(metric)
}

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

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

func (s *Server) UpdateMetric(metric metric.Measure, value interface{}) error {
	return s.storage.UpSet(metric, value)
}

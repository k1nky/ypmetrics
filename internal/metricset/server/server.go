package server

import (
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
)

type serverLogger interface {
	Info(template string, args ...interface{})
}

// Server представляет собой набор метрик с расширенным функционалом (но не прямо сейчас).
type Server struct {
	metricset.Set
	logger serverLogger
}

type serverStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	SetCounter(*metric.Counter)
	SetGauge(*metric.Gauge)
	Snapshot(*metricset.Snapshot)
}

func New(storage serverStorage, logger serverLogger) *Server {
	metricSet := metricset.NewSet(storage)
	if metricSet == nil {
		return nil
	}
	return &Server{
		Set:    *metricSet,
		logger: logger,
	}
}

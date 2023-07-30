package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/server"
)

func updateHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		sections := strings.Split(r.URL.Path, "/")
		if len(sections) != 5 {
			if len(sections) == 4 {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		m, err := metric.NewWtihValue(metric.Type(sections[2]), sections[3], sections[4])
		if errors.Is(err, metric.ErrEmptyName) {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		s.UpdateMetric(m)
		w.WriteHeader(http.StatusOK)
	}
}

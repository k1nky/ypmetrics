package main

import (
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
		if len(sections) != 3 {
			if len(sections) == 2 {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		metric, err := metric.New(metric.Type(sections[0]), sections[1])
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if err := s.UpdateMetric(metric, sections[2]); err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

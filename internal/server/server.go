package server

import (
	"net/http"
	"strings"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type Server struct {
	mux     *http.ServeMux
	storage storage.Storage
}

func New(storage storage.Storage) *Server {
	sl := &Server{
		storage: storage,
		mux:     http.NewServeMux(),
	}

	sl.mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(sl.updateMetric)))
	sl.mux.HandleFunc("/", http.NotFound)

	return sl
}

func (sl *Server) Serve() *http.ServeMux {
	return sl.mux
}

func (sl *Server) updateMetric(w http.ResponseWriter, r *http.Request) {
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
	metric, err := metric.New(sections[0], sections[1])
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if err := sl.storage.UpSet(metric, sections[2]); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

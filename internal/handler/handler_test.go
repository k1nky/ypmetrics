package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Update metric",
			request: "/update/counter/counter0/10",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Update metric with unsupported value",
			request: "/update/counter/counter0/10.10",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric without name #1",
			request: "/update/counter//10",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "Update metric without name #2",
			request: "/update/counter/",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "Update metric with invalid value",
			request: "/update/counter/counter0/invalidvalue",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric without value",
			request: "/update/counter/counter0/",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
			},
		},
	}
	server := server.New()
	handler := New(server)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestValueHandler(t *testing.T) {
	type want struct {
		statusCode int
		value      string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Counter",
			request: "/value/counter/counter1",
			want: want{
				statusCode: http.StatusOK,
				value:      "10",
			},
		},
		{
			name:    "Updated counter",
			request: "/value/counter/counter2",
			want: want{
				statusCode: http.StatusOK,
				value:      "17",
			},
		},
		{
			name:    "Gauge",
			request: "/value/gauge/gauge1",
			want: want{
				statusCode: http.StatusOK,
				value:      "10",
			},
		},
		{
			name:    "Updated gauge",
			request: "/value/gauge/gauge2",
			want: want{
				statusCode: http.StatusOK,
				value:      "0.9999",
			},
		},
		{
			name:    "Metric not exists",
			request: "/value/gauge/gauge3",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "",
			},
		},
		{
			name:    "Incompatible type",
			request: "/value/gauge/counter1",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "",
			},
		},
	}
	server := server.New()
	server.UpdateMetric(&metric.Counter{Name: "counter1", Value: 10})
	server.UpdateMetric(&metric.Counter{Name: "counter2", Value: 10})
	server.UpdateMetric(&metric.Counter{Name: "counter2", Value: 7})
	server.UpdateMetric(&metric.Gauge{Name: "gauge1", Value: 10})
	server.UpdateMetric(&metric.Gauge{Name: "gauge2", Value: 10.99})
	server.UpdateMetric(&metric.Gauge{Name: "gauge2", Value: 0.9999})

	handler := New(server)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err, "error while reading body")
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.value, string(body))
		})
	}
}

func TestAllMetricsHandler(t *testing.T) {
	s := server.New()
	s.UpdateMetric(&metric.Counter{Name: "counter1", Value: 10})
	s.UpdateMetric(&metric.Counter{Name: "counter2", Value: 10})
	type want struct {
		statusCode int
		value      string
	}
	tests := []struct {
		name   string
		server *server.Server
		want   want
	}{
		{
			name: "With values",
			want: want{
				statusCode: http.StatusOK,
				value:      "counter1 = 10\ncounter2 = 10\n",
			},
			server: s,
		},
		{
			name: "Updated counter",
			want: want{
				statusCode: http.StatusOK,
				value:      "",
			},
			server: server.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(tt.server)
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err, "error while reading body")
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.value, string(body))
		})
	}
}

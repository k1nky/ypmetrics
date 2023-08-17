package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/metricset/server"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestTypeIsValid(t *testing.T) {
	tests := []struct {
		name string
		tr   typeMetric
		want bool
	}{
		{
			name: "ValidCounter",
			tr:   TypeCounter,
			want: true,
		},
		{
			name: "ValidGauge",
			tr:   TypeGauge,
			want: true,
		},
		{
			name: "InvalidType",
			tr:   "invalidtype",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.IsValid(); got != tt.want {
				t.Errorf("Type.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric without name #2",
			request: "/update/counter/",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
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
	ms := server.New(storage.NewMemStorage(), logger.New())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, r := gin.CreateTestContext(w)
			r.POST("/update/:type/:name/:value", UpdateHandler(*ms))
			c.Request = httptest.NewRequest(http.MethodPost, tt.request, nil)
			r.ServeHTTP(w, c.Request)
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
		{
			name:    "Unsupported type",
			request: "/value/summary/counter1",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "",
			},
		},
	}

	ms := server.New(storage.NewMemStorage(), logger.New())
	ms.UpdateCounter("counter1", 10)
	ms.UpdateCounter("counter2", 10)
	ms.UpdateCounter("counter2", 7)
	ms.UpdateGauge("gauge1", 10)
	ms.UpdateGauge("gauge2", 10.99)
	ms.UpdateGauge("gauge2", 0.9999)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.request, nil)
			gin.SetMode(gin.TestMode)
			r.GET("/value/:type/:name", ValueHandler(*ms))
			r.ServeHTTP(w, c.Request)
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
	ms := server.New(storage.NewMemStorage(), logger.New())
	ms.UpdateCounter("counter1", 10)
	ms.UpdateCounter("counter2", 20)
	type want struct {
		statusCode int
		value      string
	}
	tests := []struct {
		name string
		ms   *server.Server
		want want
	}{
		{
			name: "With values",
			want: want{
				statusCode: http.StatusOK,
				value:      "counter1 = 10\ncounter2 = 20\n",
			},
			ms: ms,
		},
		{
			name: "Without values",
			want: want{
				statusCode: http.StatusOK,
				value:      "",
			},
			ms: server.New(storage.NewMemStorage(), logger.New()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, r := gin.CreateTestContext(w)
			r.GET("/", AllMetricsHandler(*tt.ms))
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			r.ServeHTTP(w, c.Request)

			result := w.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err, "error while reading body")
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.ElementsMatch(t, strings.Split(tt.want.value, "\n"), strings.Split(string(body), "\n"))
		})
	}
}

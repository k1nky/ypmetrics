package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage/mock"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
	"github.com/stretchr/testify/assert"
)

func TestTypeIsValid(t *testing.T) {
	tests := []struct {
		name string
		tr   metricType
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

func TestUpdate(t *testing.T) {
	type want struct {
		statusCode int
		c          *metric.Counter
		g          *metric.Gauge
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "New counter",
			request: "/update/counter/c0/10",
			want: want{
				statusCode: http.StatusOK,
				c:          metric.NewCounter("c0", 10),
			},
		},
		{
			name:    "New gauge",
			request: "/update/gauge/g0/10.10",
			want: want{
				statusCode: http.StatusOK,
				g:          metric.NewGauge("g0", 10.10),
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
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)
	store.EXPECT().GetCounter(gomock.Any(), "c0").Return(metric.NewCounter("c0", 10))
	store.EXPECT().GetGauge(gomock.Any(), "g0").Return(metric.NewGauge("g0", 10.10))
	store.EXPECT().UpdateCounter(gomock.Any(), gomock.Any(), gomock.Any())
	store.EXPECT().UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any())

	keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
	h := New(*keeper)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/update/:type/:name/:value", h.Update())
			c.Request = httptest.NewRequest(http.MethodPost, tt.request, nil)
			r.ServeHTTP(w, c.Request)
			result := w.Result()
			defer result.Body.Close()
			if !assert.Equal(t, tt.want.statusCode, result.StatusCode) {
				return
			}
			if result.StatusCode != http.StatusOK {
				return
			}
			if strings.Contains(tt.request, "/counter/") {
				assert.Equal(t, tt.want.c, store.GetCounter(c.Request.Context(), tt.want.c.Name))
			} else {
				assert.Equal(t, tt.want.g, store.GetGauge(c.Request.Context(), tt.want.g.Name))
			}
		})
	}
}

func TestUpdateJSON(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Update counter",
			request: `{"id": "c0", "type": "counter", "delta": 11}`,
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id": "c0", "type": "counter", "delta": 11}`,
			},
		},
		{
			name:    "Update gauge",
			request: `{"id": "g0", "type": "gauge", "value": 0.1}`,
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id": "g0", "type": "gauge", "value": 0.1}`,
			},
		},
		{
			name:    "Update counter without value",
			request: `{"id": "c0", "type": "counter"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update gauge without value",
			request: `{"id": "g0", "type": "gauge"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric without name",
			request: `{"type": "gauge", "value": 0.1}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric with empty name",
			request: `{"id": "", "type": "gauge", "value": 0.1}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric without type",
			request: `{"id": "gauge0", "value": 0.1}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Update metric with invalid value",
			request: `{"id": "g0", "type": "mygauge", "value": 0.1}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)
	store.EXPECT().GetCounter(gomock.Any(), "c0").Return(metric.NewCounter("c0", 11))
	store.EXPECT().GetGauge(gomock.Any(), "g0").Return(metric.NewGauge("g0", 0.1))
	store.EXPECT().UpdateCounter(gomock.Any(), gomock.Any(), gomock.Any())
	store.EXPECT().UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any())

	keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
	h := New(*keeper)
	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/update/", h.UpdateJSON())
			buf := bytes.NewBufferString(tt.request)
			c.Request = httptest.NewRequest(http.MethodPost, "/update/", buf)
			r.ServeHTTP(w, c.Request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if result.StatusCode != http.StatusOK {
				return
			}
			resp := bytes.Buffer{}
			if _, err := resp.ReadFrom(result.Body); !assert.NoError(t, err, "error while decoding") {
				return
			}
			assert.JSONEq(t, tt.want.body, resp.String())
		})
	}
}

func TestValue(t *testing.T) {
	type want struct {
		statusCode int
		name       string
		value      string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Counter",
			request: "/value/counter/c0",
			want: want{
				statusCode: http.StatusOK,
				name:       "c0",
				value:      "11",
			},
		},
		{
			name:    "Gauge",
			request: "/value/gauge/g0",
			want: want{
				statusCode: http.StatusOK,
				name:       "g0",
				value:      "0.1",
			},
		},
		{
			name:    "Metric not exists",
			request: "/value/gauge/g100",
			want: want{
				statusCode: http.StatusNotFound,
				name:       "g100",
				value:      "",
			},
		},
		{
			name:    "Incompatible type",
			request: "/value/gauge/c1",
			want: want{
				statusCode: http.StatusNotFound,
				name:       "c1",
				value:      "",
			},
		},
		{
			name:    "Unsupported type",
			request: "/value/summary/counter1",
			want: want{
				statusCode: http.StatusBadRequest,
				value:      "",
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			switch tt.want.name {
			case "c0":
				store.EXPECT().GetCounter(gomock.Any(), "c0").Return(metric.NewCounter("c0", 11))
			case "c1":
				store.EXPECT().GetGauge(gomock.Any(), "c1").Return(nil)
			case "g0":
				store.EXPECT().GetGauge(gomock.Any(), "g0").Return(metric.NewGauge("g0", 0.1))
			case "g100":
				store.EXPECT().GetGauge(gomock.Any(), "g100").Return(nil)
			}
			keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
			h := New(*keeper)

			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.request, nil)
			r.GET("/value/:type/:name", h.Value())
			r.ServeHTTP(w, c.Request)
			result := w.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			if !assert.NoError(t, err, "error while reading body") {
				return
			}
			if !assert.Equal(t, tt.want.statusCode, result.StatusCode) {
				return
			}
			assert.Equal(t, tt.want.value, string(body))
		})
	}
}

func TestValueJSON(t *testing.T) {
	type want struct {
		statusCode int
		value      string
		name       string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Counter",
			request: `{"id": "c0", "type": "counter"}`,
			want: want{
				statusCode: http.StatusOK,
				value:      `{"id": "c0", "type": "counter", "delta": 11}`,
				name:       "c0",
			},
		},
		{
			name:    "Gauge",
			request: `{"id": "g0", "type": "gauge"}`,
			want: want{
				statusCode: http.StatusOK,
				value:      `{"id": "g0", "type": "gauge", "value": 0.1}`,
				name:       "g0",
			},
		},
		{
			name:    "Metric not exists",
			request: `{"id": "g100", "type": "gauge"}`,
			want: want{
				statusCode: http.StatusNotFound,
				name:       "g100",
			},
		},
		{
			name:    "Incompatible type",
			request: `{"id": "g1", "type": "counter"}`,
			want: want{
				statusCode: http.StatusNotFound,
				name:       "g1",
			},
		},
		{
			name:    "Unsupported type",
			request: `{"id": "g1", "type": "summary"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mock.NewMockStorage(ctrl)
			switch tt.want.name {
			case "c0":
				store.EXPECT().GetCounter(gomock.Any(), "c0").Return(metric.NewCounter("c0", 11))
			case "g1":
				store.EXPECT().GetCounter(gomock.Any(), "g1").Return(nil)
			case "g0":
				store.EXPECT().GetGauge(gomock.Any(), "g0").Return(metric.NewGauge("g0", 0.1))
			case "g100":
				store.EXPECT().GetGauge(gomock.Any(), "g100").Return(nil)
			}
			keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
			h := New(*keeper)

			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			buf := bytes.NewBufferString(tt.request)
			c.Request = httptest.NewRequest(http.MethodPost, "/value/", buf)
			r.POST("/value/", h.ValueJSON())
			r.ServeHTTP(w, c.Request)

			result := w.Result()
			defer result.Body.Close()
			if !assert.Equal(t, tt.want.statusCode, result.StatusCode) {
				return
			}
			if result.StatusCode != http.StatusOK {
				return
			}
			assert.Contains(t, result.Header.Get("content-type"), "application/json")
			resp := bytes.Buffer{}
			if _, err := resp.ReadFrom(result.Body); !assert.NoError(t, err, "error while decoding") {
				return
			}
			assert.JSONEq(t, tt.want.value, resp.String())
		})
	}
}

func TestAllMetrics(t *testing.T) {
	type want struct {
		statusCode int
		value      string
	}

	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	tests := []struct {
		name string
		ms   metric.Metrics
		want want
	}{
		{
			name: "With values",
			want: want{
				statusCode: http.StatusOK,
				value:      "c1 = 10\ng1 = 10.1\n",
			},
			ms: metric.Metrics{
				Counters: []*metric.Counter{metric.NewCounter("c1", 10)},
				Gauges:   []*metric.Gauge{metric.NewGauge("g1", 10.1)},
			},
		},
		{
			name: "Without values",
			want: want{
				statusCode: http.StatusOK,
				value:      "",
			},
			ms: metric.Metrics{},
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			store.EXPECT().Snapshot(gomock.Any(), gomock.Any()).SetArg(1, tt.ms)
			keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
			h := New(*keeper)
			r.GET("/", h.AllMetrics())
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			r.ServeHTTP(w, c.Request)

			result := w.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			if !assert.NoError(t, err, "error while reading body") {
				return
			}
			if !assert.Equal(t, tt.want.statusCode, result.StatusCode) {
				return
			}
			assert.ElementsMatch(t, strings.Split(tt.want.value, "\n"), strings.Split(string(body), "\n"))
		})
	}
}

func TestUpdatesJSON(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Updates",
			request: `[{"id": "c0", "type": "counter", "delta": 11}, {"id": "g0", "type": "gauge", "value": 1.1}]`,
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}

	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)
	store.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).Return(nil)

	keeper := keeper.New(store, config.KeeperConfig{}, &logger.Blackhole{})
	h := New(*keeper)
	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/updates/", h.UpdatesJSON())
			buf := bytes.NewBufferString(tt.request)
			c.Request = httptest.NewRequest(http.MethodPost, "/updates/", buf)
			r.ServeHTTP(w, c.Request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if result.StatusCode != http.StatusOK {
				return
			}
		})
	}
}

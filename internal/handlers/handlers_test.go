package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
				statusCode: http.StatusBadRequest,
			},
		},
	}
	server := server.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateHandler(server))
			h(w, request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

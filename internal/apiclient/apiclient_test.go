package apiclient

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestClient_PushMetric(t *testing.T) {
	type args struct {
		typ   string
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "",
			args:    args{typ: "counter", name: "c0", value: "10"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				want := fmt.Sprintf("/update/%s/%s/%s", tt.args.typ, tt.args.name, tt.args.value)
				assert.Equal(t, want, req.URL.Path)
				rw.WriteHeader(http.StatusOK)
			}))
			defer httpserver.Close()
			c := &Client{
				EndpointURL: httpserver.URL,
				httpclient:  resty.NewWithClient(httpserver.Client()),
			}
			if err := c.PushMetric(tt.args.typ, tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Client.PushMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_PushCounter(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "",
			args:    args{name: "c1", value: 10},
			want:    `{"id":"c1", "type": "counter", "delta":10}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if !assert.Equal(t, req.Header.Get("content-type"), "application/json") {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				m := bytes.Buffer{}
				if _, err := m.ReadFrom(req.Body); err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				if !assert.JSONEq(t, tt.want, m.String()) {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				rw.WriteHeader(http.StatusOK)
			}))
			defer httpserver.Close()
			c := &Client{
				EndpointURL: httpserver.URL,
				httpclient:  resty.NewWithClient(httpserver.Client()),
			}
			if err := c.PushCounter(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Client.PushCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_PushGauge(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "",
			args:    args{name: "g1", value: 10.1},
			want:    `{"id": "g1", "type":"gauge", "value": 10.1}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if !assert.Equal(t, req.Header.Get("content-type"), "application/json") {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				m := bytes.Buffer{}
				if _, err := m.ReadFrom(req.Body); err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				if !assert.JSONEq(t, tt.want, m.String()) {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				rw.WriteHeader(http.StatusOK)
			}))
			defer httpserver.Close()
			c := &Client{
				EndpointURL: httpserver.URL,
				httpclient:  resty.NewWithClient(httpserver.Client()),
			}
			if err := c.PushGauge(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Client.PushGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "URL with schema",
			args: args{"http://localhost:8080"},
			want: &Client{EndpointURL: "http://localhost:8080"},
		},
		{
			name: "URL without schema",
			args: args{"localhost:8080"},
			want: &Client{EndpointURL: "http://localhost:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.url)
			assert.Equal(t, tt.want.EndpointURL, got.EndpointURL)
		})
	}
}

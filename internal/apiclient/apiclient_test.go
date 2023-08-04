package apiclient

import (
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

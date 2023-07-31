package apiclient

import (
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/server"
)

func TestUpdateMetric(t *testing.T) {
	srv := server.New()
	httpsrv := httptest.NewServer(handler.New(srv))
	defer httpsrv.Close()
	cli := New(WithEndpointURL(httpsrv.URL))
	cli.httpclient = resty.NewWithClient(httpsrv.Client())

	tests := []struct {
		name    string
		metric  metric.Measure
		wantErr bool
	}{
		{
			name: "Valid request",
			metric: &metric.Counter{
				Name:  "counter0",
				Value: 10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cli.UpdateMetric(tt.metric); (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

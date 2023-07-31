package agent

import (
	"testing"
	"time"

	"github.com/k1nky/ypmetrics/internal/apiclient"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name string
		args args
		want *Agent
	}{
		{
			name: "Default options",
			args: args{
				options: make([]Option, 0),
			},
			want: &Agent{
				PollInterval:   DefPollInterval,
				ReportInterval: DefReportInterval,
				storage:        &storage.MemStorage{},
				client:         apiclient.New(),
			},
		},
		{
			name: "With ReportInterval",
			args: args{
				options: []Option{WithReportInterval(20 * time.Second)},
			},
			want: &Agent{
				PollInterval:   DefPollInterval,
				ReportInterval: 20 * time.Second,
				storage:        &storage.MemStorage{},
				client:         apiclient.New(),
			},
		},
		{
			name: "With ReportInterval and PollInterval",
			args: args{
				options: []Option{WithReportInterval(20 * time.Second), WithPollInterval(7 * time.Second)},
			},
			want: &Agent{
				PollInterval:   7 * time.Second,
				ReportInterval: 20 * time.Second,
				storage:        &storage.MemStorage{},
				client:         apiclient.New(),
			},
		},
		{
			name: "With Endpoint URL",
			args: args{
				options: []Option{WithEndpoint("example.org")},
			},
			want: &Agent{
				PollInterval:   DefPollInterval,
				ReportInterval: DefReportInterval,
				storage:        &storage.MemStorage{},
				client:         apiclient.New(apiclient.WithEndpointURL("http://example.org")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.options...)
			assert.Equal(t, tt.want.PollInterval, got.PollInterval)
			assert.Equal(t, tt.want.ReportInterval, got.ReportInterval)
			assert.Equal(t, tt.want.client.EndpointURL, got.client.EndpointURL)
			assert.NotNil(t, got.storage)
		})
	}
}

package agent

import (
	"testing"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestAgentPollRuntime(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		want       *metric.Gauge
	}{
		{
			name:       "Metric Mallocs exists",
			metricName: "Mallocs",
			want:       metric.NewGauge("Mallocs", 0),
		},
		{
			name:       "Metric not exists",
			metricName: "Mallocs123",
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(config.AgentConfig{})
			a.setupPredefinedMetrics()
			a.pollRuntime()
			got := a.metricSet.GetGauge(tt.metricName)
			if got != nil {
				got.Value = 0
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

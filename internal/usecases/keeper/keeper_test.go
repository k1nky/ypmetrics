package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

func TestUniqueCounters(t *testing.T) {
	type args struct {
		counters []*metric.Counter
	}
	tests := []struct {
		name string
		args args
		want []*metric.Counter
	}{
		{
			name: "With duplicates",
			args: args{
				counters: []*metric.Counter{
					metric.NewCounter("c0", 1), metric.NewCounter("c1", 2), metric.NewCounter("c0", 3),
				},
			},
			want: []*metric.Counter{
				metric.NewCounter("c0", 4), metric.NewCounter("c1", 2),
			},
		},
		{
			name: "Without duplicates",
			args: args{
				counters: []*metric.Counter{
					metric.NewCounter("c0", 1), metric.NewCounter("c1", 2),
				},
			},
			want: []*metric.Counter{
				metric.NewCounter("c0", 1), metric.NewCounter("c1", 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueCounters(tt.args.counters)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestUniqueGauges(t *testing.T) {
	type args struct {
		gauges []*metric.Gauge
	}
	tests := []struct {
		name string
		args args
		want []*metric.Gauge
	}{
		{
			name: "With duplicates",
			args: args{
				gauges: []*metric.Gauge{
					metric.NewGauge("c0", 1), metric.NewGauge("c1", 2), metric.NewGauge("c0", 3),
				},
			},
			want: []*metric.Gauge{
				metric.NewGauge("c0", 3), metric.NewGauge("c1", 2),
			},
		},
		{
			name: "Without duplicates",
			args: args{
				gauges: []*metric.Gauge{
					metric.NewGauge("c0", 1), metric.NewGauge("c1", 2),
				},
			},
			want: []*metric.Gauge{
				metric.NewGauge("c0", 1), metric.NewGauge("c1", 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueGauge(tt.args.gauges)
			assert.Equal(t, got, tt.want)
		})
	}
}

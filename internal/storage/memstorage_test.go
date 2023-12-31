package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

func TestMemStorageGetCounter(t *testing.T) {
	type fields struct {
		counters map[string]*metric.Counter
		gauges   map[string]*metric.Gauge
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Counter
	}{
		{
			name: "Existed metric",
			fields: fields{
				counters: map[string]*metric.Counter{"counter0": metric.NewCounter("counter0", 123)},
			},
			args: args{name: "counter0"},
			want: metric.NewCounter("counter0", 123),
		},
		{
			name: "Not existed metric",
			fields: fields{
				counters: make(map[string]*metric.Counter),
				gauges:   map[string]*metric.Gauge{"counter0": metric.NewGauge("counter0", 123)},
			},
			args: args{name: "counter0"},
			want: nil,
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
			}
			got := ms.GetCounter(ctx, tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorageGetGauge(t *testing.T) {
	type fields struct {
		counters map[string]*metric.Counter
		gauges   map[string]*metric.Gauge
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Gauge
	}{
		{
			name: "Existed metric",
			fields: fields{
				gauges: map[string]*metric.Gauge{"gauge0": metric.NewGauge("gauge0", 123.123)},
			},
			args: args{name: "gauge0"},
			want: metric.NewGauge("gauge0", 123.123),
		},
		{
			name: "Not existed metric",
			fields: fields{
				gauges:   make(map[string]*metric.Gauge),
				counters: map[string]*metric.Counter{"gauge0": metric.NewCounter("gauge0", 123)},
			},
			args: args{name: "gauge0"},
			want: nil,
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
			}
			got := ms.GetGauge(ctx, tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorageUpdateCounter(t *testing.T) {
	type fields struct {
		counters map[string]*metric.Counter
		gauges   map[string]*metric.Gauge
	}
	type args struct {
		m *metric.Counter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Counter
	}{
		{
			name: "Update value",
			fields: fields{
				counters: map[string]*metric.Counter{"c0": metric.NewCounter("c0", 10)},
			},
			args: args{m: metric.NewCounter("c0", 20)},
			want: metric.NewCounter("c0", 30),
		},
		{
			name: "New value",
			fields: fields{
				counters: map[string]*metric.Counter{},
			},
			args: args{m: metric.NewCounter("c0", 20)},
			want: metric.NewCounter("c0", 20),
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
			}
			ms.UpdateCounter(ctx, tt.args.m.Name, tt.args.m.Value)
			got := ms.GetCounter(ctx, tt.args.m.Name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorageUpdateGauge(t *testing.T) {
	type fields struct {
		counters map[string]*metric.Counter
		gauges   map[string]*metric.Gauge
	}
	type args struct {
		m *metric.Gauge
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Gauge
	}{
		{
			name: "Update value",
			fields: fields{
				gauges: map[string]*metric.Gauge{"c0": metric.NewGauge("c0", 10)},
			},
			args: args{m: metric.NewGauge("c0", 20)},
			want: metric.NewGauge("c0", 20),
		},
		{
			name: "New value",
			fields: fields{
				gauges: map[string]*metric.Gauge{},
			},
			args: args{m: metric.NewGauge("c0", 20)},
			want: metric.NewGauge("c0", 20),
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
			}
			ms.UpdateGauge(ctx, tt.args.m.Name, tt.args.m.Value)
			got := ms.GetGauge(ctx, tt.args.m.Name)
			assert.Equal(t, tt.args.m, got)
		})
	}
}

func TestMemStorageSnapshot(t *testing.T) {
	type fields struct {
		counters map[string]*metric.Counter
		gauges   map[string]*metric.Gauge
	}
	type args struct {
		snap *metric.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Metrics
	}{
		{
			name: "Snapshot #1",
			fields: fields{
				counters: map[string]*metric.Counter{"c0": metric.NewCounter("c0", 10), "c1": metric.NewCounter("c1", 23)},
				gauges:   map[string]*metric.Gauge{"g1": metric.NewGauge("g1", 99.99)},
			},
			args: args{snap: &metric.Metrics{}},
			want: &metric.Metrics{
				Counters: []*metric.Counter{metric.NewCounter("c0", 10), metric.NewCounter("c1", 23)},
				Gauges:   []*metric.Gauge{metric.NewGauge("g1", 99.99)},
			},
		},
		{
			name: "Snapshot #2",
			fields: fields{
				counters: map[string]*metric.Counter{"c0": metric.NewCounter("c0", 10), "c1": metric.NewCounter("c1", 23)},
				gauges:   map[string]*metric.Gauge{"g1": metric.NewGauge("g1", 99.99)},
			},
			args: args{snap: nil},
			want: nil,
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
			}
			ms.Snapshot(ctx, tt.args.snap)
			if tt.args.snap == nil {
				assert.Equal(t, tt.want, tt.args.snap)
				return
			}
			assert.ElementsMatch(t, tt.want.Counters, tt.args.snap.Counters)
			assert.ElementsMatch(t, tt.want.Gauges, tt.args.snap.Gauges)
		})
	}
}

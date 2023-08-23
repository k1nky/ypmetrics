package metricset

import (
	"testing"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	counter *metric.Counter
	gauge   *metric.Gauge
}

func (ms *mockStorage) GetCounter(name string) *metric.Counter {
	return ms.counter
}
func (ms *mockStorage) GetGauge(name string) *metric.Gauge {
	return ms.gauge
}
func (ms *mockStorage) SetCounter(c *metric.Counter) {
	ms.counter = c
}
func (ms *mockStorage) SetGauge(g *metric.Gauge) {
	ms.gauge = g
}
func (ms *mockStorage) Snapshot(snap *metric.Metrics) {
	snap.Counters = []*metric.Counter{metric.NewCounter("c0", 123), metric.NewCounter("c1", 1)}
}

func TestSetGetOrCreateCounter(t *testing.T) {
	type fields struct {
		storage metricSetStorage
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
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{counter: metric.NewCounter("c0", 123)}},
			args:   args{name: "c0"},
			want:   metric.NewCounter("c0", 123),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "c0"},
			want:   metric.NewCounter("c0", 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Set{
				storage: tt.fields.storage,
			}
			got := s.GetOrCreateCounter(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetGetOrCreateGauge(t *testing.T) {
	type fields struct {
		storage metricSetStorage
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
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{gauge: metric.NewGauge("g0", 123.1)}},
			args:   args{name: "g0"},
			want:   metric.NewGauge("g0", 123.1),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "g0"},
			want:   metric.NewGauge("g0", 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Set{
				storage: tt.fields.storage,
			}
			got := s.GetOrCreateGauge(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetUpdateCounter(t *testing.T) {
	type fields struct {
		storage metricSetStorage
	}
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Counter
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{counter: metric.NewCounter("m1", 100)}},
			args:   args{name: "m1", value: 1000},
			want:   metric.NewCounter("m1", 1100),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "m1", value: 1000},
			want:   metric.NewCounter("m1", 1000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Set{
				storage: tt.fields.storage,
			}
			s.UpdateCounter(tt.args.name, tt.args.value)
			got := s.GetCounter(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetUpdateGauge(t *testing.T) {
	type fields struct {
		storage metricSetStorage
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metric.Gauge
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{gauge: metric.NewGauge("m1", 100)}},
			args:   args{name: "m1", value: 10.1},
			want:   metric.NewGauge("m1", 10.1),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "m1", value: 10.99},
			want:   metric.NewGauge("m1", 10.99),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Set{
				storage: tt.fields.storage,
			}
			s.UpdateGauge(tt.args.name, tt.args.value)
			got := s.storage.GetGauge(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

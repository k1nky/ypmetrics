package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	counter *Counter
	gauge   *Gauge
}

func (ms *mockStorage) GetCounter(name string) *Counter {
	return ms.counter
}
func (ms *mockStorage) GetGauge(name string) *Gauge {
	return ms.gauge
}
func (ms *mockStorage) SetCounter(c *Counter) {
	ms.counter = c
}
func (ms *mockStorage) SetGauge(g *Gauge) {
	ms.gauge = g
}
func (ms *mockStorage) Snapshot(snap *Snapshot) {
	snap.Counters = []*Counter{NewCounter("c0", 123), NewCounter("c1", 1)}
}

func TestSetGetOrCreateCounter(t *testing.T) {
	type fields struct {
		storage metricStorage
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Counter
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{counter: NewCounter("c0", 123)}},
			args:   args{name: "c0"},
			want:   NewCounter("c0", 123),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "c0"},
			want:   NewCounter("c0", 0),
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
		storage metricStorage
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Gauge
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{gauge: NewGauge("g0", 123.1)}},
			args:   args{name: "g0"},
			want:   NewGauge("g0", 123.1),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "g0"},
			want:   NewGauge("g0", 0),
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
		storage metricStorage
	}
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Counter
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{counter: NewCounter("m1", 100)}},
			args:   args{name: "m1", value: 1000},
			want:   NewCounter("m1", 1100),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "m1", value: 1000},
			want:   NewCounter("m1", 1000),
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
		storage metricStorage
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Gauge
	}{
		{
			name:   "Existing metric",
			fields: fields{storage: &mockStorage{gauge: NewGauge("m1", 100)}},
			args:   args{name: "m1", value: 10.1},
			want:   NewGauge("m1", 10.1),
		},
		{
			name:   "New metric",
			fields: fields{storage: &mockStorage{}},
			args:   args{name: "m1", value: 10.99},
			want:   NewGauge("m1", 10.99),
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

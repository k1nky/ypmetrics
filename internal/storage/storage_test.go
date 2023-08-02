package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMemStorageGet(t *testing.T) {
	type fields struct {
		values map[string]metric.Measure
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   metric.Measure
	}{
		{
			name: "Item exists",
			fields: fields{
				values: map[string]metric.Measure{
					"counter0": &metric.Counter{Name: "counter0", Value: 100},
				},
			},
			args: args{
				name: "counter0",
			},
			want: &metric.Counter{Name: "counter0", Value: 100},
		},
		{
			name: "Item not exists",
			fields: fields{
				values: map[string]metric.Measure{},
			},
			args: args{
				name: "counter0",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				values: tt.fields.values,
			}
			if got := ms.Get(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorageUpSet(t *testing.T) {
	type fields struct {
		values map[string]metric.Measure
	}
	type args struct {
		metric metric.Measure
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "New entry",
			fields: fields{
				values: map[string]metric.Measure{},
			},
			args: args{
				metric: &metric.Counter{Name: "counter0", Value: 10},
			},
			want: int64(10),
		},
		{
			name: "Update entry",
			fields: fields{
				values: map[string]metric.Measure{
					"counter0": &metric.Counter{Name: "counter0", Value: 10},
				},
			},
			args: args{
				metric: &metric.Counter{Name: "counter0", Value: 10},
			},
			want: int64(20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				values: tt.fields.values,
				lock:   sync.Mutex{},
			}
			ms.UpSet(tt.args.metric)
			m, ok := ms.values[tt.args.metric.GetName()]
			assert.True(t, ok)
			assert.Equal(t, tt.want, m.GetValue())
		})
	}
}

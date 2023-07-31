package server

import (
	"reflect"
	"testing"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func TestNew(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "Default server",
			args: args{
				options: make([]Option, 0),
			},
			want: &Server{
				storage: storage.NewMemStorage(),
			},
		},
		{
			name: "Server with storage option",
			args: args{
				options: []Option{WithStorage(storage.NewMemStorage())},
			},
			want: &Server{
				storage: storage.NewMemStorage(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerGetMetric(t *testing.T) {
	type args struct {
		typ  metric.Type
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Measure
		wantErr bool
	}{
		{
			name: "Counter",
			args: args{
				typ:  metric.TypeCounter,
				name: "counter1",
			},
			want: &metric.Counter{
				Name:  "counter1",
				Value: 1,
			},
			wantErr: false,
		},
		{
			name: "Not exists",
			args: args{
				typ:  metric.TypeCounter,
				name: "counter2",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Empty name",
			args: args{
				typ:  metric.TypeCounter,
				name: "",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Incompatible type",
			args: args{
				typ:  metric.TypeGauge,
				name: "counter1",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Invalid type",
			args: args{
				typ:  "mytype",
				name: "counter1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	s := New()
	s.UpdateMetric(&metric.Counter{Name: "counter1", Value: 1})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetMetric(tt.args.typ, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

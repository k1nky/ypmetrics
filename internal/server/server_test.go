package server

import (
	"reflect"
	"testing"

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

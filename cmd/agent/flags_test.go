package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_duration_Set(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		d       *duration
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "With suffix",
			d:    new(duration),
			args: args{
				s: "10s",
			},
			want:    time.Second * 10,
			wantErr: false,
		},
		{
			name: "Without suffix",
			d:    new(duration),
			args: args{
				s: "10",
			},
			want:    time.Second * 10,
			wantErr: false,
		},
		{
			name: "Invalid value",
			d:    new(duration),
			args: args{
				s: "10seconds",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("duration.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, time.Duration(*tt.d))
		})
	}
}

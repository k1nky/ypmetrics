package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterUpdate(t *testing.T) {
	type fields struct {
		Name  string
		Value int64
	}
	type args struct {
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Counter
	}{
		{
			name:   "With zero value",
			fields: fields{Name: "counter0", Value: 10},
			args:   args{value: 0},
			want:   &Counter{namedMetric: namedMetric{Name: "counter0"}, Value: 10},
		},
		{
			name:   "With  positive value",
			fields: fields{Name: "counter0", Value: 10},
			args:   args{value: 10},
			want:   &Counter{namedMetric: namedMetric{Name: "counter0"}, Value: 20},
		},
		{
			name:   "With negative value",
			fields: fields{Name: "counter0", Value: 10},
			args:   args{value: -20},
			want:   &Counter{namedMetric: namedMetric{Name: "counter0"}, Value: -10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				namedMetric: namedMetric{Name: tt.fields.Name},
				Value:       tt.fields.Value,
			}
			c.Update(tt.args.value)
			assert.Equal(t, tt.want, c)
		})
	}
}

func TestGaugeUpdate(t *testing.T) {
	type fields struct {
		Name  string
		Value float64
	}
	type args struct {
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Gauge
	}{
		{
			name:   "With zero value",
			fields: fields{Name: "gauge0", Value: 10},
			args:   args{value: 0},
			want:   &Gauge{namedMetric: namedMetric{Name: "gauge0"}, Value: 0},
		},
		{
			name:   "With  positive value",
			fields: fields{Name: "gauge0", Value: 10},
			args:   args{value: 10.99},
			want:   &Gauge{namedMetric: namedMetric{Name: "gauge0"}, Value: 10.99},
		},
		{
			name:   "With negative value",
			fields: fields{Name: "gauge0", Value: 10},
			args:   args{value: -20.11},
			want:   &Gauge{namedMetric: namedMetric{Name: "gauge0"}, Value: -20.11},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Gauge{
				namedMetric: namedMetric{Name: tt.fields.Name},
				Value:       tt.fields.Value,
			}
			c.Update(tt.args.value)
			assert.Equal(t, tt.want, c)
		})
	}
}

func TestCounterString(t *testing.T) {
	type fields struct {
		Name  string
		Value int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Full filled",
			fields: fields{Name: "counter0", Value: 99},
			want:   "99",
		},
		{
			name:   "Without name",
			fields: fields{Name: "", Value: 99},
			want:   "99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Counter{
				namedMetric: namedMetric{Name: tt.fields.Name},
				Value:       tt.fields.Value,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("Counter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeString(t *testing.T) {
	type fields struct {
		Name  string
		Value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Full filled",
			fields: fields{Name: "gauge0", Value: 99.99},
			want:   "99.99",
		},
		{
			name:   "Without name",
			fields: fields{Name: "", Value: 99.11},
			want:   "99.11",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Gauge{
				namedMetric: namedMetric{Name: tt.fields.Name},
				Value:       tt.fields.Value,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("Counter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

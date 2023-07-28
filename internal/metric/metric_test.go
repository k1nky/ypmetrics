package metric

import (
	"reflect"
	"testing"
)

func TestType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tr   Type
		want bool
	}{
		{
			name: "ValidCounter",
			tr:   TypeCounter,
			want: true,
		},
		{
			name: "ValidGauge",
			tr:   TypeGauge,
			want: true,
		},
		{
			name: "InvalidType",
			tr:   "invalidtype",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.IsValid(); got != tt.want {
				t.Errorf("Type.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		typ  string
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    Measure
		wantErr bool
	}{
		{
			name: "ValidCounter",
			args: args{
				typ:  TypeCounter,
				name: "counter0",
			},
			want: &Counter{
				Name: "counter0",
			},
			wantErr: false,
		},
		{
			name: "ValidGauge",
			args: args{
				typ:  TypeGauge,
				name: "gauge0",
			},
			want: &Gauge{
				Name: "gauge0",
			},
			wantErr: false,
		},
		{
			name: "InvalidType",
			args: args{
				typ:  "NotCounter",
				name: "notcounter",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.typ, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeUpdateError(t *testing.T) {
	type fields struct {
		Name  string
		Value float64
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "String float value",
			fields: fields{
				Name:  "",
				Value: 0,
			},
			args: args{
				value: "100.1",
			},
			want:    100.1,
			wantErr: false,
		},
		{
			name: "Float value",
			fields: fields{
				Name:  "",
				Value: 0,
			},
			args: args{
				value: 100.1,
			},
			want:    100.1,
			wantErr: false,
		},
		{
			name: "String int value",
			fields: fields{
				Name:  "",
				Value: 0,
			},
			args: args{
				value: "100",
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "Int value",
			fields: fields{
				Name:  "",
				Value: 50,
			},
			args: args{
				value: 100,
			},
			want:    50,
			wantErr: true,
		},
		{
			name: "String invalid value",
			fields: fields{
				Name:  "",
				Value: 100,
			},
			args: args{
				value: "invalidvalue",
			},
			want:    100,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			if err := g.Update(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Gauge.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if g.Value != tt.want {
				t.Errorf("Gauge.Update() value = %f, want %f", g.Value, tt.want)
			}
		})
	}
}

func TestCounterUpdate(t *testing.T) {
	type fields struct {
		Name  string
		Value int64
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "String value",
			fields: fields{
				Name:  "",
				Value: 0,
			},
			args: args{
				value: "15",
			},
			want:    15,
			wantErr: false,
		},
		{
			name: "Int value",
			fields: fields{
				Name:  "",
				Value: 50,
			},
			args: args{
				value: 100,
			},
			want:    50,
			wantErr: true,
		},
		{
			name: "Int64 value",
			fields: fields{
				Name:  "",
				Value: 50,
			},
			args: args{
				value: int64(100),
			},
			want:    150,
			wantErr: false,
		},
		{
			name: "String invalid value",
			fields: fields{
				Name:  "",
				Value: 100,
			},
			args: args{
				value: "10.1",
			},
			want:    100,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			if err := c.Update(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Counter.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if c.Value != tt.want {
				t.Errorf("Counter.Update() value = %d, want %d", c.Value, tt.want)
			}
		})
	}
}

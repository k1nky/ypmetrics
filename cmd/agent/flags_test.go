package main

import (
	"os"
	"testing"
	"time"

	"github.com/k1nky/ypmetrics/internal/agent"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		osargs  []string
		env     map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name:   "Default",
			osargs: []string{"agent"},
			env:    map[string]string{},
			want: &Config{
				Address:        "localhost:8080",
				ReportInterval: agent.DefReportInterval,
				PollInterval:   agent.DefPollInterval,
			},
			wantErr: false,
		},
		{
			name:   "With argument",
			osargs: []string{"server", "-a", ":8090", "-r", "30", "-p", "10"},
			env:    map[string]string{},
			want: &Config{
				Address:        "localhost:8090",
				ReportInterval: 30 * time.Second,
				PollInterval:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name:   "With environment variable",
			osargs: []string{"server"},
			env: map[string]string{
				"ADDRESS":         "127.0.0.1:9000",
				"REPORT_INTERVAL": "30",
				"POLL_INTERVAL":   "10",
			},
			want: &Config{
				Address:        "127.0.0.1:9000",
				ReportInterval: 30 * time.Second,
				PollInterval:   10 * time.Second,
			},
			wantErr: false,
		},
		{
			name:   "With argument and environment variable",
			osargs: []string{"server", "-a", ":8090"},
			env:    map[string]string{"ADDRESS": "127.0.0.1:9000"},
			want: &Config{
				Address:        "127.0.0.1:9000",
				ReportInterval: agent.DefReportInterval,
				PollInterval:   agent.DefPollInterval,
			},
			wantErr: false,
		},
		{
			name:    "With invalid argument",
			osargs:  []string{"server", "-t"},
			env:     map[string]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "With invalid argument value",
			osargs:  []string{"server", "-a", "127.0.0.1/8000"},
			env:     map[string]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable value",
			osargs:  []string{"server"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable and argument value",
			osargs:  []string{"server", "-a", "127.0.0.2/8000"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var (
				err error
				c   *Config
			)
			if c, err = Parse(flag.NewFlagSet("agent", flag.ContinueOnError)); (err != nil) != tt.wantErr {
				t.Errorf("Config.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, c)
		})
	}
}

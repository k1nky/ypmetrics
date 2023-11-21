package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		osargs  []string
		env     map[string]string
		want    PollerConfig
		wantErr bool
	}{
		{
			name:   "Default",
			osargs: []string{"agent"},
			env:    map[string]string{},
			want: PollerConfig{
				Address:             "localhost:8080",
				ReportIntervalInSec: DefaultPollerReportIntervalInSec,
				PollIntervalInSec:   DefaultPollerPollIntervalInSec,
				LogLevel:            "info",
				RateLimit:           0,
			},
			wantErr: false,
		},
		{
			name:   "With argument",
			osargs: []string{"server", "-a", ":8090", "-r", "30", "-p", "10", "--log-level", "error", "-k", "secret", "-l", "12"},
			env:    map[string]string{},
			want: PollerConfig{
				Address:             "localhost:8090",
				ReportIntervalInSec: 30,
				PollIntervalInSec:   10,
				LogLevel:            "error",
				Key:                 "secret",
				RateLimit:           12,
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
				"LOG_LEVEL":       "debug",
				"KEY":             "secret",
				"RATE_LIMIT":      "12",
			},
			want: PollerConfig{
				Address:             "127.0.0.1:9000",
				ReportIntervalInSec: 30,
				PollIntervalInSec:   10,
				LogLevel:            "debug",
				Key:                 "secret",
				RateLimit:           12,
			},
			wantErr: false,
		},
		{
			name:   "With argument and environment variable",
			osargs: []string{"server", "-a", ":8090"},
			env:    map[string]string{"ADDRESS": "127.0.0.1:9000"},
			want: PollerConfig{
				Address:             "127.0.0.1:9000",
				ReportIntervalInSec: DefaultReportIntervalInSec,
				PollIntervalInSec:   DefaultPollIntervalInSec,
				LogLevel:            "info",
			},
			wantErr: false,
		},
		{
			name:    "With invalid argument",
			osargs:  []string{"server", "-t"},
			env:     map[string]string{},
			want:    PollerConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid argument value",
			osargs:  []string{"server", "-a", "127.0.0.1/8000"},
			env:     map[string]string{},
			want:    PollerConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable value",
			osargs:  []string{"server"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    PollerConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable and argument value",
			osargs:  []string{"server", "-a", "127.0.0.2/8000"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    PollerConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := PollerConfig{}
			if err := ParsePollerConfig(&c); err != nil {
				if (err != nil) != tt.wantErr {
					t.Errorf("Config.Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, c)
		})
	}
}

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		osargs    []string
		env       map[string]string
		jsonValue []byte
		want      Poller
		wantErr   bool
	}{
		{
			name:      "Default",
			osargs:    []string{"agent"},
			env:       map[string]string{},
			jsonValue: nil,
			want: Poller{
				Address:             "localhost:8080",
				ReportIntervalInSec: DefaultPollerReportIntervalInSec,
				PollIntervalInSec:   DefaultPollerPollIntervalInSec,
				LogLevel:            "info",
				RateLimit:           0,
			},
			wantErr: false,
		},
		{
			name:      "Only arguments",
			osargs:    []string{"server", "-a", ":8090", "-r", "30", "-p", "10", "--log-level", "error", "-k", "secret", "-l", "12"},
			env:       map[string]string{},
			jsonValue: nil,
			want: Poller{
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
			name:      "Only environment variables",
			osargs:    []string{"server"},
			jsonValue: nil,
			env: map[string]string{
				"ADDRESS":         "127.0.0.1:9000",
				"REPORT_INTERVAL": "30",
				"POLL_INTERVAL":   "10",
				"LOG_LEVEL":       "debug",
				"KEY":             "secret",
				"RATE_LIMIT":      "12",
			},
			want: Poller{
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
			name:   "Only JSON",
			osargs: []string{"server"},
			jsonValue: []byte(`
				{
					"address": "127.0.0.1:9000",
					"crypto_key":"key.pem",
					"report_interval":30,
					"poll_interval":10,
					"log_level":"debug",
					"key":"secret",
					"rate_limit": 2
				}
			`),
			env: map[string]string{},
			want: Poller{
				Address:             "127.0.0.1:9000",
				CryptoKey:           "key.pem",
				ReportIntervalInSec: 30,
				PollIntervalInSec:   10,
				LogLevel:            "debug",
				Key:                 "secret",
				RateLimit:           2,
			},
			wantErr: false,
		},
		{
			name:      "Parse priority",
			osargs:    []string{"server", "-a", ":8090", "-p", "100"},
			env:       map[string]string{"ADDRESS": "127.0.0.1:9000", "CRYPTO_KEY": "key.pem"},
			jsonValue: []byte(`{"address":"1.1.1.1:80", "key":"key"}`),
			want: Poller{
				Address:             "127.0.0.1:9000",
				ReportIntervalInSec: DefaultPollerReportIntervalInSec,
				PollIntervalInSec:   100,
				LogLevel:            "info",
				Key:                 "key",
				CryptoKey:           "key.pem",
			},
			wantErr: false,
		},
		{
			name:      "With invalid argument",
			osargs:    []string{"server", "-t"},
			env:       map[string]string{},
			jsonValue: nil,
			want:      Poller{},
			wantErr:   true,
		},
		{
			name:      "With invalid argument value",
			osargs:    []string{"server", "-a", "127.0.0.1/8000"},
			env:       map[string]string{},
			jsonValue: nil,
			want:      Poller{},
			wantErr:   true,
		},
		{
			name:      "With invalid evironment variable value",
			osargs:    []string{"server"},
			env:       map[string]string{"ADDRESS": "127.0.0.1/8000"},
			jsonValue: nil,
			want:      Poller{},
			wantErr:   true,
		},
		{
			name:      "With invalid evironment variable and argument value",
			osargs:    []string{"server", "-a", "127.0.0.2/8000"},
			env:       map[string]string{"ADDRESS": "127.0.0.1/8000"},
			jsonValue: nil,
			want:      Poller{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := DefaultPollerConfig
			if err := ParsePollerConfig(&c, tt.jsonValue); err != nil {
				if (err != nil) != tt.wantErr {
					t.Errorf("Config.Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, c)
		})
	}
}

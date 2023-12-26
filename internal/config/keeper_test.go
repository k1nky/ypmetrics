package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name      string
		osargs    []string
		env       map[string]string
		jsonValue []byte
		want      Keeper
		wantErr   bool
	}{
		{
			name:      "Default",
			osargs:    []string{"server"},
			env:       map[string]string{},
			jsonValue: nil,
			want: Keeper{
				Address:            "localhost:8080",
				StoreIntervalInSec: DefaultKeeperStoreIntervalInSec,
				FileStoragePath:    "",
				Restore:            true,
				DatabaseDSN:        "",
				LogLevel:           "info",
				Key:                "",
			},
			wantErr: false,
		},
		{
			name:      "Only arguments",
			osargs:    []string{"server", "-a", ":8090", "-i", "10", "-r", "false", "-f", "/tmp/123", "-d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable", "--log-level", "error", "-k", "mysecret"},
			env:       map[string]string{},
			jsonValue: nil,
			want: Keeper{
				Address:            "localhost:8090",
				StoreIntervalInSec: 10,
				FileStoragePath:    "/tmp/123",
				Restore:            false,
				DatabaseDSN:        "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
				LogLevel:           "error",
				Key:                "mysecret",
			},
			wantErr: false,
		},
		{
			name:   "Only environment",
			osargs: []string{"server"},
			env: map[string]string{
				"ADDRESS":           "127.0.0.1:9000",
				"STORE_INTERVAL":    "99",
				"FILE_STORAGE_PATH": "/tmp/321",
				"RESTORE":           "false",
				"DATABASE_DSN":      "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
				"LOG_LEVEL":         "debug",
				"KEY":               "mysecret",
			},
			jsonValue: nil,
			want: Keeper{
				Address:            "127.0.0.1:9000",
				StoreIntervalInSec: 99,
				FileStoragePath:    "/tmp/321",
				Restore:            false,
				DatabaseDSN:        "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
				LogLevel:           "debug",
				Key:                "mysecret",
			},
			wantErr: false,
		},
		{
			name:   "Only JSON",
			osargs: []string{"server"},
			env:    map[string]string{},
			jsonValue: []byte(`
				{
					"address": "127.0.0.1:9000",
					"store_interval": 99,
					"store_file": "/tmp/321",
					"restore": false,
					"database_dsn": "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
					"log_level": "debug",
					"key": "mysecret"
				}
			`),
			want: Keeper{
				Address:            "127.0.0.1:9000",
				StoreIntervalInSec: 99,
				FileStoragePath:    "/tmp/321",
				Restore:            false,
				DatabaseDSN:        "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
				LogLevel:           "debug",
				Key:                "mysecret",
			},
			wantErr: false,
		},
		{
			name:   "Priority",
			osargs: []string{"server", "-a", ":8090", "-i", "11", "-d", "postgres://localhost:6432/praktikum"},
			env:    map[string]string{"ADDRESS": "127.0.0.1:9000", "RESTORE": "true", "DATABASE_DSN": "postgres://localhost:5432/praktikum"},
			jsonValue: []byte(`
				{
					"address": "127.0.0.1:9002",
					"store_interval": 99,
					"key": "mysecret"
				}
			`),
			want: Keeper{
				Address:            "127.0.0.1:9000",
				StoreIntervalInSec: 11,
				FileStoragePath:    "",
				Restore:            true,
				DatabaseDSN:        "postgres://localhost:5432/praktikum",
				LogLevel:           "info",
				Key:                "mysecret",
			},
			wantErr: false,
		},
		{
			name:      "With invalid argument",
			osargs:    []string{"server", "-t"},
			env:       map[string]string{},
			jsonValue: nil,
			want:      Keeper{},
			wantErr:   true,
		},
		{
			name:      "With invalid argument value",
			osargs:    []string{"server", "-a", "127.0.0.1/8000"},
			env:       map[string]string{},
			jsonValue: nil,
			want:      Keeper{},
			wantErr:   true,
		},
		{
			name:      "With invalid evironment variable value",
			osargs:    []string{"server"},
			env:       map[string]string{"ADDRESS": "127.0.0.1/8000"},
			jsonValue: nil,
			want:      Keeper{},
			wantErr:   true,
		},
		{
			name:   "With invalid JSON",
			osargs: []string{"server"},
			env:    map[string]string{},
			jsonValue: []byte(`
				{
					"address": "127.0.0.1:9002",
					"store_interval": 
					"key": "mysecret"
				}
			`),
			want:    Keeper{},
			wantErr: true,
		},
		{
			name:      "With invalid evironment variable and argument value",
			osargs:    []string{"server", "-a", "127.0.0.2/8000"},
			env:       map[string]string{"ADDRESS": "127.0.0.1/8000"},
			jsonValue: nil,
			want:      Keeper{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := DefaultKeeperConfig
			if err := ParseKeeperConfig(&c, tt.jsonValue); err != nil {
				if (err != nil) != tt.wantErr {
					t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, c)
		})
	}
}

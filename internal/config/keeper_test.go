package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		osargs  []string
		env     map[string]string
		want    KeeperConfig
		wantErr bool
	}{
		{
			name:   "Default",
			osargs: []string{"server"},
			env:    map[string]string{},
			want: KeeperConfig{
				Address:            "localhost:8080",
				StoreIntervalInSec: DefStoreIntervalInSec,
				FileStoragePath:    DefFileStoragePath,
				Restore:            true,
				DatabaseDSN:        "",
			},
			wantErr: false,
		},
		{
			name:   "With argument",
			osargs: []string{"server", "-a", ":8090", "-i", "10", "-r", "false", "-f", "/tmp/123", "-d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"},
			env:    map[string]string{},
			want: KeeperConfig{
				Address:            "localhost:8090",
				StoreIntervalInSec: 10,
				FileStoragePath:    "/tmp/123",
				Restore:            false,
				DatabaseDSN:        "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
			},
			wantErr: false,
		},
		{
			name:   "With environment variable",
			osargs: []string{"server"},
			env: map[string]string{
				"ADDRESS":           "127.0.0.1:9000",
				"STORE_INTERVAL":    "99",
				"FILE_STORAGE_PATH": "/tmp/321",
				"RESTORE":           "false",
				"DATABASE_DSN":      "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
			},
			want: KeeperConfig{
				Address:            "127.0.0.1:9000",
				StoreIntervalInSec: 99,
				FileStoragePath:    "/tmp/321",
				Restore:            false,
				DatabaseDSN:        "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
			},
			wantErr: false,
		},
		{
			name:   "With argument and environment variable",
			osargs: []string{"server", "-a", ":8090", "-i", "11", "-d", "postgres://localhost:6432/praktikum"},
			env:    map[string]string{"ADDRESS": "127.0.0.1:9000", "RESTORE": "true", "DATABASE_DSN": "postgres://localhost:5432/praktikum"},
			want: KeeperConfig{
				Address:            "127.0.0.1:9000",
				StoreIntervalInSec: 11,
				FileStoragePath:    DefFileStoragePath,
				Restore:            true,
				DatabaseDSN:        "postgres://localhost:5432/praktikum",
			},
			wantErr: false,
		},
		{
			name:    "With invalid argument",
			osargs:  []string{"server", "-t"},
			env:     map[string]string{},
			want:    KeeperConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid argument value",
			osargs:  []string{"server", "-a", "127.0.0.1/8000"},
			env:     map[string]string{},
			want:    KeeperConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable value",
			osargs:  []string{"server"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    KeeperConfig{},
			wantErr: true,
		},
		{
			name:    "With invalid evironment variable and argument value",
			osargs:  []string{"server", "-a", "127.0.0.2/8000"},
			env:     map[string]string{"ADDRESS": "127.0.0.1/8000"},
			want:    KeeperConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			c := KeeperConfig{}
			if err := ParseKeeperConfig(&c); err != nil {
				if (err != nil) != tt.wantErr {
					t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, c)
		})
	}
}

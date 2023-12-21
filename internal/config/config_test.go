// Пакет config представляет инструменты для работы с конфигурациями сервера и агента.
// Конфигурации могут быть разобраны как из аргументов командной строки, так и из переменных окружения.
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name   string
		osargs []string
		env    map[string]string
		want   string
	}{
		{
			name:   "Unset",
			osargs: []string{"agent"},
			env:    map[string]string{},
			want:   "",
		},
		{
			name:   "Argument",
			osargs: []string{"agent", "-c", "config.json"},
			env:    map[string]string{},
			want:   "config.json",
		},
		{
			name:   "Environment",
			osargs: []string{"agent"},
			env:    map[string]string{"CONFIG": "config.json"},
			want:   "config.json",
		},
		{
			name:   "Priority",
			osargs: []string{"agent", "-c", "config1.json"},
			env:    map[string]string{"CONFIG": "config2.json"},
			want:   "config2.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osargs
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got := GetConfigPath()
			assert.Equal(t, tt.want, got)
		})
	}
}

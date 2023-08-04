package config

import (
	"os"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// ServerConfig конфигурация сервера
type ServerConfig struct {
	// Адрес и порт, который будет слушать сервер
	Address NetAddress `env:"ADDRESS"`
}

func parseServerConfigFromCmd(c *ServerConfig) error {
	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	address := NetAddress("localhost:8080")
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}
	*c = ServerConfig{
		Address: address,
	}
	return nil
}

func parseServerConfigFromEnv(c *ServerConfig) error {
	if err := env.Parse(c); err != nil {
		return err
	}
	if len(c.Address) != 0 {
		if err := c.Address.Set(c.Address.String()); err != nil {
			return err
		}
	}
	return nil
}

// ParseServerConfig разбирает настройки сервера из аргументов командной строки
// и переменных окружения. Переменные окружения имеют более высокий
// приоритет, чем аргументы.
func ParseServerConfig(c *ServerConfig) error {

	if err := parseServerConfigFromCmd(c); err != nil {
		return err
	}
	if err := parseServerConfigFromEnv(c); err != nil {
		return err
	}
	return nil
}
package main

import (
	"net"
	"os"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// netAddress строка вида [<хост>]:<порт> и реализует интерфейс pflag.Value
type netAddress string

type Config struct {
	Address netAddress
}

func (a netAddress) String() string {
	return string(a)
}

func (a *netAddress) Set(s string) error {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if len(host) == 0 {
		// если не указан хост, то используем localhost по умолчанию
		s = "localhost:" + port
	}
	*a = netAddress(s)
	return nil
}

func (a *netAddress) Type() string {
	return "string"
}

func parseFromCmd(cmd *flag.FlagSet) (*Config, error) {
	if cmd == nil {
		cmd = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	address := netAddress("localhost:8080")
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	return &Config{
		Address: address,
	}, nil
}

func parseFromEnv() (*Config, error) {
	type cfg struct {
		Address netAddress `env:"ADDRESS"`
	}
	c := &cfg{}
	if err := env.Parse(c); err != nil {
		return nil, err
	}
	if len(c.Address) != 0 {
		if err := c.Address.Set(c.Address.String()); err != nil {
			return nil, err
		}
	}
	return &Config{
		Address: c.Address,
	}, nil
}

// Parse разбирает настройки сервера из аргументов командной строки
// и переменных окружения. Переменные окружения имеют более высокий
// приоритет, чем аргументы.
// CommandLine по умолчанию из пакета pflag не используем, т.к.
// он усложняет тестирование. Поэтому передаем ссылку на новую CommandLine
// в аргументе cmd метода.
func Parse(cmd *flag.FlagSet) (*Config, error) {
	configFromCmd, err := parseFromCmd(cmd)
	if err != nil {
		return nil, err
	}
	configFromEnv, err := parseFromEnv()
	if err != nil {
		return nil, err
	}
	if len(configFromEnv.Address) == 0 {
		configFromEnv.Address = configFromCmd.Address
	}
	return configFromEnv, nil
}

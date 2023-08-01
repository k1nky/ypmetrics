package main

import (
	"net"
	"os"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// NetAddress строка вида [<хост>]:<порт> и реализует интерфейс pflag.Value
type NetAddress string

type Config struct {
	Address NetAddress `env:"ADDRESS"`
}

func (a NetAddress) String() string {
	return string(a)
}

func (a *NetAddress) Set(s string) error {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if len(host) == 0 {
		// если не указан хост, то используем localhost по умолчанию
		s = "localhost:" + port
	}
	*a = NetAddress(s)
	return nil
}

func (a *NetAddress) Type() string {
	return "string"
}

// Parse разбирает настройки сервера из аргументов командной строки
// и переменных окружения. Переменные окружения имеют более высокий
// приоритет, чем аргументы.
// CommandLine по умолчанию из пакета pflag не используем, т.к.
// он усложняет тестирование. Поэтому передаем ссылку на новую CommandLine
// в аргументе cmd метода.
func (c *Config) Parse(cmd *flag.FlagSet) error {

	address := NetAddress("localhost:8080")

	if cmd == nil {
		cmd = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}
	if err := env.Parse(c); err != nil {
		return err
	}
	if len(c.Address) != 0 {
		if err := address.Set(string(c.Address)); err != nil {
			c.Address = ""
			return err
		}
	}
	c.Address = address
	return nil
}

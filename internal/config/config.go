// Пакет config представляет инструменты для работы с конфигурациями сервера и агента.
// Конфигурации могут быть разобраны как из аргументов командной строки, так и из переменных окружения.
package config

import (
	"errors"
	"net"
	"os"

	flag "github.com/spf13/pflag"
)

// NetAddress строка вида [<хост>]:<порт>.
// Данный тип реализует интерфейс pflag.Value.
type NetAddress string

type Subnet string

// Возвращает строковое представление сетевого адреса.
func (a NetAddress) String() string {
	return string(a)
}

func (s Subnet) String() string {
	return string(s)
}

func (s Subnet) ToIPNet() (*net.IPNet, error) {
	if len(s) == 0 {
		return nil, nil
	}
	_, network, err := net.ParseCIDR(s.String())
	return network, err
}

// Задает значение для сетевого адреса из строки.
// Возвращает ошибку, если строка не соответсвует формату [<хост>]:<порт>.
// Если хост не указан, например, ":8080", то в качестве хоста будет подставлен localhost.
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

func (s *Subnet) Set(value string) error {
	if len(value) != 0 {
		_, _, err := net.ParseCIDR(value)
		if err != nil {
			return err
		}
	}
	*s = Subnet(value)
	return nil
}

func (a *NetAddress) Type() string {
	return "string"
}

func (s *Subnet) Type() string {
	return "string"
}

// GetConfigPath возвращает путь к конфигурационному файлу, указанному либо в
// переменной окружения CONFIG или как значение аргумента --config/-config/-c.
// Вернет пустую строку, если значение не было указано.
func GetConfigPath() string {
	if configPath := os.Getenv("CONFIG"); len(configPath) != 0 {
		return configPath
	}
	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cmd.Usage = func() {}
	configPath := cmd.StringP("config", "c", "", "")
	if err := cmd.Parse(os.Args[1:]); err != nil {
		return ""
	}
	return *configPath
}

// IsHelpWanted вернет true, если ошибка связана с запросом о помощи.
func IsHelpWanted(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}

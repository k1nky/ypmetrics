// Пакет config представляет инструменты для работы с конфигурациями сервера и агента
package config

import "net"

// NetAddress строка вида [<хост>]:<порт> и реализует интерфейс pflag.Value
type NetAddress string

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

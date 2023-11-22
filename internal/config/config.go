// Пакет config представляет инструменты для работы с конфигурациями сервера и агента.
// Конфигурации могут быть разобраны как из аргументов командной строки, так и из переменных окружения.
package config

import "net"

// NetAddress строка вида [<хост>]:<порт>.
// Данный тип реализует интерфейс pflag.Value.
type NetAddress string

// Возвращает строковое представление сетевого адреса.
func (a NetAddress) String() string {
	return string(a)
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

func (a *NetAddress) Type() string {
	return "string"
}

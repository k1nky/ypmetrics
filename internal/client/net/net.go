package net

import (
	"fmt"
	"net"
)

// RetriveClientAddress возвращает адрес локального хоста.
func RetriveClientAddress() (net.IP, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifs {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ipv4 := ipnet.IP.To4()
			// исключаем loopback
			if ipv4 == nil || ipv4[0] == 127 {
				continue
			}
			return ipv4, nil
		}
	}
	return nil, fmt.Errorf("could not retrive host address")
}

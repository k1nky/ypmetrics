package main

import (
	"fmt"
	"net"

	flag "github.com/spf13/pflag"
)

type NetAddress struct {
	Host string
	Port string
}

var (
	address *NetAddress = &NetAddress{Host: "localhost", Port: "8080"}
)

func (a NetAddress) String() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

func (a *NetAddress) Set(s string) error {
	var err error
	a.Host, a.Port, err = net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if len(a.Host) == 0 {
		a.Host = "localhost"
	}
	return nil
}

func (a *NetAddress) Type() string {
	return "string"
}

func parseFlags() {
	flag.VarP(address, "address", "a", "address and port to listen")

	flag.Parse()
}

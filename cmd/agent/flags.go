package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/k1nky/ypmetrics/internal/agent"
	flag "github.com/spf13/pflag"
)

type netAddress struct {
	Host string
	Port string
}

type duration time.Duration

var (
	address        *netAddress = &netAddress{Host: "localhost", Port: "8080"}
	reportInterval duration    = duration(agent.DefReportInterval)
	pollInterval   duration    = duration(agent.DefPollInterval)
)

func (d *duration) Set(s string) error {
	// используем секунды как единицы измерения по умолчанию, если не указан суффикс
	if seconds, err := strconv.Atoi(s); err == nil {
		*d = duration(time.Second * time.Duration(seconds))
		return nil
	}
	v, err := time.ParseDuration(s)
	*d = duration(v)
	return err
}

func (d *duration) Type() string {
	return "duration"
}

func (d *duration) String() string { return (*time.Duration)(d).String() }

func (a netAddress) String() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

func (a *netAddress) Set(s string) error {
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

func (a *netAddress) Type() string {
	return "string"
}

func parseFlags() {
	flag.VarP(address, "address", "a", "адрес и порт сервера, формат <хост>:<порт>")
	flag.VarP(&reportInterval, "report-interval", "r", "интервал отправки метрик на сервер")
	flag.VarP(&pollInterval, "poll-interval", "p", "интервал сбора метрик")

	flag.Parse()
}

package main

import (
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/k1nky/ypmetrics/internal/agent"
	flag "github.com/spf13/pflag"
)

type netAddress string

type Config struct {
	Address        netAddress
	ReportInterval time.Duration
	PollInterval   time.Duration
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
	reportInterval := cmd.UintP("report-interval", "r", uint(agent.DefReportInterval.Seconds()), "интервал отправки метрик на сервер")
	pollInterval := cmd.UintP("poll-interval", "p", uint(agent.DefPollInterval.Seconds()), "интервал сбора метрик")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	return &Config{
		Address:        address,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		PollInterval:   time.Duration(*pollInterval) * time.Second,
	}, nil
}

func parseFromEnv() (*Config, error) {
	type cfg struct {
		Address        netAddress `env:"ADDRESS"`
		ReportInterval uint       `env:"REPORT_INTERVAL"`
		PollInterval   uint       `env:"POLL_INTERVAL"`
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
		Address:        c.Address,
		ReportInterval: time.Duration(c.ReportInterval) * time.Second,
		PollInterval:   time.Duration(c.PollInterval) * time.Second,
	}, nil
}

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
	if configFromEnv.PollInterval == 0 {
		configFromEnv.PollInterval = configFromCmd.PollInterval
	}
	if configFromEnv.ReportInterval == 0 {
		configFromEnv.ReportInterval = configFromCmd.ReportInterval
	}
	return configFromEnv, nil
}

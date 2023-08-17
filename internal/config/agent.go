package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// AgentConfig конфигурация агента
type AgentConfig struct {
	// Адрес сервера, к которому будет подключаться агент
	Address NetAddress `env:"ADDRESS"`
	// Интервал отправки метрик на сервер (в секундах)
	ReportIntervalInSec uint `env:"REPORT_INTERVAL"`
	// Интервал сбора метрик (в секундах)
	PollIntervalInSec uint `env:"POLL_INTERVAL"`
}

const (
	DefPollIntervalInSec   = 2
	DefReportIntervalInSec = 10
)

func (c AgentConfig) ReportInterval() time.Duration {
	return time.Duration(c.ReportIntervalInSec) * time.Second
}

func (c AgentConfig) PollInterval() time.Duration {
	return time.Duration(c.PollIntervalInSec) * time.Second
}

func parseAgentConfigFromCmd(c *AgentConfig) error {
	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	address := NetAddress("localhost:8080")
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")
	reportInterval := cmd.UintP("report-interval", "r", DefReportIntervalInSec, "интервал отправки метрик на сервер")
	pollInterval := cmd.UintP("poll-interval", "p", DefPollIntervalInSec, "интервал сбора метрик")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}
	*c = AgentConfig{
		Address:             address,
		ReportIntervalInSec: *reportInterval,
		PollIntervalInSec:   *pollInterval,
	}
	return nil
}

func parseAgentConfigFromEnv(c *AgentConfig) error {
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

// ParseAgentConfig возвращает конфиг приложения. Опции разбираются из аргументов командной строки
// и переменных окружения. Переменные окружения имеют приоритет выше чем аргументы командной строки.
func ParseAgentConfig(c *AgentConfig) error {
	if err := parseAgentConfigFromCmd(c); err != nil {
		return err
	}
	if err := parseAgentConfigFromEnv(c); err != nil {
		return err
	}
	return nil
}

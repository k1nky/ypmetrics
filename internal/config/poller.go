package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// Значения по умолчанию
const (
	DefaultPollerPollIntervalInSec   = 2
	DefaultPollerReportIntervalInSec = 10
	DefaultPollerRateLimit           = 0
	DefaultPollerLogLevel            = "info"
	DefaultPollerAddress             = "localhost:8080"
)

// Poller конфигурация агента.
type Poller struct {
	// Адрес сервера, к которому будет подключаться агент
	// Address адрес и порт, на который будут отправляться метрики. По умолчанию localhost:8080.
	// Допустимый формат [хост]:<порт>.
	Address NetAddress `env:"ADDRESS" json:"address"`
	// CryptoKey Путь до файла с приватным ключом (для расшифровки сообщений от агента).
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// ReportIntervalInSec Интервал отправки метрик на сервер (в секундах).
	ReportIntervalInSec uint `env:"REPORT_INTERVAL" json:"report_interval"`
	// PollIntervalInSec Интервал сбора метрик (в секундах).
	PollIntervalInSec uint `env:"POLL_INTERVAL" json:"poll_interval"`
	// LogLevel уровень логирования. По умолчанию info.
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	// Ключ подписи передаваемых данных.
	Key string `env:"KEY" json:"key"`
	// RateLimit ограничение передаваемых метрик за раз. По умолчанию ограничения нет.
	RateLimit uint `env:"RATE_LIMIT" json:"rate_limit"`
	// EnableProfiling доступ к профилировщику. По умолчанию недоступен.
	EnableProfiling bool `env:"ENABLE_PPROF" json:"enable_profiling"`
}

// DefaultPollerConfig конфиг по умолчанию.
var DefaultPollerConfig = Poller{
	Address:             DefaultPollerAddress,
	CryptoKey:           "",
	ReportIntervalInSec: DefaultPollerReportIntervalInSec,
	PollIntervalInSec:   DefaultPollerPollIntervalInSec,
	LogLevel:            DefaultPollerLogLevel,
	Key:                 "",
	RateLimit:           DefaultPollerRateLimit,
	EnableProfiling:     false,
}

// ParsePollerConfig возвращает конфиг Poller'a. Опции разбираются из аргументов командной строки
// и переменных окружения. Переменные окружения имеют приоритет выше чем аргументы командной строки.
func ParsePollerConfig(c *Poller, jsonValue []byte) error {
	if err := parsePollerConfigFromJSON(c, jsonValue); err != nil {
		return err
	}
	if err := parsePollerConfigFromCmd(c); err != nil {
		return err
	}
	if err := parsePollerConfigFromEnv(c); err != nil {
		return err
	}
	return nil
}

// ReportInterval возвращает интервал отправки метрик на сервер в виде time.Duration.
func (c Poller) ReportInterval() time.Duration {
	return time.Duration(c.ReportIntervalInSec) * time.Second
}

// ReportInterval возвращает интервал сбора метрик в виде time.Duration.
func (c Poller) PollInterval() time.Duration {
	return time.Duration(c.PollIntervalInSec) * time.Second
}

func parsePollerConfigFromJSON(c *Poller, jsonValue []byte) error {
	if len(jsonValue) == 0 {
		return nil
	}
	return json.Unmarshal(jsonValue, c)
}

func parsePollerConfigFromCmd(c *Poller) error {

	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	address := NetAddress(c.Address)
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")
	cryptoKey := cmd.StringP("crypto-key", "", c.CryptoKey, "путь до файла с приватным ключом (для расшифровки сообщений от агента).")
	reportInterval := cmd.UintP("report-interval", "r", c.ReportIntervalInSec, "интервал отправки метрик на сервер")
	pollInterval := cmd.UintP("poll-interval", "p", c.PollIntervalInSec, "интервал сбора метрик")
	logLevel := cmd.StringP("log-level", "", c.LogLevel, "уровень логирования")
	key := cmd.StringP("key", "k", c.Key, "ключ хеша")
	rateLimit := cmd.UintP("rate-limit", "l", c.RateLimit, "количество одновременно исходящих запросов на сервер")
	enableProfiling := cmd.BoolP("enable-pprof", "", c.EnableProfiling, "включить профилироовщик")
	cmd.StringP("config", "c", "", "путь к конфигурационному файлу")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}

	*c = Poller{
		Address:             address,
		CryptoKey:           *cryptoKey,
		ReportIntervalInSec: *reportInterval,
		PollIntervalInSec:   *pollInterval,
		LogLevel:            *logLevel,
		Key:                 *key,
		RateLimit:           *rateLimit,
		EnableProfiling:     *enableProfiling,
	}
	return nil
}

func parsePollerConfigFromEnv(c *Poller) error {
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

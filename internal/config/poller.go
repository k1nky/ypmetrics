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
	DefaultPollerShutdownTimeout     = 10
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
	// EnableProfiling доступ к профилировщику. По умолчанию недоступен.
	EnableProfiling bool `env:"ENABLE_PPROF" json:"enable_profiling"`
	// GRPCAddress адрес подключения к grpc серверу. Данное подключение имеет боле высокий приоритет, чем http.
	GRPCAddress NetAddress `env:"GRPC_ADDRESS" json:"grpc_address"`
	// GRPCStream отправлять метрики в потоке.
	GRPCStream bool `env:"GRPC_STREAM" json:"grpc_stream"`
	// Ключ подписи передаваемых данных.
	Key string `env:"KEY" json:"key"`
	// LogLevel уровень логирования. По умолчанию info.
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	// PollIntervalInSec Интервал сбора метрик (в секундах).
	PollIntervalInSec uint `env:"POLL_INTERVAL" json:"poll_interval"`
	// RateLimit ограничение передаваемых метрик за раз. По умолчанию ограничения нет.
	RateLimit uint `env:"RATE_LIMIT" json:"rate_limit"`
	// ReportIntervalInSec Интервал отправки метрик на сервер (в секундах).
	ReportIntervalInSec uint `env:"REPORT_INTERVAL" json:"report_interval"`
	// Таймаут отправки метрик при завершении программы
	ShutdownTimeoutInSec uint `env:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout"`
}

// DefaultPollerConfig конфиг по умолчанию.
var DefaultPollerConfig = Poller{
	Address:              DefaultPollerAddress,
	CryptoKey:            "",
	EnableProfiling:      false,
	GRPCAddress:          "",
	GRPCStream:           true,
	Key:                  "",
	LogLevel:             DefaultPollerLogLevel,
	PollIntervalInSec:    DefaultPollerPollIntervalInSec,
	RateLimit:            DefaultPollerRateLimit,
	ReportIntervalInSec:  DefaultPollerReportIntervalInSec,
	ShutdownTimeoutInSec: DefaultPollerShutdownTimeout,
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

func (c Poller) ShutdownTimeout() time.Duration {
	return time.Duration(c.ShutdownTimeoutInSec) * time.Second
}

func parsePollerConfigFromJSON(c *Poller, jsonValue []byte) error {
	if len(jsonValue) == 0 {
		return nil
	}
	if err := json.Unmarshal(jsonValue, c); err != nil {
		return err
	}
	if len(c.Address) != 0 {
		if err := c.Address.Set(c.Address.String()); err != nil {
			return err
		}
	}
	if len(c.GRPCAddress) != 0 {
		if err := c.GRPCAddress.Set(c.GRPCAddress.String()); err != nil {
			return err
		}
	}
	return nil
}

func parsePollerConfigFromCmd(c *Poller) error {

	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	address := c.Address
	cmd.VarP(&address, "address", "a", "адрес и порт http-сервера, формат: [<адрес>]:<порт>")
	cryptoKey := cmd.StringP("crypto-key", "", c.CryptoKey, "путь до файла с приватным ключом (для расшифровки сообщений от агента).")
	enableProfiling := cmd.BoolP("enable-pprof", "", c.EnableProfiling, "включить профилироовщик")
	grpcAddress := c.GRPCAddress
	cmd.VarP(&grpcAddress, "grpc-address", "", "адрес и порт gRPC-сервера, формат: [<адрес>]:<порт>")
	grpcStream := cmd.BoolP("grpc-stream", "", c.GRPCStream, "")
	key := cmd.StringP("key", "k", c.Key, "ключ хеша")
	logLevel := cmd.StringP("log-level", "", c.LogLevel, "уровень логирования")
	pollInterval := cmd.UintP("poll-interval", "p", c.PollIntervalInSec, "интервал сбора метрик")
	rateLimit := cmd.UintP("rate-limit", "l", c.RateLimit, "количество одновременно исходящих запросов на сервер")
	reportInterval := cmd.UintP("report-interval", "r", c.ReportIntervalInSec, "интервал отправки метрик на сервер")
	shutdownTimeout := cmd.UintP("shutdown-timeout", "", c.ShutdownTimeoutInSec, "таймаут завершения программы")
	cmd.StringP("config", "c", "", "путь к конфигурационному файлу")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}

	*c = Poller{
		Address:              address,
		CryptoKey:            *cryptoKey,
		EnableProfiling:      *enableProfiling,
		GRPCAddress:          grpcAddress,
		GRPCStream:           *grpcStream,
		Key:                  *key,
		LogLevel:             *logLevel,
		PollIntervalInSec:    *pollInterval,
		RateLimit:            *rateLimit,
		ReportIntervalInSec:  *reportInterval,
		ShutdownTimeoutInSec: *shutdownTimeout,
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
	if len(c.GRPCAddress) != 0 {
		if err := c.GRPCAddress.Set(c.GRPCAddress.String()); err != nil {
			return err
		}
	}

	return nil
}

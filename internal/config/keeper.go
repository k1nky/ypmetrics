package config

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// Значения по умолчанию
const (
	DefaultKeeperStoreIntervalInSec = 300
	DefaultKeeperAddress            = ":8080"
	DefaultgRPCAddress              = ""
	DefaultKeeperLogLevel           = "info"
)

// Keeper конфигурация сервера сбора метрик.
type Keeper struct {
	// Address адрес и порт, который будет слушать http-сервер. По умолчанию localhost:8080.
	// Допустимый формат [хост]:<порт>.
	Address NetAddress `env:"ADDRESS" json:"address"`
	// CryptoKey Путь до файла с публичным ключом (для зашифровки передаваемых агентом данных).
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// DatabaseDSN строка подключения к базе данных метрик.
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// EnableProfiling доступ к профилировщику. По умолчанию false.
	EnableProfiling bool `env:"ENABLE_PPROF" json:"enable_pprof"`
	// FileStoragePath путь до файла, в котором будут сохранятся метрики.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// Address адрес и порт, который будет слушать grpc-сервер. По умолчанию localhost:8081.
	// Допустимый формат [хост]:<порт>.
	GRPCAddress NetAddress `env:"GRPC_ADDRESS" json:"grpc_address"`
	// Key секрет для формирования и проверки подписи данных.
	Key string `env:"KEY" json:"key"`
	// LogLevel уровень логирования. По умолчанию info.
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	// Restore восстанавливать метрики из хранилища при старте сервера. По умолчанию true.
	Restore bool `env:"RESTORE" json:"restore"`
	// StoreIntervalInSec интервал в секундах сброса метрик из памяти на диск. По умолчанию 300.
	// Актуально только для файлового хранилища метрик.
	StoreIntervalInSec uint `env:"STORE_INTERVAL" json:"store_interval"`
	// TrustedSubnet доверенная подсеть. При пустом значении переменной trusted_subnet метрики должны обрабатываться сервером без дополнительных ограничений.
	TrustedSubnet Subnet `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// DefaultKeeperConfig конфиг сервера по умолчанию.
var DefaultKeeperConfig = Keeper{
	Address:            DefaultKeeperAddress,
	CryptoKey:          "",
	DatabaseDSN:        "",
	EnableProfiling:    false,
	FileStoragePath:    "",
	GRPCAddress:        DefaultgRPCAddress,
	Restore:            true,
	StoreIntervalInSec: DefaultKeeperStoreIntervalInSec,
	TrustedSubnet:      "",
	Key:                "",
	LogLevel:           DefaultKeeperLogLevel,
}

// ParseKeeperConfig разбирает настройки Keeper'a из файла конфигурации и/или аргументов командной строки
// и/или переменных окружения. Переменные окружения имеют более высокий приоритет, чем аргументы.
// Конфигурационный файл имеет наименьший приоритет.
func ParseKeeperConfig(c *Keeper, jsonValue []byte) error {
	if err := parseKeeperConfigFromJSON(c, jsonValue); err != nil {
		return err
	}
	if err := parseKeeperConfigFromCmd(c); err != nil {
		return err
	}
	if err := parseKeeperConfigFromEnv(c); err != nil {
		return err
	}
	return nil
}

// StorageInterval возвращает интервал сброса метрик из памяти на диск в виде time.Duration.
func (cfg Keeper) StorageInterval() time.Duration {
	return time.Duration(cfg.StoreIntervalInSec) * time.Second
}

func parseKeeperConfigFromJSON(c *Keeper, jsonValue []byte) error {
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
	if len(c.TrustedSubnet) != 0 {
		if err := c.TrustedSubnet.Set(c.TrustedSubnet.String()); err != nil {
			return err
		}
	}
	return nil
}

func parseKeeperConfigFromCmd(c *Keeper) error {
	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	address := c.Address
	cmd.VarP(&address, "address", "a", "адрес и порт http-сервера, формат: [<адрес>]:<порт>")
	cryproKey := cmd.StringP("crypto-key", "", c.CryptoKey, "путь до файла с публичным ключом")
	databaseDSN := cmd.StringP("database-dsn", "d", c.DatabaseDSN, "адрес подключения к БД")
	enableProfiling := cmd.BoolP("enable-pprof", "", c.EnableProfiling, "включить профилировщик")
	grpcAddress := c.GRPCAddress
	cmd.VarP(&grpcAddress, "grpc-address", "", "адрес и порт grpc-сервера, формат: [<адрес>]:<порт>")
	key := cmd.StringP("key", "k", c.Key, "ключ хеширования")
	logLevel := cmd.StringP("log-level", "", c.LogLevel, "уровень логирования")
	// для аргумента --restore запрашиваем сначала значение как строку, а потом уже конверитруем в bool
	// это связано с тем, что формат передачи bool аргументов отличается от требуемого
	// https://github.com/spf13/pflag/issues/288
	restore := cmd.StringP("restore", "r", strconv.FormatBool(c.Restore), "загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	storeInterval := cmd.UintP("store-interval", "i", c.StoreIntervalInSec, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной).")
	storagePath := cmd.StringP("storage-path", "f", c.FileStoragePath, "полное имя файла, куда сохраняются текущие значения")
	trustedSubnet := c.TrustedSubnet
	cmd.VarP(&trustedSubnet, "trusted-subnet", "t", "")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}
	restoreValue, err := strconv.ParseBool(*restore)
	if err != nil {
		return err
	}

	*c = Keeper{
		Address:            address,
		CryptoKey:          *cryproKey,
		DatabaseDSN:        *databaseDSN,
		EnableProfiling:    *enableProfiling,
		FileStoragePath:    *storagePath,
		GRPCAddress:        grpcAddress,
		Key:                *key,
		LogLevel:           *logLevel,
		Restore:            restoreValue,
		StoreIntervalInSec: *storeInterval,
		TrustedSubnet:      trustedSubnet,
	}
	return nil
}

func parseKeeperConfigFromEnv(c *Keeper) error {
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
	if len(c.TrustedSubnet) != 0 {
		if err := c.TrustedSubnet.Set(c.TrustedSubnet.String()); err != nil {
			return err
		}
	}
	return nil
}

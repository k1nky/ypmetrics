package config

import (
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// Значения по умолчанию
const (
	DefaultKeeperStoreIntervalInSec = 300
	DefaultKeeperAddress            = "localhost:8080"
	DefaultKeeperLogLevel           = "info"
)

// Keeper конфигурация сервера сбора метрик.
type Keeper struct {
	// Address адрес и порт, который будет слушать сервер. По умолчанию localhost:8080.
	// Допустимый формат [хост]:<порт>.
	Address NetAddress `env:"ADDRESS"`
	// CryptoKey Путь до файла с публичным ключом (для зашифровки передаваемых агентом данных).
	CryptoKey string `env:"CRYPTO_KEY"`
	// StoreIntervalInSec интервал в секундах сброса метрик из памяти на диск. По умолчанию 300.
	// Актуально только для файлового хранилища метрик.
	StoreIntervalInSec uint `env:"STORE_INTERVAL"`
	// FileStoragePath путь до файла, в котором будут сохранятся метрики.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore восстанавливать метрики из хранилища при старте сервера. По умолчанию true.
	Restore bool `env:"RESTORE"`
	// DatabaseDSN строка подключения к базе данных метрик.
	DatabaseDSN string `env:"DATABASE_DSN"`
	// LogLevel уровень логирования. По умолчанию info.
	LogLevel string `env:"LOG_LEVEL"`
	// Key секрет для формирования и проверки подписи данных.
	Key string `env:"KEY"`
	// EnableProfiling доступ к профилировщику. По умолчанию false.
	EnableProfiling bool `env:"ENABLE_PPROF"`
}

// ParseKeeperConfig разбирает настройки Keeper'a из аргументов командной строки
// и переменных окружения. Переменные окружения имеют более высокий приоритет, чем аргументы.
func ParseKeeperConfig(c *Keeper) error {

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

func parseKeeperConfigFromCmd(c *Keeper) error {
	cmd := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	address := NetAddress(DefaultKeeperAddress)
	cmd.VarP(&address, "address", "a", "адрес и порт сервера, формат: [<адрес>]:<порт>")
	cryproKey := cmd.StringP("crypto-key", "", "", "путь до файла с публичным ключом")
	storeInterval := cmd.UintP("store-interval", "i", DefaultKeeperStoreIntervalInSec, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной).")
	storagePath := cmd.StringP("storage-path", "f", "", "полное имя файла, куда сохраняются текущие значения")
	// для аргумента --restore запрашиваем сначала значение как строку, а потом уже конверитруем в bool
	// это связано с тем, что формат передачи bool аргументов отличается от требуемого
	// https://github.com/spf13/pflag/issues/288
	restore := cmd.StringP("restore", "r", "true", "загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	databaseDSN := cmd.StringP("database-dsn", "d", "", "адрес подключения к БД")
	logLevel := cmd.StringP("log-level", "", DefaultKeeperLogLevel, "уровень логирования")
	key := cmd.StringP("key", "k", "", "ключ хеширования")
	enableProfiling := cmd.BoolP("enable-pprof", "", false, "включить профилировщик")

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
		StoreIntervalInSec: *storeInterval,
		FileStoragePath:    *storagePath,
		Restore:            restoreValue,
		DatabaseDSN:        *databaseDSN,
		LogLevel:           *logLevel,
		Key:                *key,
		EnableProfiling:    *enableProfiling,
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
	return nil
}

package storage

import "time"

type Config struct {
	DSN           string
	StoragePath   string
	StoreInterval time.Duration
	Restore       bool
}

// NewStorage фабрика хранилищ.
func NewStorage(cfg Config, l storageLogger) Storage {
	switch {
	case len(cfg.DSN) > 0:
		return NewDBStorage(l)
	case len(cfg.StoragePath) > 0:
		if cfg.StoreInterval == 0 {
			return NewSyncFileStorage(l)
		}
		return NewAsyncFileStorage(l)
	default:
		return NewMemStorage()
	}
}

package storage

import "time"

type Config struct {
	DSN           string
	StoragePath   string
	StoreInterval time.Duration
	Restore       bool
	Retrier       storageRetrier
}

// NewStorage фабрика хранилищ.
func NewStorage(cfg Config, l storageLogger, retrier storageRetrier) Storage {
	switch {
	case len(cfg.DSN) > 0:
		return NewDBStorage(l, retrier)
	case len(cfg.StoragePath) > 0:
		if cfg.StoreInterval == 0 {
			return NewSyncFileStorage(l, retrier)
		}
		return NewAsyncFileStorage(l, retrier)
	default:
		return NewMemStorage()
	}
}

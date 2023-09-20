package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/retrier"
)

// FileStorage хранит текущие метрики в памяти.
// Позволяет сохранять свое состояние в формате JSON в файл или любой io.Writer.
type FileStorage struct {
	MemStorage
	writeLock sync.Mutex
	retrier   storageRetrier
	logger    storageLogger
}

// AsyncFileStorage хранит текущие метрики в памяти, но периодически сохраняет их в файл.
type AsyncFileStorage struct {
	FileStorage
	stopFlush chan struct{}
}

// SyncFileStorage хранит текущие метрики в памяти и сохраняет их в файл после каждого изменения.
type SyncFileStorage struct {
	FileStorage
	writer *os.File
}

// Можно было бы реализовать синхронный и асинхронный режим в рамках одно типа,
// но логика методов становится более ветвистой и тестировать не удобно.

// NewFileStorage возвращает новое файловое хранилище.
func NewFileStorage(logger storageLogger, retrier storageRetrier) *FileStorage {
	return &FileStorage{
		MemStorage: MemStorage{
			counters: make(map[string]*metric.Counter),
			gauges:   make(map[string]*metric.Gauge),
		},
		logger:  logger,
		retrier: retrier,
	}
}

// NewAsyncFileStorage возвращает новое файловое хранилище, сохранение изменений в котором,
// выполняется асинхронно с заданной периодичностью.
func NewAsyncFileStorage(logger storageLogger, retrier storageRetrier) *AsyncFileStorage {
	return &AsyncFileStorage{
		FileStorage: FileStorage{
			MemStorage: MemStorage{
				counters: make(map[string]*metric.Counter),
				gauges:   make(map[string]*metric.Gauge),
			},
			logger:  logger,
			retrier: retrier,
		},
		stopFlush: make(chan struct{}),
	}
}

// NewSyncFileStorage возвращает новое файловое хранилище, сохранение изменений в котором,
// выполняется синхронно.
func NewSyncFileStorage(logger storageLogger, retrier storageRetrier) *SyncFileStorage {
	return &SyncFileStorage{
		FileStorage: FileStorage{
			MemStorage: MemStorage{
				counters: make(map[string]*metric.Counter),
				gauges:   make(map[string]*metric.Gauge),
			},
			logger:  logger,
			retrier: retrier,
		},
	}
}

// Flush делает срез метрик и сохраняет его в поток
func (fs *FileStorage) Flush(w io.Writer) error {
	snap := metric.Metrics{}
	ctx := context.Background()
	fs.Snapshot(ctx, &snap)

	if err := json.NewEncoder(w).Encode(snap); err != nil {
		return err
	}
	return nil
}

// Restore восстанавливает метрики из потока
func (fs *FileStorage) Restore(r io.Reader) error {
	snap := metric.Metrics{}
	if err := json.NewDecoder(r).Decode(&snap); err != nil {
		return err
	}

	counters := make(map[string]*metric.Counter)
	gauges := make(map[string]*metric.Gauge)

	for _, c := range snap.Counters {
		counters[c.Name] = metric.NewCounter(c.Name, c.Value)
	}
	for _, g := range snap.Gauges {
		gauges[g.Name] = metric.NewGauge(g.Name, g.Value)
	}

	fs.countersLock.Lock()
	defer fs.countersLock.Unlock()
	fs.gaugesLock.Lock()
	defer fs.gaugesLock.Unlock()
	fs.counters = counters
	fs.gauges = gauges

	return nil
}

// WriteToFile сохраняет метрики в файл. Файл должен быть предварительно открыт.
func (fs *FileStorage) WriteToFile(f *os.File) error {
	var err error
	for fs.retrier.Init(retrier.AlwaysRetry); fs.retrier.Next(err); {
		err = fs.writeToFile(f)
		if err != nil {
			fs.logger.Errorf("WriteToFile: %v", err)
		}
	}
	return err
}

func (fs *FileStorage) writeToFile(f *os.File) error {
	fs.writeLock.Lock()
	defer fs.writeLock.Unlock()
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if err := fs.Flush(f); err != nil {
		return err
	}
	return f.Sync()
}

// Close закрывает асинхронное файловое хранилище
func (afs *AsyncFileStorage) Close() error {
	close(afs.stopFlush)
	return nil
}

// Close закрывает синхронное файловое хранилище
func (sfs *SyncFileStorage) Close() error {
	return sfs.writer.Close()
}

// Open открывает асинхронное файловое хранилище
func (afs *AsyncFileStorage) Open(cfg Config) error {
	f, err := os.OpenFile(cfg.StoragePath, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	if cfg.Restore {
		if err := afs.Restore(f); err != nil {
			if !os.IsNotExist(err) {
				afs.logger.Errorf("Open: %v", err)
			}
		}
	}
	go func() {
		t := time.NewTicker(cfg.StoreInterval)
		defer f.Close()
		for {
			select {
			case <-afs.stopFlush:
				return
			case <-t.C:
				if err := afs.WriteToFile(f); err != nil {
					afs.logger.Errorf("Flash: %v", err)
				}
			}
		}
	}()
	return nil
}

// Open открывает синхронное файловое хранилище
func (sfs *SyncFileStorage) Open(cfg Config) error {
	f, err := os.OpenFile(cfg.StoragePath, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	if cfg.Restore {
		if err := sfs.Restore(f); err != nil {
			if !os.IsNotExist(err) {
				sfs.logger.Errorf("Open: %v", err)
			}
		}
	}
	sfs.writer = f
	return nil
}

// SetCounter записывает значение метрики типа Counter и сохраняет изменения в файл.
func (sfs *SyncFileStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	if err := sfs.MemStorage.UpdateCounter(ctx, name, value); err != nil {
		return err
	}
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Errorf("SetCounter: %v", err)
		return err
	}
	return nil
}

// SetGauge записывает значение метрики типа Gauge и сохраняет изменения в файл.
func (sfs *SyncFileStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	if err := sfs.MemStorage.UpdateGauge(ctx, name, value); err != nil {
		return err
	}
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Errorf("SetGauge: %v", err)
		return err
	}
	return nil
}

func (sfs *SyncFileStorage) UpdateMetrics(ctx context.Context, metrics metric.Metrics) error {
	if err := sfs.MemStorage.UpdateMetrics(ctx, metrics); err != nil {
		return err
	}
	if err := sfs.writeToFile(sfs.writer); err != nil {
		sfs.logger.Errorf("SetGauge: %v", err)
		return err
	}
	return nil
}

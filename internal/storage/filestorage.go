package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// FileStorage хранит текущие метрики в памяти.
// Позволяет сохранять свое состояние в формате JSON в файл или любой io.Writer.
type FileStorage struct {
	MemStorage
	writeLock sync.Mutex
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
func NewFileStorage(logger storageLogger) *FileStorage {
	return &FileStorage{
		MemStorage: MemStorage{
			counters: make(map[string]*metric.Counter),
			gauges:   make(map[string]*metric.Gauge),
		},
		logger: logger,
	}
}

// NewAsyncFileStorage возвращает новое файловое хранилище, сохранение изменений в котором,
// выполняется асинхронно с заданной периодичностью.
func NewAsyncFileStorage(logger storageLogger) *AsyncFileStorage {
	return &AsyncFileStorage{
		FileStorage: FileStorage{
			MemStorage: MemStorage{
				counters: make(map[string]*metric.Counter),
				gauges:   make(map[string]*metric.Gauge),
			},
			logger: logger,
		},
		stopFlush: make(chan struct{}),
	}
}

// NewSyncFileStorage возвращает новое файловое хранилище, сохранение изменений в котором,
// выполняется синхронно.
func NewSyncFileStorage(logger storageLogger) *SyncFileStorage {
	return &SyncFileStorage{
		FileStorage: FileStorage{
			MemStorage: MemStorage{
				counters: make(map[string]*metric.Counter),
				gauges:   make(map[string]*metric.Gauge),
			},
			logger: logger,
		},
	}
}

// Flush делает срез метрик и сохраняет его в поток
func (fs *FileStorage) Flush(w io.Writer) error {
	snap := metric.Metrics{}
	fs.Snapshot(&snap)

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
func (afs *AsyncFileStorage) Open(filename string, restore bool, flushInterval time.Duration) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	if err := afs.Restore(f); err != nil {
		if !os.IsNotExist(err) {
			afs.logger.Error("Open: %v", err)
		}
	}
	go func() {
		t := time.NewTicker(flushInterval)
		defer f.Close()
		for {
			select {
			case <-afs.stopFlush:
				return
			case <-t.C:
				if err := afs.WriteToFile(f); err != nil {
					afs.logger.Error("Flash: %v", err)
				}
			}
		}
	}()
	return nil
}

// Open открывает синхронное файловое хранилище
func (sfs *SyncFileStorage) Open(filename string, restore bool) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	if err := sfs.Restore(f); err != nil {
		if !os.IsNotExist(err) {
			sfs.logger.Error("Open: %v", err)
		}
	}
	sfs.writer = f
	return nil
}

// SetCounter записывает значение метрики типа Counter и сохраняет изменения в файл.
func (sfs *SyncFileStorage) UpdateCounter(name string, value int64) {
	sfs.MemStorage.UpdateCounter(name, value)
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Error("SetCounter: %v", err)
	}
}

// SetGauge записывает значение метрики типа Gauge и сохраняет изменения в файл.
func (sfs *SyncFileStorage) UpdateGauge(name string, value float64) {
	sfs.MemStorage.UpdateGauge(name, value)
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Error("SetGauge: %v", err)
	}
}

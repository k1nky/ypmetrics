package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/metric"
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
	flushInterval time.Duration
	isClosed      bool
	closeLock     sync.RWMutex
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
func NewAsyncFileStorage(logger storageLogger, flushInterval time.Duration) *AsyncFileStorage {
	return &AsyncFileStorage{
		FileStorage: FileStorage{
			MemStorage: MemStorage{
				counters: make(map[string]*metric.Counter),
				gauges:   make(map[string]*metric.Gauge),
			},
			logger: logger,
		},
		flushInterval: flushInterval,
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
	// выставляем флаг, что горутина, в которой периодически сохраняются метрики
	// должна закрыться
	// TODO: канал подходит лучше для этой задачи, оставить для будущих спринтов
	afs.closeLock.Lock()
	defer afs.closeLock.Unlock()
	afs.isClosed = true
	return nil
}

// Close закрывает синхронное файловое хранилище
func (sfs *SyncFileStorage) Close() error {
	return sfs.writer.Close()
}

// Open открывает асинхронное файловое хранилище
func (afs *AsyncFileStorage) Open(filename string, restore bool) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	if err := afs.Restore(f); err != nil {
		afs.logger.Error("Open: %v", err)
	}
	go func() {
		defer f.Close()
		for {
			time.Sleep(afs.flushInterval)
			afs.closeLock.RLock()
			defer afs.closeLock.RUnlock()
			if afs.isClosed {
				return
			}
			if err := afs.WriteToFile(f); err != nil {
				afs.logger.Error("Flash: %v", err)
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
		sfs.logger.Error("Open: %v", err)
	}
	sfs.writer = f
	return nil
}

// SetCounter записывает значение метрики типа Counter и сохраняет изменения в файл.
func (sfs *SyncFileStorage) SetCounter(m *metric.Counter) {
	sfs.MemStorage.SetCounter(m)
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Error("SetCounter: %v", err)
	}
}

// SetGauge записывает значение метрики типа Gauge и сохраняет изменения в файл.
func (sfs *SyncFileStorage) SetGauge(m *metric.Gauge) {
	sfs.MemStorage.SetGauge(m)
	if err := sfs.WriteToFile(sfs.writer); err != nil {
		sfs.logger.Error("SetGauge: %v", err)
	}
}

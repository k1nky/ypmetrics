package storage

import (
	"encoding/gob"
	"os"
	"sync"
	"time"

	"github.com/k1nky/ypmetrics/internal/metric"
)

type DurableMemStorage struct {
	MemStorage
	FlushInterval time.Duration
	logger        storageLogger
	data          *os.File
	dataLock      sync.Mutex
}

func NewDurableMemStorage(filename string, flushInterval time.Duration, logger storageLogger) (*DurableMemStorage, error) {
	data, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	dms := &DurableMemStorage{
		MemStorage: MemStorage{
			counters: make(map[string]*metric.Counter),
			gauges:   make(map[string]*metric.Gauge),
		},
		data:          data,
		FlushInterval: flushInterval,
	}

	if dms.FlushInterval != 0 {
		dms.background()
	}

	return dms, nil
}

func (dms *DurableMemStorage) background() {
	go func() {
		for {
			time.Sleep(dms.FlushInterval)
			dms.Flush()
		}
	}()
}

func (dms *DurableMemStorage) Flush() {
	snap := metric.Metrics{}
	dms.Snapshot(&snap)

	dms.dataLock.Lock()
	defer dms.dataLock.Unlock()
	if err := gob.NewEncoder(dms.data).Encode(snap); err != nil {
		dms.logger.Error("Flash: %v", err)
	}
}

func (dms *DurableMemStorage) Restore() error {
	snap := metric.Metrics{}
	if err := gob.NewDecoder(dms.data).Decode(&snap); err != nil {
		return err
	}

	dms.countersLock.Lock()
	defer dms.countersLock.Unlock()
	for k := range dms.counters {
		delete(dms.counters, k)
	}
	for _, c := range snap.Counters {
		dms.counters[c.Name] = metric.NewCounter(c.Name, c.Value)
	}

	dms.gaugesLock.Lock()
	defer dms.gaugesLock.Unlock()
	for k := range dms.gauges {
		delete(dms.gauges, k)
	}
	for _, g := range snap.Gauges {
		dms.gauges[g.Name] = metric.NewGauge(g.Name, g.Value)
	}

	return nil
}

func (dms *DurableMemStorage) Close() error {
	if dms.data != nil {
		return dms.data.Close()
	}
	return nil
}

func (dms *DurableMemStorage) SetCounter(m *metric.Counter) {
	dms.MemStorage.SetCounter(m)
	if dms.FlushInterval == 0 {
		dms.Flush()
	}
}

func (dms *DurableMemStorage) SetGauge(m *metric.Gauge) {
	dms.MemStorage.SetGauge(m)
	if dms.FlushInterval == 0 {
		dms.Flush()
	}
}

package agent

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

const (
	DefPollInterval   = 2 * time.Second
	DefReportInterval = 10 * time.Second
)

type Agent struct {
	storage        storage.Storage
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type Option func(*Agent)

var runtimeMetricsList = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

func WithStorage(storage storage.Storage) Option {
	return func(a *Agent) {
		a.storage = storage
	}
}

func WithPollInterval(interval time.Duration) Option {
	return func(a *Agent) {
		a.PollInterval = interval
	}
}

func WithReportInterval(interval time.Duration) Option {
	return func(a *Agent) {
		a.ReportInterval = interval
	}
}

func New(options ...Option) *Agent {
	s := &Agent{
		PollInterval:   DefPollInterval,
		ReportInterval: DefReportInterval,
	}

	for _, opt := range options {
		opt(s)
	}
	if s.storage == nil {
		s.storage = storage.NewMemStorage()
	}

	return s
}

func (a *Agent) Run() {

	for _, v := range runtimeMetricsList {
		a.storage.Set(&metric.Gauge{
			Name: v,
		})
	}
	a.storage.Set(&metric.Counter{
		Name: "PollCounter",
	})
	a.storage.Set(&metric.Gauge{
		Name: "RandomValue",
	})

	go func() {
		a.report()
		time.Sleep(a.ReportInterval)
	}()
	for {
		a.pollRuntime()
		a.storage.Get("PollCounter").Update(1)
		a.storage.Get("RandomValue").Update(randomFloat())
		time.Sleep(a.PollInterval)
	}
}

func (a *Agent) report() {

}

func (a *Agent) pollRuntime() {
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	a.storage.Get("Alloc").Update(memStat.Alloc)
	a.storage.Get("BuckHashSys").Update(memStat.BuckHashSys)
	a.storage.Get("Frees").Update(memStat.Frees)
	a.storage.Get("GCCPUFraction").Update(memStat.GCCPUFraction)
	a.storage.Get("GCSys").Update(memStat.GCSys)
	a.storage.Get("HeapAlloc").Update(memStat.HeapAlloc)
	a.storage.Get("HeapIdle").Update(memStat.HeapIdle)
	a.storage.Get("HeapInuse").Update(memStat.HeapInuse)
	a.storage.Get("HeapObjects").Update(memStat.HeapObjects)
	a.storage.Get("HeapReleased").Update(memStat.HeapReleased)
	a.storage.Get("HeapSys").Update(memStat.HeapSys)
	a.storage.Get("LastGC").Update(memStat.LastGC)
	a.storage.Get("Lookups").Update(memStat.Lookups)
	a.storage.Get("MCacheInuse").Update(memStat.MCacheInuse)
	a.storage.Get("MCacheSys").Update(memStat.MCacheSys)
	a.storage.Get("MSpanInuse").Update(memStat.MSpanInuse)
	a.storage.Get("MSpanSys").Update(memStat.MSpanSys)
	a.storage.Get("Mallocs").Update(memStat.Mallocs)
	a.storage.Get("NextGC").Update(memStat.NextGC)
	a.storage.Get("NumForcedGC").Update(memStat.NumForcedGC)
	a.storage.Get("NumGC").Update(memStat.NumGC)
	a.storage.Get("OtherSys").Update(memStat.OtherSys)
	a.storage.Get("PauseTotalNs").Update(memStat.PauseTotalNs)
	a.storage.Get("StackInuse").Update(memStat.StackInuse)
	a.storage.Get("StackSys").Update(memStat.StackSys)
	a.storage.Get("Sys").Update(memStat.Sys)
	a.storage.Get("TotalAlloc").Update(memStat.TotalAlloc)
}

func randomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.NormFloat64()
}

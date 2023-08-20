package main

import (
	"os"

	"github.com/k1nky/ypmetrics/internal/apiclient"
	"github.com/k1nky/ypmetrics/internal/collector"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/metricset/agent"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func main() {
	l := logger.New()
	cfg := config.AgentConfig{}
	if err := config.ParseAgentConfig(&cfg); err != nil {
		l.Error("config: %s", err)
		os.Exit(1)
	}
	// для агента храним метрики в памяти
	stor := storage.NewMemStorage()
	client := apiclient.New(string(cfg.Address))
	a := agent.New(cfg, stor, l, client)
	a.AddCollector(collector.PollCounter{}, collector.Random{}, collector.Runtime{})
	a.Run()
}

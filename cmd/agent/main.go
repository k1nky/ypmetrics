package main

import (
	"context"
	"os"

	"github.com/k1nky/ypmetrics/internal/apiclient"
	"github.com/k1nky/ypmetrics/internal/collector"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/k1nky/ypmetrics/internal/usecases/poller"
)

func main() {
	l := logger.New()
	cfg := config.PollerConfig{}
	if err := config.ParsePollerConfig(&cfg); err != nil {
		l.Error("config: %s", err)
		os.Exit(1)
	}
	Run(l, cfg)
}

func Run(l *logger.Logger, cfg config.PollerConfig) {
	// для агента храним метрики в памяти
	store := storage.NewMemStorage()
	defer store.Close()
	client := apiclient.New(string(cfg.Address))
	p := poller.New(cfg, store, l, client)
	p.AddCollector(collector.PollCounter{}, collector.Random{}, collector.Runtime{})
	ctx, cancel := context.WithCancel(context.TODO())
	p.Run(ctx)
	defer cancel()
}

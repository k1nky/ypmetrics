package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/k1nky/ypmetrics/internal/apiclient"
	"github.com/k1nky/ypmetrics/internal/collector"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/k1nky/ypmetrics/internal/usecases/poller"
)

const (
	DefaultProfilerAddress = "localhost:8099"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	l := logger.New()
	cfg := config.Poller{}
	if err := config.ParsePollerConfig(&cfg); err != nil {
		l.Errorf("config: %s", err)
		exit(1)
	}
	if err := l.SetLevel(cfg.LogLevel); err != nil {
		l.Errorf("logger: %s", err)
		exit(1)
	}
	l.Debugf("config: %+v", cfg)
	showVersion()
	Run(l, cfg)
}

func Run(l *logger.Logger, cfg config.Poller) {
	// для агента храним метрики в памяти
	store := storage.NewMemStorage()
	defer store.Close()

	client := apiclient.New(string(cfg.Address), l)
	// сначала сжимаем данные, затем подписываем
	client.SetGzip().SetKey(cfg.Key)

	p := poller.New(cfg, store, l, client)
	p.AddCollector(
		&collector.PollCounter{},
		&collector.Random{},
		&collector.Runtime{},
		&collector.Gops{},
	)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	p.Run(ctx)
	if cfg.EnableProfiling {
		exposeProfiler(ctx, l)
	}
	<-ctx.Done()
	time.Sleep(time.Second)
}

func exposeProfiler(ctx context.Context, l *logger.Logger) {
	server := http.Server{
		Addr: DefaultProfilerAddress,
	}
	go func() {
		l.Infof("expose profiler on %s/debug/pprof", DefaultProfilerAddress)
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				l.Errorf("unexpected profiler closing: %v", err)
			}
		}
	}()
	go func() {
		<-ctx.Done()
		if err := server.Close(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				l.Errorf("unexpected error: %v", err)
			}
		}
	}()
}

func exit(rc int) {
	os.Exit(rc)
}

func showVersion() {
	s := strings.Builder{}
	fmt.Fprintf(&s, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&s, "Build date: %s\n", buildDate)
	fmt.Fprintf(&s, "Build commit: %s\n", buildCommit)
	fmt.Println(s.String())
}

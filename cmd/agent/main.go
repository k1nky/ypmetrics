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

	"github.com/k1nky/ypmetrics/internal/client"
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
	cfg := config.DefaultPollerConfig
	if err := parseConfig(&cfg); err != nil {
		if config.IsHelpWanted(err) {
			return
		}
		l.Errorf("config: %s", err)
		exit(1)
	}
	if err := l.SetLevel(cfg.LogLevel); err != nil {
		l.Errorf("logger: %s", err)
		exit(1)
	}
	l.Debugf("config: %+v", cfg)
	showVersion()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	run(ctx, l, cfg)
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

func parseConfig(c *config.Poller) error {
	var (
		err       error
		jsonValue []byte
	)
	configPath := config.GetConfigPath()
	if len(configPath) != 0 {
		// файл с конфигом указан, поэтому читаем сначала его
		if jsonValue, err = os.ReadFile(configPath); err != nil {
			return err
		}
	}
	return config.ParsePollerConfig(c, jsonValue)
}

func run(ctx context.Context, l *logger.Logger, cfg config.Poller) {
	// для агента храним метрики в памяти
	store := storage.NewMemStorage()
	defer store.Close()

	client, err := client.New(ctx, client.Config{
		CryptoKey:        cfg.CryptoKey,
		GRPCAddress:      cfg.GRPCAddress.String(),
		GRPCPushToStream: cfg.GRPCStream,
		HTTPAddress:      cfg.Address.String(),
		Key:              cfg.Key,
	}, l)
	if err != nil {
		l.Errorf("create metric client: %s", err)
		exit(1)
	}
	defer client.Close()

	p := poller.New(cfg, store, l, client)
	p.AddCollector(
		&collector.PollCounter{},
		&collector.Random{},
		&collector.Runtime{},
		&collector.Gops{},
	)

	done := p.Run(ctx)
	if cfg.EnableProfiling {
		exposeProfiler(ctx, l)
	}
	// ожидаем завершения программы по сигналу
	<-ctx.Done()
	// ожидаем завершения отправки метрик или принудительно по таймауту
	select {
	case <-done:
	case <-time.After(cfg.ShutdownTimeout()):
	}
}

func showVersion() {
	s := strings.Builder{}
	fmt.Fprintf(&s, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&s, "Build date: %s\n", buildDate)
	fmt.Fprintf(&s, "Build commit: %s\n", buildCommit)
	fmt.Println(s.String())
}

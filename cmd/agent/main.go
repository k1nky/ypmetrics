package main

import (
	"context"
	"crypto/rsa"
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
	"github.com/k1nky/ypmetrics/internal/crypto"
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
			exit(0)
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
	run(l, cfg)
}

func run(l *logger.Logger, cfg config.Poller) {
	// для агента храним метрики в памяти
	store := storage.NewMemStorage()
	defer store.Close()

	client := apiclient.New(string(cfg.Address), l)
	key, err := readCryptoKey(cfg.CryptoKey)
	if err != nil {
		l.Errorf("config: %s", err)
		exit(1)
	}
	// сжимаем данные -> шифруем -> подписываем
	client.SetGzip().SetEncrypt(key).SetKey(cfg.Key)

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

func readCryptoKey(path string) (*rsa.PublicKey, error) {
	if len(path) == 0 {
		return nil, nil
	}
	f, err := os.Open(path)
	defer func() { _ = f.Close() }()
	if err != nil {
		return nil, err
	}
	key, err := crypto.ReadPublicKey(f)
	return key, err
}

func showVersion() {
	s := strings.Builder{}
	fmt.Fprintf(&s, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&s, "Build date: %s\n", buildDate)
	fmt.Fprintf(&s, "Build commit: %s\n", buildCommit)
	fmt.Println(s.String())
}

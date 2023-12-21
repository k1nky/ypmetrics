package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/crypto"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/handler/middleware"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/retrier"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
)

const (
	DefaultProfilerPrefix = "/debug/pprof"
)

const (
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	DefaultCloseTimeout = 5 * time.Second
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func parseConfig(cfg *config.Keeper) error {
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
	return config.ParseKeeperConfig(cfg, jsonValue)
}

func main() {
	l := logger.New()
	cfg := config.DefaultKeeperConfig
	err := parseConfig(&cfg)
	if err != nil {
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

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	run(ctx, l, cfg)
}

func run(ctx context.Context, l *logger.Logger, cfg config.Keeper) {
	storeConfig := storage.Config{
		DSN:           cfg.DatabaseDSN,
		StoragePath:   cfg.FileStoragePath,
		StoreInterval: cfg.StorageInterval(),
		Restore:       cfg.Restore,
	}
	store := storage.NewStorage(storeConfig, l, retrier.New())
	if err := store.Open(storeConfig); err != nil {
		l.Errorf("opening storage: %v", err)
	}
	defer store.Close()

	uc := keeper.New(store, cfg, l)
	h := handler.New(*uc)

	decryptKey, err := readCryptoKey(cfg.CryptoKey)
	if err != nil {
		l.Errorf("config: %s", err)
		exit(1)
	}

	router := newRouter(h, l, cfg.Key, decryptKey)
	if cfg.EnableProfiling {
		l.Infof("expose profiler on %s", DefaultProfilerPrefix)
		exposeProfiler(router)
	}

	l.Infof("starting on %s", cfg.Address)
	runHTTPServer(ctx, cfg.Address.String(), router, l)
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}

func runHTTPServer(ctx context.Context, addr string, handler http.Handler, l *logger.Logger) {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		WriteTimeout: DefaultWriteTimeout,
		ReadTimeout:  DefaultReadTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				l.Errorf("unexpected server closing: %v", err)
			}
		}
	}()
	// отслеживаем завершение программы
	go func() {
		<-ctx.Done()
		l.Debugf("closing http server")
		c, cancel := context.WithTimeout(context.Background(), DefaultCloseTimeout)
		defer cancel()
		srv.Shutdown(c)
	}()
}

func newRouter(h handler.Handler, l *logger.Logger, sealKey string, decryptKey *rsa.PrivateKey) *gin.Engine {
	router := gin.New()
	// логируем запрос
	router.Use(middleware.Logger(l))
	if len(sealKey) > 0 {
		// если указан ключ, то проверяем подпись полученных данных
		router.Use(middleware.NewSeal(sealKey).Use())
	}
	if decryptKey != nil {
		// указан ключ шифрования, то расшифровываем тело запроса
		router.Use(middleware.NewDecrypter(decryptKey).Use())
	}
	// при необходимости раcпаковываем/запаковываем данные
	router.Use(middleware.NewGzip([]string{"application/json", "text/html"}).Use())

	router.GET("/", h.AllMetrics())
	router.GET("/ping", h.Ping())
	router.POST("/updates/", middleware.RequireContentType("application/json"), h.UpdatesJSON())

	valueRoutes := router.Group("/value")
	valueRoutes.POST("/", middleware.RequireContentType("application/json"), h.ValueJSON())
	valueRoutes.GET("/:type/:name", h.Value())

	updateRoutes := router.Group("/update")
	updateRoutes.POST("/", middleware.RequireContentType("application/json"), h.UpdateJSON())
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", h.Update())

	return router
}

func exposeProfiler(r *gin.Engine) {
	g := r.Group(DefaultProfilerPrefix)
	g.GET("/", gin.WrapF(pprof.Index))
	g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	g.GET("/profile", gin.WrapF(pprof.Profile))
	g.GET("/trace", gin.WrapF(pprof.Trace))
	g.GET("/symbol", gin.WrapF(pprof.Symbol))
	g.POST("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	g.GET("/block", gin.WrapH(pprof.Handler("block")))
	g.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	g.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	g.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	g.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}

func exit(rc int) {
	os.Exit(rc)
}

func readCryptoKey(path string) (*rsa.PrivateKey, error) {
	if len(path) == 0 {
		return nil, nil
	}
	f, err := os.Open(path)
	defer func() { _ = f.Close() }()
	if err != nil {
		return nil, err
	}
	key, err := crypto.ReadPrivateKey(f)
	return key, err
}

func showVersion() {
	s := strings.Builder{}
	fmt.Fprintf(&s, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&s, "Build date: %s\n", buildDate)
	fmt.Fprintf(&s, "Build commit: %s\n", buildCommit)
	fmt.Println(s.String())
}

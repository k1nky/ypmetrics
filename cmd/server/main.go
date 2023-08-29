package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/handler/middleware"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/metricset"
	"github.com/k1nky/ypmetrics/internal/storage"
)

type fileStorage interface {
	Open(filename string) error
	RestoreFromFile(filename string) error
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	SetCounter(*metric.Counter)
	SetGauge(*metric.Gauge)
	Snapshot(*metric.Metrics)
	Close()
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func parseConfig() (config.KeeperConfig, error) {
	cfg := config.KeeperConfig{}
	err := config.ParseKeeperConfig(&cfg)
	return cfg, err
}

func openStorage(cfg config.KeeperConfig, log *logger.Logger) (fileStorage, error) {
	var store fileStorage

	if cfg.StoreIntervalInSec == 0 {
		store = storage.NewSyncFileStorage(log)
	} else {
		store = storage.NewAsyncFileStorage(log, cfg.StorageInterval())
	}
	if cfg.Restore {
		store.RestoreFromFile(cfg.FileStoragePath)
	}
	err := store.Open(cfg.FileStoragePath)
	return store, err
}

func main() {
	logger := logger.New()
	cfg, err := parseConfig()
	if err != nil {
		logger.Error("config: %s", err)
		os.Exit(1)
	}

	store, err := openStorage(cfg, logger)
	if err != nil {
		logger.Error("opening storage: %v", err)
	}
	defer store.Close()

	// handler - слой для работы с метриками по HTTP
	metrics := metricset.NewSet(store)
	router := newRouter(metrics, logger)

	logger.Info("starting on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address.String(), router); err != nil {
		panic(err)
	}
}

func newRouter(metrics *metricset.Set, log *logger.Logger) *gin.Engine {
	h := handler.New(metrics)

	router := gin.New()
	router.Use(middleware.Logger(log), middleware.NewGzip([]string{"application/json", "text/html"}).Use())

	router.GET("/", h.AllMetrics())

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

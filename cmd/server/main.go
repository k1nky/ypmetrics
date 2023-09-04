package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/handler/middleware"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
)

type metricStorage interface {
	GetCounter(name string) *metric.Counter
	GetGauge(name string) *metric.Gauge
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	Snapshot(*metric.Metrics)
	Close() error
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func parseConfig() (config.KeeperConfig, error) {
	cfg := config.KeeperConfig{}
	err := config.ParseKeeperConfig(&cfg)
	return cfg, err
}

func openStorage(cfg config.KeeperConfig, log *logger.Logger) (metricStorage, error) {
	switch {
	case len(cfg.DatabaseDSN) > 0:
		s := storage.NewDBStorage(log)
		return s, s.Open(cfg.DatabaseDSN)
	case cfg.StoreIntervalInSec == 0:
		s := storage.NewSyncFileStorage(log)
		return s, s.Open(cfg.FileStoragePath, cfg.Restore)
	case cfg.StoreIntervalInSec > 0:
		s := storage.NewAsyncFileStorage(log)
		return s, s.Open(cfg.FileStoragePath, cfg.Restore, cfg.StorageInterval())
	default:
		return storage.NewMemStorage(), nil
	}
}

func main() {
	l := logger.New()
	cfg, err := parseConfig()
	if err != nil {
		l.Error("config: %s", err)
		os.Exit(1)
	}
	Run(l, cfg)
}

func Run(l *logger.Logger, cfg config.KeeperConfig) {
	store, err := openStorage(cfg, l)
	if err != nil {
		l.Error("opening storage: %v", err)
	}

	defer store.Close()
	uc := keeper.New(store, cfg, l)
	h := handler.New(*uc)
	router := newRouter(h, l)

	l.Info("starting on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address.String(), router); err != nil {
		panic(err)
	}

}

func newRouter(h handler.Handler, l *logger.Logger) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(l), middleware.NewGzip([]string{"application/json", "text/html"}).Use())

	router.GET("/", h.AllMetrics())
	router.GET("/ping", h.Ping())

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

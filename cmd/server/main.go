package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/handler/middleware"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/retrier"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func parseConfig() (config.KeeperConfig, error) {
	cfg := config.KeeperConfig{}
	err := config.ParseKeeperConfig(&cfg)
	return cfg, err
}

func main() {
	l := logger.New()
	cfg, err := parseConfig()
	if err != nil {
		l.Error("config: %s", err)
		os.Exit(1)
	}
	if err := l.SetLevel(cfg.LogLevel); err != nil {
		l.Error("config: %s", err)
		os.Exit(1)
	}
	l.Debug("config: %+v", cfg)

	Run(l, cfg)
}

func Run(l *logger.Logger, cfg config.KeeperConfig) {
	storeConfig := storage.Config{
		DSN:           cfg.DatabaseDSN,
		StoragePath:   cfg.FileStoragePath,
		StoreInterval: cfg.StorageInterval(),
		Restore:       cfg.Restore,
	}
	store := storage.NewStorage(storeConfig, l, retrier.New())
	if err := store.Open(storeConfig); err != nil {
		l.Error("opening storage: %v", err)
	}
	defer store.Close()

	uc := keeper.New(store, cfg, l)
	h := handler.New(*uc)
	router := newRouter(h, l, cfg.Key)

	l.Info("starting on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address.String(), router); err != nil {
		panic(err)
	}
}

func newRouter(h handler.Handler, l *logger.Logger, key string) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger(l))
	if len(key) > 0 {
		router.Use(middleware.NewSeal(key).Use())
	}
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

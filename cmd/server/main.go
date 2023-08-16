package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	logger := logger.New()
	cfg := config.ServerConfig{}
	if err := config.ParseServerConfig(&cfg); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if err := run(cfg, logger); err != nil {
		panic(err)
	}
}

func run(cfg config.ServerConfig, l *logger.Logger) error {

	ms := metric.NewSet(storage.NewMemStorage())

	router := gin.New()
	router.Use(handler.Logger(l))
	router.GET("/", handler.AllMetricsHandler(*ms))
	valueRoutes := router.Group("/value")
	valueRoutes.GET("/:type/:name", handler.ValueHandler(*ms))
	updateRoutes := router.Group("/update")
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", handler.UpdateHandler(*ms))

	l.Info("starting on ", cfg.Address)
	return http.ListenAndServe(cfg.Address.String(), router)
}

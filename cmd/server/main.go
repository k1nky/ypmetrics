package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/metricset/server"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	logger := logger.New()

	cfg := config.ServerConfig{}
	if err := config.ParseServerConfig(&cfg); err != nil {
		logger.Error("config: %s", err)
		os.Exit(1)
	}

	router := newRouter(cfg, logger)
	logger.Info("starting on %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address.String(), router); err != nil {
		panic(err)
	}
}

func newRouter(cfg config.ServerConfig, l *logger.Logger) *gin.Engine {
	metrics := server.New(storage.NewMemStorage(), l)
	h := handler.New(metrics)

	router := gin.New()
	router.Use(handler.Logger(l))

	router.GET("/", h.AllMetrics())

	valueRoutes := router.Group("/value")
	valueRoutes.POST("/", handler.RequireContentType("application/json"), h.ValueJSON())
	valueRoutes.GET("/:type/:name", h.Value())

	updateRoutes := router.Group("/update")
	updateRoutes.POST("/", handler.RequireContentType("application/json"), h.UpdateJSON())
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", h.Update())

	return router
}

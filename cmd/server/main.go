package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func main() {
	cfg := config.ServerConfig{}
	if err := config.ParseServerConfig(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg config.ServerConfig) error {

	ms := metric.NewSet(storage.NewMemStorage())

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/", handler.AllMetricsHandler(*ms))
	valueRoutes := router.Group("/value")
	valueRoutes.GET("/:type/:name", handler.ValueHandler(*ms))
	updateRoutes := router.Group("/update")
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", handler.UpdateHandler(*ms))

	log.Println("server starting on ", cfg.Address)
	return http.ListenAndServe(cfg.Address.String(), router)
}

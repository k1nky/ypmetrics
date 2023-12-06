package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/k1nky/ypmetrics/internal/config"
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

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func parseConfig() (config.Keeper, error) {
	cfg := config.Keeper{}
	err := config.ParseKeeperConfig(&cfg)
	return cfg, err
}

func main() {
	l := logger.New()
	cfg, err := parseConfig()
	if err != nil {
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

func Run(l *logger.Logger, cfg config.Keeper) {
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
	router := newRouter(h, l, cfg.Key)
	if cfg.EnableProfiling {
		l.Infof("expose profiler on %s", DefaultProfilerPrefix)
		exposeProfiler(router)
	}

	l.Infof("starting on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address.String(), router); err != nil {
		panic(err)
	}
}

func newRouter(h handler.Handler, l *logger.Logger, key string) *gin.Engine {
	router := gin.New()
	// логируем запрос
	router.Use(middleware.Logger(l))
	if len(key) > 0 {
		// если указан ключ, то проверяем подпись полученных данных
		router.Use(middleware.NewSeal(key).Use())
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

func showVersion() {
	s := strings.Builder{}
	fmt.Fprintf(&s, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&s, "Build date: %s\n", buildDate)
	fmt.Fprintf(&s, "Build commit: %s\n", buildCommit)
	fmt.Println(s.String())
}

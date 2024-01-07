package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/crypto"
	grpchandler "github.com/k1nky/ypmetrics/internal/handler/grpc"
	grpcmw "github.com/k1nky/ypmetrics/internal/handler/grpc/middleware"
	httphandler "github.com/k1nky/ypmetrics/internal/handler/http"
	httpmw "github.com/k1nky/ypmetrics/internal/handler/http/middleware"
	"github.com/k1nky/ypmetrics/internal/logger"
	pb "github.com/k1nky/ypmetrics/internal/proto"
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

func newRouter(h httphandler.Handler, l *logger.Logger, sealKey string, decryptKey *rsa.PrivateKey, trustedSubnet *net.IPNet) *gin.Engine {
	router := gin.New()
	// логируем запрос
	router.Use(httpmw.Logger(l))
	if trustedSubnet != nil {
		// проверяем адрес источника запроса
		router.Use(httpmw.XRealIP(*trustedSubnet))
	}
	if len(sealKey) > 0 {
		// если указан ключ, то проверяем подпись полученных данных
		router.Use(httpmw.NewSeal(sealKey).Use())
	}
	if decryptKey != nil {
		// указан ключ шифрования, то расшифровываем тело запроса
		router.Use(httpmw.NewDecrypter(decryptKey).Use())
	}
	// при необходимости раcпаковываем/запаковываем данные
	router.Use(httpmw.NewGzip([]string{"application/json", "text/html"}).Use())

	router.GET("/", h.AllMetrics())
	router.GET("/ping", h.Ping())
	router.POST("/updates/", httpmw.RequireContentType("application/json"), h.UpdatesJSON())

	valueRoutes := router.Group("/value")
	valueRoutes.POST("/", httpmw.RequireContentType("application/json"), h.ValueJSON())
	valueRoutes.GET("/:type/:name", h.Value())

	updateRoutes := router.Group("/update")
	updateRoutes.POST("/", httpmw.RequireContentType("application/json"), h.UpdateJSON())
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", h.Update())

	return router
}

func newRPCServer(h *grpchandler.Handler, sealKey string, trustedSubnet *net.IPNet, l *logger.Logger) *grpc.Server {
	unaryInterceptors := []grpc.UnaryServerInterceptor{grpcmw.LoggerUnaryInterceptor(l)}
	streamInterceptors := []grpc.StreamServerInterceptor{grpcmw.LoggerStreamInterceptor(l)}
	if trustedSubnet != nil {
		unaryInterceptors = append(unaryInterceptors, grpcmw.XRealIPUnaryInterceptor(*trustedSubnet))
		streamInterceptors = append(streamInterceptors, grpcmw.XRealIPStreamInterceptor(*trustedSubnet))
	}
	if len(sealKey) > 0 {
		unaryInterceptors = append(unaryInterceptors, grpcmw.SealUnaryInterceptor(sealKey))
	}
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(unaryInterceptors...), grpc.ChainStreamInterceptor(streamInterceptors...))
	pb.RegisterMetricsServer(srv, h)
	return srv
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
	h := httphandler.New(uc)
	gh := grpchandler.New(uc)

	decryptKey, err := readCryptoKey(cfg.CryptoKey)
	if err != nil {
		l.Errorf("config: %s", err)
		exit(1)
	}
	trustedSubnet, err := cfg.TrustedSubnet.ToIPNet()
	if err != nil {
		l.Errorf("config: %s", err)
		exit(1)
	}
	router := newRouter(h, l, cfg.Key, decryptKey, trustedSubnet)

	if cfg.EnableProfiling {
		l.Infof("expose profiler on %s", DefaultProfilerPrefix)
		exposeProfiler(router)
	}

	l.Infof("starting http on %s", cfg.Address)
	runHTTPServer(ctx, cfg.Address.String(), router, l)
	if len(cfg.GRPCAddress.String()) != 0 {
		l.Infof("starting gRPC on %s", cfg.GRPCAddress)
		srv := newRPCServer(gh, cfg.Key, trustedSubnet, l)
		runRPCServer(ctx, cfg.GRPCAddress.String(), srv, l)
	}
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}

func runRPCServer(ctx context.Context, addr string, srv *grpc.Server, l *logger.Logger) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		l.Errorf("gRPC server failed to start %v", err)
		return
	}
	go func() {
		if err := srv.Serve(listen); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				l.Errorf("unexpected server closing: %v", err)
			}
		}
	}()
	go func() {
		<-ctx.Done()
		l.Debugf("closing gRPC server")
		srv.GracefulStop()
	}()
}

func runHTTPServer(ctx context.Context, addr string, h http.Handler, l *logger.Logger) {
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
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

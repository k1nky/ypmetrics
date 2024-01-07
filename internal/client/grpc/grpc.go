// Пакет grpc реализует grpc клиент для работы с сервером сбора метрик.
package grpc

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/k1nky/ypmetrics/internal/client/grpc/middleware"
	"github.com/k1nky/ypmetrics/internal/client/logger"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	pb "github.com/k1nky/ypmetrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	// MaxRequestTimeout максимальный таймаут для унарного запроса
	MaxRequestTimeout = 10 * time.Second
	// MaxQueueSize максимальный размер очереди на отправку метрик в потоке
	MaxQueueSize = 10
	// MaxAttempts максимальное количество попыток переподключения
	MaxAttempts = math.MaxInt
)

// Client grpc клиент сервера метрик
type Client struct {
	addr         string
	metricsQueue chan metric.Metrics
	cc           *grpc.ClientConn
	pushToStream bool
	key          string
	log          logger.Logger
}

// New возврщает grpc клиента сервера метрик
func New(addr string, l logger.Logger, key string, pushToStream bool) *Client {
	return &Client{
		addr:         addr,
		log:          l,
		key:          key,
		pushToStream: pushToStream,
		metricsQueue: make(chan metric.Metrics, MaxQueueSize),
	}
}

// Open устанавливает подключение к grpc серверу
func (c *Client) Open(ctx context.Context) error {
	var retryPolicy = fmt.Sprintf(`{
		"methodConfig": [{
			"name": [{"service": ""}],
			"waitForReady": true,
			"retryPolicy": {
				"MaxAttempts": %d,
				"InitialBackoff": ".01s",
				"MaxBackoff": ".01s",
				"BackoffMultiplier": 1.0,
				"RetryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}]
	}`, MaxAttempts)
	unaryInterceptors := []grpc.UnaryClientInterceptor{middleware.XRealIPUnaryInterceptor()}
	if c.key != "" {
		unaryInterceptors = append(unaryInterceptors, middleware.SealUnaryInterceptor(c.key))
	}
	conn, err := grpc.Dial(c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
		grpc.WithChainStreamInterceptor(middleware.XRealIPStreamInterceptor()),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
	)
	if err != nil {
		return err
	}
	c.cc = conn
	go func() {
		if err := c.streamMetrics(ctx); err != nil {
			c.log.Errorf("grpc client: %s", err)
		}
	}()
	return nil
}

// Close закрывает подключение к grpc серверу.
func (c *Client) Close() error {
	if c.cc != nil {
		close(c.metricsQueue)
		return c.cc.Close()
	}
	return nil
}

// PushCounter отправляет счетчик с именем name и значением value на сервер.
func (c *Client) PushCounter(name string, value int64) error {
	if c.pushToStream {
		// включена отправка в потоке, поэтому передаем метрику в очередь отправки
		metrics := metric.NewMetrics()
		metrics.Counters = append(metrics.Counters, metric.NewCounter(name, value))
		c.metricsQueue <- *metrics
		return nil
	}
	metricsClient := pb.NewMetricsClient(c.cc)
	ctx, cancel := context.WithTimeout(context.Background(), MaxRequestTimeout)
	defer cancel()
	_, err := metricsClient.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		Type:  pb.Type_COUNTER,
		Name:  name,
		Delta: value,
	})
	return err
}

// PushGauge отправляет измеритель с именем name и значением value на сервер.
func (c *Client) PushGauge(name string, value float64) error {
	if c.pushToStream {
		// включена отправка в потоке, поэтому передаем метрику в очередь отправки
		metrics := metric.NewMetrics()
		metrics.Gauges = append(metrics.Gauges, metric.NewGauge(name, value))
		c.metricsQueue <- *metrics
		return nil
	}
	metricsClient := pb.NewMetricsClient(c.cc)
	ctx, cancel := context.WithTimeout(context.Background(), MaxRequestTimeout)
	defer cancel()
	_, err := metricsClient.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		Type:  pb.Type_GAUGE,
		Name:  name,
		Value: value,
	})
	return err
}

func (c *Client) PushMetrics(metrics metric.Metrics) error {
	if c.pushToStream {
		c.metricsQueue <- metrics
		return nil
	}
	for _, m := range metrics.Counters {
		if err := c.PushCounter(m.Name, m.Value); err != nil {
			return err
		}
	}
	for _, m := range metrics.Gauges {
		if err := c.PushGauge(m.Name, m.Value); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) streamMetrics(ctx context.Context) error {
	metricsClient := pb.NewMetricsClient(c.cc)
	stream, err := metricsClient.UpdateMetrics(ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	for {
		select {
		case <-ctx.Done():
			return nil
		case in, ok := <-c.metricsQueue:
			if !ok {
				return nil
			}
			metrics := make([]*pb.UpdateMetricRequest, 0, len(in.Counters)+len(in.Gauges))
			for _, v := range in.Counters {
				metrics = append(metrics, &pb.UpdateMetricRequest{
					Type:  pb.Type_COUNTER,
					Name:  v.Name,
					Delta: v.Value,
				})
			}
			for _, v := range in.Gauges {
				metrics = append(metrics, &pb.UpdateMetricRequest{
					Type:  pb.Type_GAUGE,
					Name:  v.Name,
					Value: v.Value,
				})
			}
			if err := stream.Send(&pb.UpdateMetricsRequest{
				Mertics: metrics,
			}); err != nil {
				s, ok := status.FromError(err)
				if ok {
					c.log.Errorf("grpc client: stream metrics: %d %s", s.Code(), s.Message())
				} else {
					c.log.Errorf("grpc client: stream metrics: %s", err)
				}
			}
		}
	}
}

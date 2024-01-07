package client

import (
	"context"
	"crypto/rsa"
	"os"

	grpcclient "github.com/k1nky/ypmetrics/internal/client/grpc"
	httpclient "github.com/k1nky/ypmetrics/internal/client/http"
	"github.com/k1nky/ypmetrics/internal/client/logger"
	"github.com/k1nky/ypmetrics/internal/crypto"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

type Config struct {
	// CryptoKey путь до файл с публичным ключом сервера метрик
	CryptoKey string
	// GRPCAddress адрес grpc сервера метрик
	GRPCAddress string
	// GRPCPushToStream передавать метрики в потоке grpc
	GRPCPushToStream bool
	// HTTPAddress адрсе http сервера метрик
	HTTPAddress string
	// Key ключ подписи передаваемых данных
	Key string
}

type Client interface {
	Close() error
	PushCounter(name string, value int64) error
	PushGauge(name string, value float64) error
	PushMetrics(metrics metric.Metrics) error
}

func readCryptoKey(path string) (*rsa.PublicKey, error) {
	if len(path) == 0 {
		return nil, nil
	}
	f, err := os.Open(path)
	defer func() { _ = f.Close() }()
	if err != nil {
		return nil, err
	}
	key, err := crypto.ReadPublicKey(f)
	return key, err
}

// New метод-фабрика для клиента сервера метрик.
func New(ctx context.Context, cfg Config, l logger.Logger) (Client, error) {
	var client Client
	if cfg.GRPCAddress != "" {
		c := grpcclient.New(cfg.GRPCAddress, l, cfg.Key, cfg.GRPCPushToStream)
		if err := c.Open(ctx); err != nil {
			return nil, err
		}
		client = c
	} else {
		c := httpclient.New(cfg.HTTPAddress, l)
		key, err := readCryptoKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		// сжимаем данные -> шифруем -> подписываем
		c.SetGzip().SetEncrypt(key).SetKey(cfg.Key)
		client = c
	}
	return client, nil
}

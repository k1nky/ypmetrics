// Пакет apiclient реализует клиент для работы с сервером сбора метрик
package apiclient

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/metric"
)

const DefEndpointURL = "http://localhost:8080"

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

// Option опция конфигурации клиента сервера сбора метрик.
// Используются при создании нового клиента через функцию New.
type Option func(*Client)

// Client клиент для сервера сбора метрик
type Client struct {
	// URL сервера сбора метрик в формате <протокол>://<хост>[:порт]
	EndpointURL string
	httpclient  *resty.Client
}

// WithEndpointURL задает URL сервера сбора метрик
func WithEndpointURL(url string) Option {
	return func(c *Client) {
		c.EndpointURL = url
	}
}

// New возвращает нового клиента для сервера сбора метрик
func New(options ...Option) *Client {
	c := &Client{
		EndpointURL: DefEndpointURL,
		httpclient:  resty.New(),
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// UpdateMetric отправляет значение указанной метрики на сервер
func (c *Client) UpdateMetric(metric metric.Measure) (err error) {
	req := c.httpclient.R()
	req.Method = http.MethodPost
	if url, err := url.JoinPath(c.EndpointURL, "update", metric.String()); err != nil {
		return err
	} else {
		req.URL = url
	}
	resp, err := req.Send()
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return ErrUnexpectedStatusCode
	}

	return
}

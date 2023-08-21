// Пакет apiclient реализует клиент для работы с сервером сбора метрик
package apiclient

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/protocol"
)

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

// Client клиент для сервера сбора метрик
type Client struct {
	// URL сервера сбора метрик в формате <протокол>://<хост>[:порт]
	EndpointURL string
	httpclient  *resty.Client
}

// New возвращает нового клиента для сервера сбора метрик
func New(url string) *Client {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	c := &Client{
		EndpointURL: url,
		httpclient:  resty.New().SetTimeout(5 * time.Second),
	}
	return c
}

// PushMetric отправляет метрику на сервер
func (c *Client) PushMetric(typ, name, value string) (err error) {
	var (
		requestURL string
		resp       *resty.Response
	)
	if requestURL, err = url.JoinPath(c.EndpointURL, "update", typ, name, value); err != nil {
		return err
	}
	if resp, err = c.httpclient.R().Post(requestURL); err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return ErrUnexpectedStatusCode
	}

	return
}

func (c *Client) PushCounter(name string, value int64) (err error) {
	var (
		requestURL string
		resp       *resty.Response
	)

	if requestURL, err = url.JoinPath(c.EndpointURL, "update/"); err != nil {
		return
	}
	if resp, err = c.httpclient.R().
		SetHeader("content-type", "application/json").
		SetBody(protocol.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &value,
		}).
		Post(requestURL); err != nil {
		return
	}
	if resp.StatusCode() != http.StatusOK {
		return ErrUnexpectedStatusCode
	}
	return nil
}

func (c *Client) PushGauge(name string, value float64) (err error) {
	var (
		requestURL string
		resp       *resty.Response
	)

	if requestURL, err = url.JoinPath(c.EndpointURL, "update/"); err != nil {
		return
	}
	if resp, err = c.httpclient.R().
		SetHeader("content-type", "application/json").
		SetBody(protocol.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}).
		Post(requestURL); err != nil {
		return
	}
	if resp.StatusCode() != http.StatusOK {
		return ErrUnexpectedStatusCode
	}
	return nil
}

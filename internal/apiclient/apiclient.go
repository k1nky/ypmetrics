// Пакет apiclient реализует клиент для работы с сервером сбора метрик
package apiclient

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
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
		httpclient:  resty.New(),
	}
	return c
}

// PushMetric отправляет метрику на сервер
func (c *Client) PushMetric(typ, name, value string) (err error) {
	req := c.httpclient.R()
	req.Method = http.MethodPost
	if url, err := url.JoinPath(c.EndpointURL, "update", typ, name, value); err != nil {
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

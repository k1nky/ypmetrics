// Пакет apiclient реализует клиент для работы с сервером сбора метрик
package apiclient

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/protocol"
	"github.com/k1nky/ypmetrics/internal/retrier"
)

const (
	DefaultRequestTimeout = 5 * time.Second
	CounterType           = "counter"
	GaugeType             = "gauge"
)

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

// Client клиент для сервера сбора метрик
type Client struct {
	// URL сервера сбора метрик в формате <протокол>://<хост>[:порт]
	EndpointURL string
	Retries     []time.Duration
	httpclient  *resty.Client
}

// New возвращает нового клиента для сервера сбора метрик
func New(url string) *Client {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	c := &Client{
		EndpointURL: url,
		Retries:     []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
		httpclient:  resty.New().SetTimeout(DefaultRequestTimeout),
	}
	return c
}

// newRequest это shortcut для создания нового запроса
func (c *Client) newRequest() *resty.Request {
	return c.httpclient.R().SetHeader("accept-encoding", "gzip")
}

// PushMetric отправляет метрику на сервер
func (c *Client) PushMetric(typ, name, value string) error {

	path, err := url.JoinPath("update/", typ, name, value)
	if err != nil {
		return err
	}
	return c.postData(path, "text/plain", nil)
}

// PushCounter отправляет счетчик на сервер в формате JSON
func (c *Client) PushCounter(name string, value int64) error {
	return c.postData("update/", "application/json", protocol.Metrics{
		ID:    name,
		MType: CounterType,
		Delta: &value,
	})
}

// PushGauge отправляет измеритель на сервер в формате JSON
func (c *Client) PushGauge(name string, value float64) (err error) {
	return c.postData("update/", "application/json", protocol.Metrics{
		ID:    name,
		MType: GaugeType,
		Value: &value,
	})
}

// PushMetrics отправляет метрики на сервер в формате JSON
func (c *Client) PushMetrics(metrics metric.Metrics) (err error) {
	metricsCount := len(metrics.Counters) + len(metrics.Gauges)
	if metricsCount == 0 {
		return nil
	}
	m := make([]protocol.Metrics, 0, metricsCount)
	for _, c := range metrics.Counters {
		m = append(m, protocol.Metrics{ID: c.Name, MType: CounterType, Delta: &c.Value})
	}
	for _, g := range metrics.Gauges {
		m = append(m, protocol.Metrics{ID: g.Name, MType: GaugeType, Value: &g.Value})
	}
	return c.postData("updates/", "application/json", m)
}

// Отправляет POST запрос по пути path с типом контента contentType и телом body
func (c *Client) postData(path string, contentType string, body interface{}) (err error) {
	var (
		requestURL string
		resp       *resty.Response
	)
	if requestURL, err = url.JoinPath(c.EndpointURL, path); err != nil {
		return err
	}
	// формируем запрос
	request := c.newRequest().SetHeader("content-type", contentType).SetBody(body)
	request.Method = http.MethodPost
	request.URL = requestURL
	if resp, err = c.send(request); err != nil {
		return err
	}
	// код ответа отличный от 200 не будем считать ошибкой отправки данных
	if resp.StatusCode() != http.StatusOK {
		return ErrUnexpectedStatusCode
	}
	return nil
}

// Отправляет сформированный запрос на сервер. Если при отправке возникнут ошибки,
// запрос будет отправлен повторно.
func (c *Client) send(request *resty.Request) (response *resty.Response, err error) {

	for r := retrier.New(c.shouldRetry); r.Next(err); {
		response, err = request.Send()
	}
	return
}

// Определяет условие, при котором неуспешно отправленный запрос должен быть отправлен повторно.
func (c *Client) shouldRetry(err error) bool {
	var e *url.Error

	// ошибка при отправке запроса скорее всего будет ошибкой транспорта, поэтому можно всегда повторно
	// отправлять запрос. В рамках данного проекта, не будем повторять запрос, если возникла разбора запроса.
	if errors.As(err, &e) {
		if e.Op == "parse" {
			return false
		}
	}
	return true
}

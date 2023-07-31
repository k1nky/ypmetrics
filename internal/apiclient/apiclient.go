package apiclient

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/metric"
)

const DefBaseURL = "http://localhost:8080"

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

type Option func(*Client)

type Client struct {
	BaseURL    string
	httpclient *resty.Client
}

func WithBaseURL(base string) Option {
	return func(c *Client) {
		c.BaseURL = base
	}
}

func New(options ...Option) *Client {
	c := &Client{
		BaseURL:    DefBaseURL,
		httpclient: resty.New(),
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Client) UpdateMetric(metric metric.Measure) (err error) {
	req := c.httpclient.R()
	req.Method = http.MethodPost
	if url, err := url.JoinPath(c.BaseURL, "update", metric.String()); err != nil {
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

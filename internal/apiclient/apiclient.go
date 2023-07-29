package apiclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/metric"
)

const DefBaseURL = "http://localhost:8080"

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

type Option func(*Client)

type Client struct {
	BaseURL string
}

func WithBaseURL(base string) Option {
	return func(c *Client) {
		c.BaseURL = base
	}
}

func New(options ...Option) *Client {
	c := &Client{
		BaseURL: DefBaseURL,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Client) UpdateMetric(metric metric.Measure) (err error) {
	req := new(http.Request)
	if req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/update/%s", c.BaseURL, metric), nil); err != nil {
		return
	}
	req.Header.Add("content-type", "plain/text")
	_, err = c.doRequest(req)
	return
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnexpectedStatusCode
	}
	return body, nil
}

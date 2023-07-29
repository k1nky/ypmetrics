package apiclient

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/metric"
)

const DefBaseUrl = "http://localhost:8080"

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code, want 200")
)

type Option func(*Client)

type Client struct {
	BaseUrl string
}

func WithBaseUrl(base string) Option {
	return func(c *Client) {
		c.BaseUrl = base
	}
}

func New(options ...Option) *Client {
	c := &Client{
		BaseUrl: DefBaseUrl,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Client) UpdateMetric(metric metric.Measure) (err error) {
	req := new(http.Request)
	if req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.BaseUrl, metric), nil); err != nil {
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnexpectedStatusCode
	}
	return body, nil
}

package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSignRequestSHA256(t *testing.T) {
	httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h := req.Header.Get("HashSHA256")
		assert.Len(t, h, 64)
		assert.Equal(t, "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b", h)
		rw.WriteHeader(http.StatusOK)
	}))
	defer httpserver.Close()
	cli := resty.NewWithClient(httpserver.Client())
	cli.SetPreRequestHook(SignRequestSHA256("secret"))
	r := cli.R().SetBody("hello")
	r.Post(httpserver.URL)
}

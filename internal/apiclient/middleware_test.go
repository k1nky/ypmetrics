package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSignRequestSHA256(t *testing.T) {
	tests := []struct {
		data string
		hash string
	}{
		{data: "hello", hash: "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b"},
		{data: "hello world", hash: "734cc62f32841568f45715aeb9f4d7891324e6d948e4c6c60c0621cdac48623a"},
	}

	seal := NewSeal("secret")
	for _, tt := range tests {
		httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			h := req.Header.Get("HashSHA256")
			assert.Len(t, h, 64)
			assert.Equal(t, tt.hash, h)
			rw.WriteHeader(http.StatusOK)
		}))

		cli := resty.NewWithClient(httpserver.Client())
		cli.SetPreRequestHook(seal.Use())
		r := cli.R().SetBody(tt.data)
		r.Post(httpserver.URL)
		httpserver.Close()
	}
}

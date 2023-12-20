package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEncryptRequest(t *testing.T) {
	tests := []struct {
		data []byte
	}{
		{data: bytes.Repeat([]byte("abcdef12345"), 1)},
		{data: bytes.Repeat([]byte("abcdef12345"), 100)},
	}

	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	e := NewEncrypter(&key.PublicKey)
	for _, tt := range tests {
		httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body := bytes.NewBuffer(nil)
			body.ReadFrom(req.Body)
			decrypted, err := crypto.DecryptRSA(key, body.Bytes())
			assert.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
			rw.WriteHeader(http.StatusOK)
		}))

		cli := resty.NewWithClient(httpserver.Client())
		cli.SetPreRequestHook(e.Use())
		r := cli.R().SetBody(tt.data)
		r.Post(httpserver.URL)
		httpserver.Close()
	}
}

func TestDecryptWithoutTrueKey(t *testing.T) {
	data := bytes.Repeat([]byte("abcdef12345"), 100)
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	e := NewEncrypter(&key.PublicKey)
	httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		key, _ := rsa.GenerateKey(rand.Reader, 4096)
		body := bytes.NewBuffer(nil)
		body.ReadFrom(req.Body)
		decrypted, err := crypto.DecryptRSA(key, body.Bytes())
		assert.Error(t, err)
		assert.NotEqual(t, data, decrypted)
		rw.WriteHeader(http.StatusOK)
	}))

	cli := resty.NewWithClient(httpserver.Client())
	cli.SetPreRequestHook(e.Use())
	r := cli.R().SetBody(data)
	r.Post(httpserver.URL)
	httpserver.Close()
}

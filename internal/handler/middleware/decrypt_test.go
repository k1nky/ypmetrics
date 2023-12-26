package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/crypto"
	"github.com/stretchr/testify/assert"
)

// func TestDecrypt(t *testing.T) {
// 	key, _ := rsa.GenerateKey(rand.Reader, 4096)
// 	plainData := bytes.Repeat([]byte("abcdef12345"), 100)
// 	encryptedData, _ := crypto.EncryptRSA(&key.PublicKey, plainData)

// 	tests := []struct {
// 		name       string
// 		wantStatus int
// 		data       []byte
// 	}{
// 		{
// 			name:       "With valid hash",
// 			wantStatus: http.StatusOK,
// 			data:       encryptedData,
// 		},
// 		{
// 			name:       "With invalid hash",
// 			wantStatus: http.StatusBadRequest,
// 			data:       []byte("bye"),
// 		},
// 		{
// 			name:       "Without header",
// 			wantStatus: http.StatusOK,
// 			data:       []byte{},
// 		},
// 	}
// 	gin.SetMode(gin.TestMode)
// 	for _, tt := range tests {
// 		w := httptest.NewRecorder()
// 		c, r := gin.CreateTestContext(w)
// 		r.Any("/", NewSeal("secret").Use(), func(c *gin.Context) {
// 			c.Status(http.StatusOK)
// 		})
// 		buf := &bytes.Buffer{}
// 		buf.WriteString(tt.data)
// 		c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
// 		if len(tt.hashValue) > 0 {
// 			c.Request.Header.Set("HashSHA256", tt.hashValue)
// 		}
// 		r.ServeHTTP(w, c.Request)
// 		result := w.Result()
// 		defer result.Body.Close()
// 		assert.Equal(t, tt.wantStatus, result.StatusCode, tt.name)
// 	}
// }

func TestDecryptWithBody(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	plainData := bytes.Repeat([]byte("abcdef12345"), 100)
	encryptedData, _ := crypto.EncryptRSA(&key.PublicKey, plainData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Any("/", NewDecrypter(key).Use(), func(c *gin.Context) {
		body := bytes.NewBuffer(nil)
		body.ReadFrom(c.Request.Body)
		assert.Equal(t, plainData, body.Bytes())
		c.Status(http.StatusOK)
	})
	buf := bytes.NewBuffer(encryptedData)
	c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
	r.ServeHTTP(w, c.Request)
	result := w.Result()
	defer result.Body.Close()
	assert.Equal(t, http.StatusOK, result.StatusCode)
}

func TestDecryptWithoutBody(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	plainData := bytes.Repeat([]byte(""), 1)
	encryptedData, _ := crypto.EncryptRSA(&key.PublicKey, plainData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Any("/", NewDecrypter(key).Use(), func(c *gin.Context) {
		body := bytes.NewBuffer(nil)
		body.ReadFrom(c.Request.Body)
		assert.Equal(t, plainData, body.Bytes())
		c.Status(http.StatusOK)
	})
	buf := bytes.NewBuffer(encryptedData)
	c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
	r.ServeHTTP(w, c.Request)
	result := w.Result()
	defer result.Body.Close()
	assert.Equal(t, http.StatusOK, result.StatusCode)
}

func TestDecryptInvalidBody(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Any("/", NewDecrypter(key).Use(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	buf := bytes.NewBuffer([]byte("invalid body"))
	c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
	r.ServeHTTP(w, c.Request)
	result := w.Result()
	defer result.Body.Close()
	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
}

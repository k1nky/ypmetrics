package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/crypto"
)

// Decrypter это middleware для асимметричного шифрования тела запроса.
type Decrypter struct {
	buffers sync.Pool
	key     *rsa.PrivateKey
}

// NewDecrypter возвращает новую middleware для расшифрования закрытым ключом key тела запроса.
func NewDecrypter(key *rsa.PrivateKey) *Decrypter {
	return &Decrypter{
		buffers: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
		key: key,
	}
}

// Use расшифровывает тело запроса.
// Применимо для POST запросов с непустым телом.
func (d *Decrypter) Use() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !d.shouldUse(ctx.Request) {
			return
		}
		buf := d.buffers.Get().(*bytes.Buffer)
		defer d.buffers.Put(buf)
		buf.Reset()

		if _, err := buf.ReadFrom(ctx.Request.Body); err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		body, err := crypto.DecryptRSA(d.key, buf.Bytes())
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	}
}

// Определяет потребность в расшировании запроса.
func (d *Decrypter) shouldUse(r *http.Request) bool {
	if r.ContentLength != 0 && r.Method == http.MethodPost {
		return true
	}
	return false
}

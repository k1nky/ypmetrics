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

// Seal это middleware для подписи тела запроса.
// Подпись будет проставляться в заголовок HashSHA256.
type Decrypter struct {
	buffers sync.Pool
	key     *rsa.PrivateKey
}

// NewSeal возвращает новую middleware для подписи с ключом secret.
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

// Use добавляет заголовок HashSHA256 с подписью передаваемых данных по алгоритму sha256.
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
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	}
}

// Определяет потребность в формировании подписи для указаного запроса
func (d *Decrypter) shouldUse(r *http.Request) bool {
	if r.ContentLength != 0 && r.Method == http.MethodPost {
		return true
	}
	return false
}

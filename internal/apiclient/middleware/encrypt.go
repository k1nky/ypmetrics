package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/k1nky/ypmetrics/internal/crypto"
)

// Encrypter это middleware для асимметричного шифрования тела запроса.
type Encrypter struct {
	buffers sync.Pool
	key     *rsa.PublicKey
}

// NewEncrypter возвращает новую middleware для шифрования открытым ключом key.
func NewEncrypter(key *rsa.PublicKey) *Encrypter {
	return &Encrypter{
		buffers: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
		key: key,
	}
}

// Use шифрует тело запроса.
// Применимо для POST запросов с непустым телом.
func (e *Encrypter) Use() resty.PreRequestHook {
	return func(c *resty.Client, r *http.Request) error {
		if !e.shouldUse(r) {
			return nil
		}
		buf := e.buffers.Get().(*bytes.Buffer)
		defer e.buffers.Put(buf)
		buf.Reset()

		if _, err := buf.ReadFrom(r.Body); err != nil {
			return err
		}
		body, err := crypto.EncryptRSA(e.key, buf.Bytes())
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		// обновляем размер передаваемых данных
		r.ContentLength = int64(len(body))

		return nil
	}
}

// Определяет потребность в шифровании тела запроса.
func (e *Encrypter) shouldUse(r *http.Request) bool {
	if r.ContentLength != 0 && r.Method == http.MethodPost {
		return true
	}
	return false
}

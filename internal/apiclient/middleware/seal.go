package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

// Seal это middleware для подписи тела запроса.
// Подпись будет проставляться в заголовок HashSHA256.
type Seal struct {
	hashers sync.Pool
}

// NewSeal возвращает новую middleware для подписи с ключом secret.
func NewSeal(secret string) *Seal {
	return &Seal{
		hashers: sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, []byte(secret))
			},
		},
	}
}

// Use добавляет заголовок HashSHA256 с подписью передаваемых данных по алгоритму sha256.
// Применимо для POST запросов с непустым телом.
func (s *Seal) Use() resty.PreRequestHook {
	return func(c *resty.Client, r *http.Request) error {
		if !s.shouldSign(r) {
			return nil
		}

		h := s.hashers.Get().(hash.Hash)
		defer s.hashers.Put(h)
		h.Reset()

		buf := io.TeeReader(r.Body, h)
		body := bytes.NewBuffer(nil)
		if _, err := body.ReadFrom(buf); err != nil {
			return err
		}
		r.Body.Close()
		r.Body = io.NopCloser(body)
		r.Header.Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))

		return nil
	}
}

// Определяет потребность в формировании подписи для указаного запроса
func (s *Seal) shouldSign(r *http.Request) bool {
	if r.ContentLength != 0 && r.Method == http.MethodPost {
		return true
	}
	return false
}

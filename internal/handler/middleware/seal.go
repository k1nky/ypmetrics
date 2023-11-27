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

	"github.com/gin-gonic/gin"
)

// Seal middleware для подписи отправляемых данных и проверки подписи получаемых данных.
type Seal struct {
	hashers sync.Pool
}

func NewSeal(key string) *Seal {
	return &Seal{
		hashers: sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, []byte(key))
			},
		},
	}
}

// Use формирует подпись отправляемых данных и проверяет подпись получаемых данных.
// Подпись должна быть указана в заголовке HashSHA256.
func (s *Seal) Use() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if seal := ctx.Request.Header.Get("HashSHA256"); len(seal) > 0 {
			if valid, err := s.verify(ctx.Request, seal); !valid || err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		bw := &bufferWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: ctx.Writer,
		}
		ctx.Writer = bw
		ctx.Next()
		h, err := s.sign(bw.body.Bytes())
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		bw.Header().Set("HashSHA256", h)
		bw.ResponseWriter.Write(bw.body.Bytes())
	}
}

func (s *Seal) verify(req *http.Request, seal string) (bool, error) {
	h := s.hashers.Get().(hash.Hash)
	defer s.hashers.Put(h)
	h.Reset()

	buf := io.TeeReader(req.Body, h)
	body := bytes.NewBuffer(nil)
	if _, err := body.ReadFrom(buf); err != nil {
		return false, err
	}
	req.Body.Close()
	req.Body = io.NopCloser(body)

	got := hex.EncodeToString(h.Sum(nil))
	return seal == got, nil
}

func (s *Seal) sign(data []byte) (string, error) {
	h := s.hashers.Get().(hash.Hash)
	defer s.hashers.Put(h)
	h.Reset()

	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

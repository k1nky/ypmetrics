package apiclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

type Seal struct {
	hashers sync.Pool
}

func NewSeal(secret string) *Seal {
	return &Seal{
		hashers: sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, []byte(secret))
			},
		},
	}
}

// Добавляет заголовок HashSHA256 с подписью передаваемых данных по алгоритму sha256.
//
//	Лучше для этих целей использовать RequestMiddleware и передавать его в метод OnBeforeRequest,
//	но в нем в Body лежит interface{} для которого не удобно считать подпись,
//	а RawRequest.Body будет всегда nil (https://github.com/go-resty/resty/issues/517). Поэтому
//	используем PreRequestHook, однако в актуальной версии он может быть только один.
//	https://github.com/go-resty/resty/issues/665
func (s *Seal) Use() resty.PreRequestHook {
	return func(c *resty.Client, r *http.Request) error {
		h := s.hashers.Get().(hash.Hash)
		defer s.hashers.Put(h)
		h.Reset()

		if _, err := io.Copy(h, r.Body); err != nil {
			return err
		}
		r.Header.Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))
		return nil
	}
}

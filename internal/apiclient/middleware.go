package apiclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// https://github.com/go-resty/resty/issues/665
// https://github.com/go-resty/resty/issues/517
func SignRequestSHA256(secret string) resty.PreRequestHook {
	return func(c *resty.Client, r *http.Request) error {
		h := hmac.New(sha256.New, []byte(secret))
		buf := bytes.Buffer{}
		buf.ReadFrom(r.Body)
		if _, err := h.Write(buf.Bytes()); err != nil {
			return err
		}
		r.Header.Set("HashSHA256", string(h.Sum(nil)))
		return nil
	}
}

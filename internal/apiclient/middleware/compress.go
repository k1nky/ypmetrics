package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

type compressor struct {
	w   *gzip.Writer
	buf *bytes.Buffer
}

func (c *compressor) Reset() {
	c.buf.Reset()
	c.w.Reset(c.buf)
}

type Gzip struct {
	compressors sync.Pool
}

func NewGzip() *Gzip {
	return &Gzip{
		compressors: sync.Pool{
			New: func() any {
				return &compressor{
					w:   gzip.NewWriter(io.Discard),
					buf: bytes.NewBuffer(nil),
				}
			},
		},
	}
}

func (gz *Gzip) Use() resty.PreRequestHook {
	return func(c *resty.Client, r *http.Request) error {
		if !gz.shouldCompress(r) {
			return nil
		}

		z := gz.compressors.Get().(*compressor)
		defer gz.compressors.Put(z)
		z.Reset()

		body := bytes.NewBuffer(nil)
		if _, err := body.ReadFrom(r.Body); err != nil {
			return err
		}
		r.Body.Close()
		if _, err := z.w.Write(body.Bytes()); err != nil {
			return err
		}
		z.w.Close()
		body.Reset()
		if _, err := body.ReadFrom(z.buf); err != nil {
			return err
		}

		r.Body = io.NopCloser(body)
		r.ContentLength = int64(body.Len())
		r.Header.Set("content-encoding", "gzip")
		return nil
	}
}

func (gz *Gzip) shouldCompress(r *http.Request) bool {
	if r.ContentLength != 0 && r.Method == http.MethodPost {
		return true
	}
	return false
}

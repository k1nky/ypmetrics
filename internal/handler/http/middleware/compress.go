package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Gzip middleware для сжатия и расжатия тела запроса
type Gzip struct {
	contentTypes []string
	compressors  sync.Pool
}

// NewGzip возвращает новый экземпляр middleware для сжатия и расжатия тела запроса
func NewGzip(contentTypes []string) *Gzip {
	return &Gzip{
		contentTypes: contentTypes,
		compressors: sync.Pool{
			New: func() any {
				w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
				return w
			},
		},
	}
}

// Gzip middleware позволяет разжимать тело запроса и сжимать тело ответа.
// Тело запроса будет разжато, если указан заголовок content-encoding: gzip.
//
// Сжатие тела ответа будет выполняться при истиности следующих условий:
// 1) клиент поддерживает сжатие (заголовок accept-encoding);
// 2) тип контента ответа разрешен для сжатия (contentTypes).
//
// Если список разрешенных типов пустой, то сжимать можно тело с любым типом.
func (gh *Gzip) Use() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if gh.shouldUncompress(ctx.Request) {
			// требуется разжатие тела запроса, поэтому подменяем тело запроса
			gz, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			ctx.Request.Body = gz
		}

		gzwriter := &bufferWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}
		ctx.Writer = gzwriter
		// для всех последующих обработчиков тело ответа будем писать в поле `body` gzwriter,
		// т.к. пока не отработают все обработчики, нельзя достоверно определить потребуется ли
		// сжатие ответа (нужно проверить заголовок Content-Type)
		ctx.Next()
		if gh.shouldCompress(ctx.Request, gzwriter) {
			gzwriter.Header().Add("Content-Encoding", "gzip")
			gz := gh.compressors.Get().(*gzip.Writer)
			defer gh.compressors.Put(gz)

			gz.Reset(gzwriter.ResponseWriter)
			// сжатие требуется, в ответ пишем сжатые данные
			if _, err := gz.Write(gzwriter.body.Bytes()); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			_ = gz.Close()
		} else {
			// сжатие не требуется, в ответ пишем данные как есть
			if _, err := gzwriter.ResponseWriter.Write(gzwriter.body.Bytes()); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}
}

// shouldCompress определяет требуется ли сжатие тела ответа.
func (gh *Gzip) shouldCompress(request *http.Request, response http.ResponseWriter) bool {
	if !strings.Contains(request.Header.Get("accept-encoding"), "gzip") {
		return false
	}
	contentType := response.Header().Get("content-type")
	if len(gh.contentTypes) == 0 {
		return true
	}
	for _, ct := range gh.contentTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}
	return false
}

// shouldUncompress определяет требуется ли разжатие тела запроса
func (gh *Gzip) shouldUncompress(r *http.Request) bool {
	return strings.Contains(r.Header.Get("content-encoding"), "gzip")
}

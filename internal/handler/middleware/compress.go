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

type bufferWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (gz *bufferWriter) WriteString(s string) (int, error) {
	return gz.Write([]byte(s))
}

func (gz *bufferWriter) Write(data []byte) (int, error) {
	gz.Header().Del("Content-Length")
	return gz.body.Write(data)
}

func (gz *bufferWriter) WriteHeader(code int) {
	gz.Header().Del("Content-Length")
	gz.ResponseWriter.WriteHeader(code)
}

// shouldCompress определяет требуется ли сжатие тела ответа.
func (gh *GzipHandler) shouldCompress(request *http.Request, response http.ResponseWriter) bool {
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
func (gh *GzipHandler) shouldUncompress(r *http.Request) bool {
	return strings.Contains(r.Header.Get("content-encoding"), "gzip")
}

type GzipHandler struct {
	contentTypes []string
	compressors  sync.Pool
}

func NewGzip(contentTypes []string) *GzipHandler {
	return &GzipHandler{
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
// Сжатие тела ответа будет выполняться при истиности следующих условий:
//
//	клиент поддерживает сжатие (заголовок accept-encoding);
//	тип контента ответа разрешен для сжатия (contentTypes).
//
// Если список разрешенных типов пустой, то сжимать можно тело с любым типом.
func (gh *GzipHandler) Use() gin.HandlerFunc {
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
			gz.Write(gzwriter.body.Bytes())
			gz.Close()
		} else {
			// сжатие не требуется, в ответ пишем данные как есть
			gzwriter.ResponseWriter.Write(gzwriter.body.Bytes())
		}
	}
}

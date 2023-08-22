package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	contentTypes []string
	body         *bytes.Buffer
}

func (gz *gzipWriter) WriteString(s string) (int, error) {
	return gz.Write([]byte(s))
}

func (gz *gzipWriter) Write(data []byte) (int, error) {
	gz.Header().Del("Content-Length")
	return gz.body.Write(data)
}

func (gz *gzipWriter) WriteHeader(code int) {
	gz.Header().Del("Content-Length")
	gz.ResponseWriter.WriteHeader(code)
}

// shouldCompress определяет требуется ли сжатие тела ответа.
func (gz *gzipWriter) shouldCompress(request *http.Request, response http.ResponseWriter) bool {
	if !strings.Contains(request.Header.Get("accept-encoding"), "gzip") {
		return false
	}
	contentType := response.Header().Get("content-type")
	if len(gz.contentTypes) == 0 {
		return true
	}
	for _, ct := range gz.contentTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}
	return false
}

// shouldUncompress определяет требуется ли разжатие тела запроса
func shouldUncompress(r *http.Request) bool {
	return strings.Contains(r.Header.Get("content-encoding"), "gzip")
}

// Gzip middleware позволяет разжимать тело запроса и сжимать тело ответа.
// Тело запроса будет разжато, если указан заголовок content-encoding: gzip.
// Сжатие тела ответа будет выполняться при истиности следующих условий:
//
//	клиент поддерживает сжатие (заголовок accept-encoding);
//	тип контента ответа разрешен для сжатия (contentTypes).
//
// Если список разрешенных типов пустой, то сжимать можно тело с любым типом.
func Gzip(contentTypes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if shouldUncompress(ctx.Request) {
			// требуется разжатие тела запроса, поэтому подменяем тело запроса
			gz, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			ctx.Request.Body = gz
		}

		gzwriter := &gzipWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
			contentTypes:   contentTypes,
		}
		ctx.Writer = gzwriter
		// для всех последующих обработчиков тело ответа будем писать в поле `body` gzwriter,
		// т.к. пока не отработают все обработчики, нельзя достоверно определить потребуется ли
		// сжатие ответа (нужно проверить заголовок Content-Type)
		ctx.Next()
		if gzwriter.shouldCompress(ctx.Request, gzwriter) {
			gzwriter.Header().Add("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(gzwriter.ResponseWriter, gzip.BestSpeed)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			// сжатие требуется, в ответ пишем сжатые данные
			gz.Write(gzwriter.body.Bytes())
		} else {
			// сжатие не требуется, в ответ пишем данные как есть
			gzwriter.ResponseWriter.Write(gzwriter.body.Bytes())
		}
	}
}

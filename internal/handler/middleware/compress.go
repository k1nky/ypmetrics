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

func (gzw *gzipWriter) shouldCompress(request *http.Request, response http.ResponseWriter) bool {
	if !strings.Contains(request.Header.Get("accept-encoding"), "gzip") {
		return false
	}
	contentType := response.Header().Get("content-type")
	if len(gzw.contentTypes) == 0 {
		return true
	}
	for _, ct := range gzw.contentTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}
	return false
}

func Gzip(contentTypes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if shouldUncompress(ctx.Request) {
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
		ctx.Next()
		if gzwriter.shouldCompress(ctx.Request, gzwriter) {
			gzwriter.Header().Add("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(gzwriter.ResponseWriter, gzip.BestSpeed)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			gz.Write(gzwriter.body.Bytes())
		} else {
			gzwriter.ResponseWriter.Write(gzwriter.body.Bytes())
		}
	}
}

func shouldUncompress(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
}

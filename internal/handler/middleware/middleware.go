package middleware

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

type bufferWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// RequireContentType это middleware, который определяет требование для значения заголовка ContentType
func RequireContentType(contentType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.ContentType() != contentType {
			ctx.AbortWithStatus(http.StatusBadRequest)
		}
		ctx.Next()
	}
}

func (bw *bufferWriter) WriteString(s string) (int, error) {
	return bw.Write([]byte(s))
}

func (bw *bufferWriter) Write(data []byte) (int, error) {
	bw.Header().Del("Content-Length")
	return bw.body.Write(data)
}

func (bw *bufferWriter) WriteHeader(code int) {
	bw.Header().Del("Content-Length")
	bw.ResponseWriter.WriteHeader(code)
}

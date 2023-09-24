package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type requestLogger interface {
	Infof(template string, args ...interface{})
}

// Logger это middleware для  логирования запросов и ответов
func Logger(l requestLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()
		size := ctx.Writer.Size()
		if size < 0 {
			size = 0
		}
		l.Infof("%s %s status %d size %d duration %s", ctx.Request.Method, ctx.Request.RequestURI, ctx.Writer.Status(), size, time.Since(start))
	}
}

// LoggerWithBody это middleware для  логирования запросов (включая тело запроса) и ответов
func LoggerWithBody(l requestLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		var buf bytes.Buffer
		tee := io.TeeReader(ctx.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		ctx.Request.Body = io.NopCloser(&buf)

		ctx.Next()
		size := ctx.Writer.Size()
		if size < 0 {
			size = 0
		}
		l.Infof("%s %s status %d size %d duration %s %s", ctx.Request.Method, ctx.Request.RequestURI, ctx.Writer.Status(), size, time.Since(start), body)
	}
}

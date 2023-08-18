package handler

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type requestLogger interface {
	Info(template string, args ...interface{})
}

func Logger(l requestLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()
		size := ctx.Writer.Size()
		if size < 0 {
			size = 0
		}
		l.Info("%s %s status %d size %d duration %s", ctx.Request.Method, ctx.Request.RequestURI, ctx.Writer.Status(), size, time.Since(start))
	}
}

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
		l.Info("%s %s status %d size %d duration %s %s", ctx.Request.Method, ctx.Request.RequestURI, ctx.Writer.Status(), size, time.Since(start), body)
	}
}

func RequireJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.ContentType() != "application/json" {
			ctx.AbortWithStatus(http.StatusBadRequest)
		}
		ctx.Next()
	}
}

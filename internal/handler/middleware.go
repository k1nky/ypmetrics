package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

type logger interface {
	Info(args ...interface{})
}

func Logger(l logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		l.Info(c.Request.Method, c.Request.RequestURI, "status", c.Writer.Status(), "size", c.Writer.Size(), "duration", time.Since(start))
	}
}

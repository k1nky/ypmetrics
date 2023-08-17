package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

type requestLogger interface {
	Info(template string, args ...interface{})
}

func Logger(l requestLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}
		l.Info("%s %s status %d size %d duration %s", c.Request.Method, c.Request.RequestURI, c.Writer.Status(), size, time.Since(start))
	}
}

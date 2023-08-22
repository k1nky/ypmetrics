package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireContentType это middleware, который определяет требование для значения заголовка ContentType
func RequireContentType(contentType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.ContentType() != contentType {
			ctx.AbortWithStatus(http.StatusBadRequest)
		}
		ctx.Next()
	}
}

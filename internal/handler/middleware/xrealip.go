package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func XRealIP(subnet net.IPNet) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip, err := resolveIP(ctx.Request)
		if err != nil || !subnet.Contains(ip) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}

func resolveIP(r *http.Request) (net.IP, error) {
	ipStr := r.Header.Get("X-Real-IP")
	if len(ipStr) == 0 {
		return nil, nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}
	return ip, nil
}

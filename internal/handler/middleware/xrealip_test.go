package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestXRealIPTrustedSubnet(t *testing.T) {
	tests := []struct {
		name          string
		wantStatus    int
		xrealip       string
		trustedSubnet string
	}{
		{
			name:          "Allowed source IP",
			wantStatus:    http.StatusOK,
			xrealip:       "192.168.1.100",
			trustedSubnet: "192.168.1.0/24",
		},
		{
			name:          "Forbidden source IP",
			wantStatus:    http.StatusForbidden,
			xrealip:       "192.168.100.100",
			trustedSubnet: "192.168.1.0/25",
		},
		{
			name:          "Empty source IP",
			wantStatus:    http.StatusForbidden,
			xrealip:       "",
			trustedSubnet: "192.168.1.0/25",
		},
	}
	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		_, subnet, err := net.ParseCIDR(tt.trustedSubnet)
		if err != nil {
			t.Error(err)
			return
		}
		r.Any("/", XRealIP(*subnet), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
		c.Request.Header.Set("x-real-ip", tt.xrealip)
		r.ServeHTTP(w, c.Request)
		result := w.Result()
		defer result.Body.Close()
		assert.Equal(t, tt.wantStatus, result.StatusCode, tt.name)
	}
}

package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSealVerify(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		data       string
		hashValue  string
	}{
		{
			name:       "With valid hash",
			wantStatus: http.StatusOK,
			data:       "hello",
			hashValue:  "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b",
		},
		{
			name:       "With invalid hash",
			wantStatus: http.StatusBadRequest,
			data:       "bye",
			hashValue:  "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b",
		},
		{
			name:       "Without header",
			wantStatus: http.StatusOK,
			data:       "bye",
			hashValue:  "",
		},
	}
	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Any("/", NewSeal("secret").Use(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		buf := &bytes.Buffer{}
		buf.WriteString(tt.data)
		c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
		if len(tt.hashValue) > 0 {
			c.Request.Header.Set("HashSHA256", tt.hashValue)
		}
		r.ServeHTTP(w, c.Request)
		assert.Equal(t, tt.wantStatus, w.Result().StatusCode, tt.name)
	}
}

func TestSealResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Any("/", NewSeal("secret").Use(), func(c *gin.Context) {
		c.Writer.WriteString("hello")
		c.Status(http.StatusOK)
	})
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	r.ServeHTTP(w, c.Request)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b", w.Header().Get("HashSHA256"))
}

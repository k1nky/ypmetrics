package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireContentTypeMiddleware(t *testing.T) {
	type want struct {
		statusCode int
	}
	type contentTypeHeader struct {
		name  string
		value string
	}
	tests := []struct {
		name        string
		contentType contentTypeHeader
		want        want
	}{
		{
			name:        "Correct uppercase",
			contentType: contentTypeHeader{name: "Content-Type", value: "application/json"},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Correct lowercase",
			contentType: contentTypeHeader{name: "content-type", value: "application/json"},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Incorrect header value",
			contentType: contentTypeHeader{name: "content-type", value: "application/xml"},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Empty header value",
			contentType: contentTypeHeader{name: "content-type", value: ""},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "No header",
			contentType: contentTypeHeader{name: "", value: ""},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/", RequireContentType("application/json"))
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			c.Request.Header.Set(tt.contentType.name, tt.contentType.value)
			r.ServeHTTP(w, c.Request)

			result := w.Result()
			defer result.Body.Close()
			if !assert.Equal(t, tt.want.statusCode, result.StatusCode) {
				return
			}
		})
	}
}

package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestShouldCompress(t *testing.T) {
	type args struct {
		acceptEncoding      string
		contentType         string
		allowedContentTypes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "With allowed content type",
			args: args{
				acceptEncoding:      "gzip",
				contentType:         "application/json",
				allowedContentTypes: []string{"text/plain", "application/json"},
			},
			want: true,
		},
		{
			name: "Allow any content-type",
			args: args{
				acceptEncoding:      "gzip",
				contentType:         "application/json",
				allowedContentTypes: []string{},
			},
			want: true,
		},
		{
			name: "With not allowed content type",
			args: args{
				acceptEncoding:      "gzip",
				contentType:         "application/xml",
				allowedContentTypes: []string{"text/plain", "application/json"},
			},
			want: false,
		},
		{
			name: "With cpmpression level",
			args: args{
				acceptEncoding:      "gzip;q=1",
				contentType:         "text/plain",
				allowedContentTypes: []string{"text/plain", "application/json"},
			},
			want: true,
		},
		{
			name: "With multiple algo",
			args: args{
				acceptEncoding:      "gzip;q=1 brotli",
				contentType:         "text/plain",
				allowedContentTypes: []string{"text/plain", "application/json"},
			},
			want: true,
		},
		{
			name: "Don't compress",
			args: args{
				acceptEncoding:      "",
				contentType:         "text/plain",
				allowedContentTypes: []string{},
			},
			want: false,
		},
		{
			name: "Unsupported algo",
			args: args{
				acceptEncoding:      "brotli",
				contentType:         "text/plain",
				allowedContentTypes: []string{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if len(tt.args.acceptEncoding) > 0 {
				req.Header.Add("accept-encoding", tt.args.acceptEncoding)
			}
			resp := httptest.NewRecorder()
			resp.Header().Add("Content-type", tt.args.contentType)
			gh := NewGzip(tt.args.allowedContentTypes)

			if got := gh.shouldCompress(req, resp); got != tt.want {
				t.Errorf("shouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldUncompress(t *testing.T) {
	type args struct {
		contentEncoding string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Gzip",
			args: args{contentEncoding: "gzip"},
			want: true,
		},
		{
			name: "Gzip with level",
			args: args{contentEncoding: "gzip;q=1"},
			want: true,
		},
		{
			name: "Don't compress",
			args: args{contentEncoding: ""},
			want: false,
		},
		{
			name: "Unsupported algo",
			args: args{contentEncoding: "brotli"},
			want: false,
		},
	}
	gh := NewGzip([]string{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if len(tt.args.contentEncoding) > 0 {
				req.Header.Add("content-encoding", tt.args.contentEncoding)
			}

			if got := gh.shouldUncompress(req); got != tt.want {
				t.Errorf("shouldUncompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressResponse(t *testing.T) {
	tests := []struct {
		name           string
		shouldCompress bool
		body           string
	}{
		{
			name:           "Compress",
			shouldCompress: true,
			body:           strings.Repeat("abcdef", 100),
		},
		{
			name:           "Don't Compress",
			shouldCompress: false,
			body:           strings.Repeat("abcdef", 100),
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/", NewGzip([]string{}).Use(), func(c *gin.Context) {
				c.String(http.StatusOK, tt.body)
			})
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.shouldCompress {
				c.Request.Header.Set("Accept-Encoding", "gzip")
			}
			r.ServeHTTP(w, c.Request)

			result := w.Result()
			defer result.Body.Close()
			if !assert.Equal(t, http.StatusOK, result.StatusCode) {
				return
			}
			buf := bytes.Buffer{}
			body := result.Body
			if tt.shouldCompress {
				gz, err := gzip.NewReader(body)
				if err != nil {
					t.Error("unexpected error: ", err)
					return
				}
				body = gz
				assert.Equal(t, "gzip", result.Header.Get("content-encoding"))
			} else {
				assert.Equal(t, "", result.Header.Get("content-encoding"))
			}
			buf.ReadFrom(body)
			assert.Equal(t, tt.body, buf.String())
		})
	}
}

func TestUncompressRequest(t *testing.T) {
	tests := []struct {
		name             string
		shouldUncompress bool
		body             string
	}{
		{
			name:             "Compress",
			shouldUncompress: true,
			body:             strings.Repeat("abcdef", 100),
		},
		{
			name:             "Don't Compress",
			shouldUncompress: false,
			body:             strings.Repeat("abcdef", 100),
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/", NewGzip([]string{}).Use(), func(c *gin.Context) {
				buf := bytes.Buffer{}
				buf.ReadFrom(c.Request.Body)
				assert.Equal(t, tt.body, buf.String())
				c.Status(http.StatusOK)
			})
			buf := &bytes.Buffer{}
			if tt.shouldUncompress {
				gz, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
				if err != nil {
					t.Error("unexpected error: ", err)
					return
				}
				if _, err := gz.Write([]byte(tt.body)); err != nil {
					t.Error("unexpected error: ", err)
					return
				}
				gz.Close()
			} else {
				buf.WriteString(tt.body)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/", buf)
			if tt.shouldUncompress {
				c.Request.Header.Set("Content-Encoding", "gzip")
			}
			r.ServeHTTP(w, c.Request)
		})
	}
}

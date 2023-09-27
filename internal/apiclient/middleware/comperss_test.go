package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSealShouldCompress(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should be compressed",
			args: args{r: func() *http.Request {
				b := &bytes.Buffer{}
				b.WriteString("some data")
				r, _ := http.NewRequest(http.MethodPost, "/", b)
				return r
			}()},
			want: true,
		},
		{
			name: "POST without data",
			args: args{r: func() *http.Request {
				r, _ := http.NewRequest(http.MethodPost, "/", nil)
				return r
			}()},
			want: false,
		},
		{
			name: "PATCH without data",
			args: args{r: func() *http.Request {
				r, _ := http.NewRequest(http.MethodPatch, "/", nil)
				return r
			}()},
			want: false,
		},
		{
			name: "PATCH with data",
			args: args{r: func() *http.Request {
				b := &bytes.Buffer{}
				b.WriteString("some data")
				r, _ := http.NewRequest(http.MethodPatch, "/", b)
				return r
			}()},
			want: false,
		},
	}
	z := NewGzip()
	for _, tt := range tests {
		got := z.shouldCompress(tt.args.r)
		assert.Equal(t, tt.want, got)
	}
}

func TestGzipCompress(t *testing.T) {
	tests := []struct {
		data string
	}{
		{data: "hello world"},
		{data: "Авторы: Tony Iommi, William Ward, Ozzy Osbourne, Black Sabbath, Michael Butler"},
	}
	// middleware использует пул ресурсов, поэтому стоит запускать несколько тестов на одном
	// экземпляре middleware
	gz := NewGzip()
	for _, tt := range tests {
		httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body := bytes.NewBuffer(nil)
			body.ReadFrom(req.Body)
			// есть заголовок content-encoding
			assert.Equal(t, "gzip", req.Header.Get("content-encoding"))
			// размер тела совпадает с content-length
			assert.Equal(t, int64(body.Len()), req.ContentLength)
			// тело сжато
			buf := bytes.NewBuffer(nil)
			r, err := gzip.NewReader(body)
			if err != nil {
				t.Error(err)
				return
			}
			buf.ReadFrom(r)
			assert.Equal(t, tt.data, buf.String())
		}))

		cli := resty.NewWithClient(httpserver.Client())
		cli.SetPreRequestHook(gz.Use())
		body := bytes.NewBuffer(nil)
		body.WriteString(tt.data)
		r := cli.R().SetBody(body)
		r.Post(httpserver.URL)
		httpserver.Close()
	}
}

func TestGzipNotCompress(t *testing.T) {
	tests := []struct {
		data string
	}{
		{data: "hello world"},
	}
	gz := NewGzip()
	for _, tt := range tests {
		httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body := bytes.NewBuffer(nil)
			body.ReadFrom(req.Body)
			// заголовка content-encoding не должно быть
			assert.Equal(t, "", req.Header.Get("content-encoding"))
			// размер тела совпадает с content-length
			assert.Equal(t, int64(len(tt.data)), req.ContentLength)
			// тело не сжато
			assert.Equal(t, tt.data, body.String())
		}))

		cli := resty.NewWithClient(httpserver.Client())
		cli.SetPreRequestHook(gz.Use())
		body := bytes.NewBuffer(nil)
		body.WriteString(tt.data)
		r := cli.R().SetBody(body)
		r.Patch(httpserver.URL)
		httpserver.Close()
	}
}

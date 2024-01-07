package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSealShouldSign(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should be signed",
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
	s := NewSeal("")
	for _, tt := range tests {
		got := s.shouldSign(tt.args.r)
		assert.Equal(t, tt.want, got)
	}
}

func TestSealSignRequest(t *testing.T) {
	tests := []struct {
		data string
		hash string
	}{
		{data: "hello", hash: "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b"},
		{data: "hello world", hash: "734cc62f32841568f45715aeb9f4d7891324e6d948e4c6c60c0621cdac48623a"},
	}

	seal := NewSeal("secret")
	for _, tt := range tests {
		httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			h := req.Header.Get("HashSHA256")
			// должен быть заголовок HashSHA256 с длиной значения в 64 символа
			assert.Len(t, h, 64)
			// хеш из заголовка должен совпадать с ожидаемым значением
			assert.Equal(t, tt.hash, h)
			rw.WriteHeader(http.StatusOK)
		}))

		cli := resty.NewWithClient(httpserver.Client())
		cli.SetPreRequestHook(seal.Use())
		r := cli.R().SetBody(tt.data)
		r.Post(httpserver.URL)
		httpserver.Close()
	}
}

func TestSealNoSignRequest(t *testing.T) {

	seal := NewSeal("secret")
	httpserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "", req.Header.Get("HashSHA256"))
		rw.WriteHeader(http.StatusOK)
	}))

	cli := resty.NewWithClient(httpserver.Client())
	cli.SetPreRequestHook(seal.Use())
	r := cli.R().SetBody(nil)
	r.Post(httpserver.URL)
	httpserver.Close()
}

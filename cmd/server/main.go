package main

import (
	"log"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/server"
	"github.com/k1nky/ypmetrics/internal/storage"
)

var (
	config Config
)

func main() {

	if err := config.Parse(nil); err != nil {
		panic(err)
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	srv := server.New(server.WithStorage(storage.NewMemStorage()))
	log.Println("server starting on ", config.Address)
	handler := handler.New(srv)
	return http.ListenAndServe(config.Address.String(), handler)
}

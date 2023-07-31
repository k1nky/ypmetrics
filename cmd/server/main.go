package main

import (
	"log"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/handler"
	"github.com/k1nky/ypmetrics/internal/server"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	srv := server.New(server.WithStorage(storage.NewMemStorage()))
	log.Println("server starting")
	handler := handler.New(srv)
	return http.ListenAndServe("localhost:8080", handler)
}

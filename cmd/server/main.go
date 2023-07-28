package main

import (
	"log"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/server"
	"github.com/k1nky/ypmetrics/internal/storage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage := storage.NewMemStorage()
	srv := server.New(storage)
	log.Println("server starting")
	return http.ListenAndServe("localhost:8080", srv.Serve())
}

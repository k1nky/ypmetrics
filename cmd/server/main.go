package main

import (
	"log"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/handlers"
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
	mux := http.NewServeMux()

	mux.HandleFunc("/", http.NotFound)
	mux.Handle("/update/", handlers.UpdateHandler(srv))

	log.Println("server starting")
	return http.ListenAndServe("localhost:8080", mux)
}

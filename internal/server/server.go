package server

import (
	"log"
	"net/http"
	"time"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/handlers"
)

type Server struct {
	Logger     *log.Logger
	HTTPServer *http.Server
}

func New(logger *log.Logger) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handlers.Index)
	mux.HandleFunc("/upload", handlers.Upload)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return &Server{
		Logger:     logger,
		HTTPServer: httpServer,
	}
}

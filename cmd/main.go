package main

import (
	"log"
	"os"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "morse: ", log.LstdFlags)
	server := server.New(logger)
	err := server.HTTPServer.ListenAndServe()
	if err != nil {
		logger.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}

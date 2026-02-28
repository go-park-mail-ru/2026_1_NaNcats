package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/httphandler"
)

func main() {
	port := os.Getenv("PORT") // выделенный под сервер порт из окружения
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/register", httphandler.RegisterHandler)

	// добавить роут api/login

	log.Printf("Server is starting on port %s...", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

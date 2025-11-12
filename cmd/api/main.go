package main

import (
	"log"
	"net/http"
	"time"

	httpserver "gestionsvc/internal/transport/http"
	"gestionsvc/internal/repository"
	"gestionsvc/internal/services"
)

func main() {
	// Repo JSON 
	repo, err := repository.NewJSONStore("data")
	if err != nil {
		log.Fatal(err)
	}

	// Service métier
	booking := services.NewBookingService(repo)

	// Serveur HTTP (API)
	srv := httpserver.NewServer(booking)

	// Routeur principal
	mux := http.NewServeMux()

	// Front statique (web/)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	// API
	mux.Handle("/services", srv.Mux)
	mux.Handle("/services/", srv.Mux)     
	mux.Handle("/auth/", srv.Mux)
	mux.Handle("/admin/", srv.Mux)
	mux.Handle("/reservations", srv.Mux) 
	mux.Handle("/reservations/", srv.Mux) 

	// Serveur avec timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server listening on :8080")
	log.Fatal(server.ListenAndServe())
}
package main

import (
	"log"
	"net/http"

	httpserver "gestionsvc/internal/transport/http"
	"gestionsvc/internal/repository"
	"gestionsvc/internal/services"
)

func main() {
	// Repo JSON (dossier data/)
	repo, err := repository.NewJSONStore("data")
	if err != nil {
		log.Fatal(err)
	}

	// Service métier
	booking := services.NewBookingService(repo)

	// Server HTTP (API)
	srv := httpserver.NewServer(booking)

	// Front statique (web/)
	http.Handle("/", http.FileServer(http.Dir("web")))
	// API
	http.Handle("/services", srv.Mux)
	http.Handle("/services/", srv.Mux)
	http.Handle("/auth/", srv.Mux)
	http.Handle("/admin/", srv.Mux)
	http.Handle("/reservations", srv.Mux)
	http.Handle("/reservations/", srv.Mux)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	// Build-in chi middlewares for development convenience
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]string{"status": "ok"}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Printf("error enconding health response %v", err)
		}
	})

	log.Printf("Starting server on port %s", port)

	addr := fmt.Sprintf(":%s", port)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Hapkiduki/blog-api/internal/middleware"
	"github.com/Hapkiduki/blog-api/pkg/logger"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	// Initialize the logger — this is the ONLY place where the logger is created.
	// Everything else receives it via dependency injection.
	log, err := logger.New()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer func() {
		// Sync flushes any buffered log entries.
		// We ignore the error because Sync can fail on stdout/stderr in some environments.
		_ = log.Sync()
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	// Middleware stack (order matters!)
	r.Use(chimw.RequestID)               // 1. Generate request ID
	r.Use(chimw.RealIP)                  // 2. Extract real IP from proxy headers
	r.Use(middleware.RequestLogger(log)) // 3. Log requests with zap (our custom middleware)
	r.Use(chimw.Recoverer)               // 4. Recover from panics

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]string{"status": "ok"}
		encErr := json.NewEncoder(w).Encode(resp)
		if encErr != nil {
			reqLog := logger.FromContext(r.Context())
			reqLog.Error("failed to encode health response", zap.Error(encErr))
		}
	})

	log.Info("starting server", zap.String("port", port))

	addr := fmt.Sprintf(":%s", port)
	if srvErr := http.ListenAndServe(addr, r); srvErr != nil {
		log.Fatal("server failed to start", zap.Error(srvErr))
	}
}

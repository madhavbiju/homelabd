package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/madhavbiju/homelabd/internal/api/handlers"
	"github.com/madhavbiju/homelabd/internal/api/middleware"
	"github.com/madhavbiju/homelabd/internal/database"
)

// NewRouter sets up the Chi router and all routes.
func NewRouter(db *database.Database) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(chimiddleware.RealIP)

	healthHandler := handlers.NewHealthHandler(db)

	// Health and Ready endpoints
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Future routes will go here
		r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
			WriteJSON(w, http.StatusOK, map[string]string{"message": "pong"})
		})
	})

	return r
}

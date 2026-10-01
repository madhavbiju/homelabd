package api

import (
	"net/http"


	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/madhavbiju/homelabd/internal/api/handlers"
	"github.com/madhavbiju/homelabd/internal/api/middleware"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/auth"
	"github.com/madhavbiju/homelabd/internal/database"
	"github.com/madhavbiju/homelabd/internal/system"
)

// NewRouter sets up the Chi router and all routes.
func NewRouter(db *database.Database, authSvc *auth.Service, auditSvc *audit.Service, sysSvc *system.Service) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(chimiddleware.RealIP)

	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(authSvc, auditSvc, db)
	sysHandler := handlers.NewSystemHandler(sysSvc)

	// Health and Ready endpoints
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/tokens", authHandler.CreateToken)
		
		// Protected routes
		r.Group(func(r chi.Router) {
			// Require at least system:read (or admin)
			r.Use(middleware.RequireAuth(authSvc, "system:read"))
			
			r.Route("/system", func(r chi.Router) {
				r.Get("/", sysHandler.GetInfo)
				r.Get("/resources", sysHandler.GetResources)
				r.Get("/storage", sysHandler.GetStorage)
				r.Get("/network", sysHandler.GetNetwork)
				r.Get("/processes", sysHandler.GetProcesses)
			})
		})
	})

	return r
}

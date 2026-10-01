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
	"github.com/madhavbiju/homelabd/internal/docker"
	"github.com/madhavbiju/homelabd/internal/system"
)

// NewRouter sets up the Chi router and all routes.
func NewRouter(db *database.Database, authSvc *auth.Service, auditSvc *audit.Service, sysSvc *system.Service, dockerSvc *docker.Service) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(chimiddleware.RealIP)

	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(authSvc, auditSvc, db)
	sysHandler := handlers.NewSystemHandler(sysSvc)
	var dockerHandler *handlers.DockerHandler
	if dockerSvc != nil {
		dockerHandler = handlers.NewDockerHandler(dockerSvc, auditSvc)
	}

	// Health and Ready endpoints
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/tokens", authHandler.CreateToken)
		
		// Protected routes
		r.Group(func(r chi.Router) {
			
			r.Route("/system", func(r chi.Router) {
				// Require at least system:read (or admin)
				r.Use(middleware.RequireAuth(authSvc, "system:read"))
				r.Get("/", sysHandler.GetInfo)
				r.Get("/resources", sysHandler.GetResources)
				r.Get("/storage", sysHandler.GetStorage)
				r.Get("/network", sysHandler.GetNetwork)
				r.Get("/processes", sysHandler.GetProcesses)
			})

			if dockerHandler != nil {
				r.Route("/docker", func(r chi.Router) {
					// Read routes
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireAuth(authSvc, "docker:read"))
						r.Get("/", dockerHandler.GetInfo)
						r.Get("/containers", dockerHandler.ListContainers)
						r.Get("/containers/{id}", dockerHandler.GetContainer)
						r.Get("/containers/{id}/logs", dockerHandler.GetLogs)
						r.Get("/containers/{id}/stats", dockerHandler.GetStats)
						r.Get("/images", dockerHandler.ListImages)
						r.Get("/images/{id}", dockerHandler.GetImage)
					})
					
					// Operate routes
					r.Group(func(r chi.Router) {
						r.Use(middleware.RequireAuth(authSvc, "docker:operate"))
						r.Post("/containers/{id}/{operation}", dockerHandler.OperateContainer)
						r.Post("/images/pull", dockerHandler.PullImage)
					})
				})
			}
		})
	})

	return r
}

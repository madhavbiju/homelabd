package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/madhavbiju/homelabd/internal/api"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/auth"
	"github.com/madhavbiju/homelabd/internal/config"
	"github.com/madhavbiju/homelabd/internal/database"
	"github.com/madhavbiju/homelabd/internal/docker"
	"github.com/madhavbiju/homelabd/internal/logger"
	"github.com/madhavbiju/homelabd/internal/system"
)

func main() {
	configPath := flag.String("config", "", "path to configuration file")
	flag.Parse()

	// Load Configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Logger
	logger.Init(cfg.Server.LogLevel)
	slog.Info("Starting homelabd", "config_path", *configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Database
	db, err := database.Connect(ctx, cfg.Database.Path)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		slog.Error("Failed to apply database migrations", "error", err)
		os.Exit(1)
	}

	// Initialize Services
	auditSvc := audit.NewService(db)
	authSvc := auth.NewService(db)
	sysSvc := system.NewService()
	
	var dockerSvc *docker.Service
	if cfg.Docker.Enabled {
		ds, err := docker.NewService(cfg.Docker.Socket)
		if err != nil {
			slog.Warn("Docker is enabled but failed to initialize. Continuing without Docker.", "error", err)
		} else {
			dockerSvc = ds
			slog.Info("Docker integration initialized")
		}
	}

	// Initialize Router
	router := api.NewRouter(db, authSvc, auditSvc, sysSvc, dockerSvc)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start Server
	go func() {
		slog.Info("Server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited properly")
}

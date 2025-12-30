package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eslutz/Messagarr/internal/api"
	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/logger"

	"github.com/eslutz/Messagarr/docs"
)

// version is set at build time via ldflags
var version = "dev"

// @title Messagarr API
// @version {{VERSION}}
// @description Messagarr is a lightweight notification aggregation service that routes messages
// @description to multiple channels (Email, Discord, Slack, Teams) based on priority.

// @contact.name Messagarr Support
// @contact.url https://github.com/eslutz/Messagarr

// @license.name MIT
// @license.url https://github.com/eslutz/Messagarr/blob/main/LICENSE

// @host localhost:4545
// @BasePath /

// @tag.name notifications
// @tag.description Notification operations
// @tag.name health
// @tag.description Health and monitoring endpoints

func main() {
	// Set swagger version dynamically
	docs.SwaggerInfo.Version = version

	// Initialize logger
	logger.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set log level from config
	logger.SetLevel(cfg.LogLevel)

	slog.Info("Starting Messagarr", "version", version)

	// Create HTTP server
	server := api.NewServer(cfg, version)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: server.Router(),
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	if err := httpServer.Shutdown(ctx); err != nil {
		cancel()
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	cancel()

	slog.Info("Server exited")
}

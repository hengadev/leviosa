package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/app"
)

// serve runs the HTTP server and background workers until SIGINT/SIGTERM. It is the
// default mode.
func serve(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("serve takes no arguments, got %q", args)
	}

	// Setup logger
	logger := setupLogger()

	// Load configuration
	cfg, err := app.LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.InfoContext(ctx, "Starting Leviosa modular monolith",
		"environment", cfg.Environment,
		"port", cfg.ServerPort,
	)

	// Create dependency injection container
	container, err := app.NewContainer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	// Create HTTP server
	server := app.NewServer(container, logger)

	// Start background workers
	reminderCtx, reminderCancel := context.WithCancel(context.Background())
	go container.ReminderScheduler.Start(reminderCtx)

	// Setup graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start server in goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.Start(ctx); err != nil {
			serverErrCh <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.InfoContext(ctx, "Shutdown signal received")

		// Stop background workers
		reminderCancel()
		container.ReminderScheduler.Stop()

		// Give server 30 seconds to shutdown gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		logger.InfoContext(ctx, "Server stopped gracefully")
		return nil

	case err := <-serverErrCh:
		// Stop background workers on server error
		reminderCancel()
		container.ReminderScheduler.Stop()
		return fmt.Errorf("server error: %w", err)
	}
}

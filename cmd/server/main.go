package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aaronmatei/expense_tracker_api_go/internal/config"
	"github.com/aaronmatei/expense_tracker_api_go/internal/database"
	"github.com/aaronmatei/expense_tracker_api_go/internal/server"
)

func main() {
	// Structured logging via slog. JSON in production for log aggregators,
	// text in development for readable terminal output.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel, cfg.IsProduction())
	slog.SetDefault(logger)

	// Root context that gets cancelled on SIGINT/SIGTERM. Anything that
	// respects context (DB queries, HTTP requests) will be told to stop.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to database", "max_conns", pool.Config().MaxConns)

	handler := server.New(pool, server.Config{
		JWTSecret:   cfg.JWTSecret,
		JWTExpiry:   cfg.JWTExpiry(),
		CORSOrigins: cfg.CORSOrigins,
	})

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Run the server in a goroutine so main can wait for the shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "port", cfg.Port, "env", cfg.AppEnv)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for either: signal to shut down, OR the server crashed.
	select {
	case err := <-serverErr:
		slog.Error("server error", "err", err)
		os.Exit(1)
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining...")
	}

	// Give in-flight requests up to 15 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped cleanly...")
}

// newLogger returns a slog.Logger configured for the environment.
// Production gets JSON output (machine-parseable); dev gets text (human-readable).
func newLogger(level string, isProd bool) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: slogLevel}
	if isProd {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

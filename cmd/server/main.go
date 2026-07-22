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

	"github.com/shengyu/gogo/internal/config"
	"github.com/shengyu/gogo/internal/httpapi"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cfg := config.Load()
	logger := newLogger(cfg.LogLevel)

	router := httpapi.NewRouter(httpapi.RouterOptions{
		Environment: cfg.Environment,
		Version:     version,
		Commit:      commit,
		Logger:      logger,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting",
			"address", server.Addr,
			"environment", cfg.Environment,
			"version", version,
			"commit", commit,
		)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case sig := <-shutdownSignals:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		_ = server.Close()
		os.Exit(1)
	}

	logger.Info("HTTP server stopped")
}

func newLogger(level string) *slog.Logger {
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

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
}

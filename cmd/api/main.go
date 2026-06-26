package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/v-saprykin/subscription-aggregator/internal/config"
	sqldb "github.com/v-saprykin/subscription-aggregator/internal/db/sqlc"
	"github.com/v-saprykin/subscription-aggregator/internal/httpserver"
	"github.com/v-saprykin/subscription-aggregator/internal/subscription"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	dbPool, err := pgxpool.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	if err := dbPool.Ping(dbCtx); err != nil {
		logger.Error("failed to connect to database", "error", err)
		dbPool.Close()
		os.Exit(1)
	}
	logger.Info("database connected")

	queries := sqldb.New(dbPool)
	subscriptionRepo := subscription.NewRepository(queries)
	subscriptionService := subscription.NewService(subscriptionRepo)
	subscriptionHandler := subscription.NewHandler(subscriptionService, logger)

	server := httpserver.New(cfg.HTTPAddr, logger, subscriptionHandler.RegisterRoutes)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		logger.Error("server failed", "error", err)
		dbPool.Close()
		os.Exit(1)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("server shutting down")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		dbPool.Close()
		os.Exit(1)
	}
	dbPool.Close()
	logger.Info("server stopped")
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

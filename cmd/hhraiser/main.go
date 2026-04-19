package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rycln/hhraiser/internal/app"
	"github.com/rycln/hhraiser/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	config.InitLogger(cfg.LogLevel)

	a, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("hhraiser started")

	if err := a.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("app stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("hhraiser stopped")
}

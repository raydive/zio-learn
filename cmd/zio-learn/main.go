// Command zio-learn serves the local Scala playground.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"zio-learn/internal/app"
)

func main() {
	if err := run(); err != nil {
		slog.Error("playground failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.Run(ctx, os.Getenv("LISTEN_ADDR"))
}

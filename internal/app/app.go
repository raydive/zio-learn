// Package app composes HTTP, transpilation, and execution and manages shutdown.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"zio-learn/internal/handler"
	"zio-learn/internal/playground"
	"zio-learn/internal/transpiler"
)

// Run serves the editor until ctx is canceled.
func Run(ctx context.Context, address string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}
	if address == "" {
		address = "127.0.0.1:8080"
	}
	p := playground.New(120 * time.Second)
	t := transpiler.New(config.ScalaVersion, config.ScalaVersions, config.Dependencies)
	h, err := handler.New(config, t, p)
	if err != nil {
		return err
	}
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	return serve(ctx, listener, h, p)
}

func serve(ctx context.Context, listener net.Listener, h http.Handler, p *playground.Playground) error {
	defer func() { _ = listener.Close() }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.Shutdown(shutdownCtx); err != nil {
			slog.Warn("shutdown playground", "error", err)
		}
	}()
	httpServer := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.Serve(listener) }()
	slog.Info("HTTP server listening", "address", listener.Addr().String())
	select {
	case err := <-serveErr:
		_ = httpServer.Close()
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(shutdownCtx); err != nil {
		slog.Warn("shutdown playground", "error", err)
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_ = httpServer.Close()
		return fmt.Errorf("shutdown HTTP: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}

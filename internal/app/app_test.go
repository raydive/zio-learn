package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"zio-learn/internal/playground"
)

func TestServeStopsExecution(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "shutdown"
		if fail {
			name = "listen failure"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "scala-cli"), []byte("#!/bin/sh\nsleep 30\n"), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := playground.New(time.Minute)
			run, err := p.Start(ctx, "playground.sc", "")
			if err != nil {
				t.Fatal(err)
			}
			<-run.Events
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			if fail {
				if err := listener.Close(); err != nil {
					t.Fatal(err)
				}
			}
			h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			done := make(chan error, 1)
			go func() { done <- serve(ctx, listener, h, p) }()
			if !fail {
				response, err := http.Get("http://" + listener.Addr().String())
				if err != nil {
					t.Fatal(err)
				}
				if err := response.Body.Close(); err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != http.StatusNoContent {
					t.Fatal(response.StatusCode)
				}
				cancel()
			}
			select {
			case err := <-done:
				if (err != nil) != fail {
					t.Fatalf("serve error=%v", err)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("serve did not stop")
			}
			if _, err := p.Start(context.Background(), "playground.sc", ""); !errors.Is(err, playground.ErrClosed) {
				t.Fatalf("expected closed playground, got %v", err)
			}
		})
	}
}

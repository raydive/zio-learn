package playground

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeCLI(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	cli := filepath.Join(dir, "scala-cli")
	if err := os.WriteFile(cli, []byte("#!/bin/sh\n"+source), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return cli
}

func TestExecutionAndCleanup(t *testing.T) {
	cli := fakeCLI(t, "printf '%s' \"$PWD\" > \"$0.directory\"\ncat playground.sc\nprintf '\\n日本語\\n'\nexit 7\n")
	p := New(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	run, err := p.Start(ctx, "playground.sc", "example source")
	if err != nil {
		t.Fatal(err)
	}
	var events []Event
	for event := range run.Events {
		events = append(events, event)
	}
	if len(events) < 3 || events[0].Type != "started" || events[len(events)-1].Type != "finished" {
		t.Fatalf("events=%+v", events)
	}
	var text string
	for _, event := range events {
		text += event.Text
	}
	result := events[len(events)-1]
	if result.Status != "error" || result.ExitCode != 7 || !strings.Contains(text, "example source\n日本語") {
		t.Fatalf("result=%+v output=%q", result, text)
	}
	dir, err := os.ReadFile(cli + ".directory")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(string(dir)); !os.IsNotExist(err) {
		t.Fatalf("directory remains: %s (%v)", dir, err)
	}
}

func TestStartFailureReleasesSlot(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	p := New(time.Second)
	for range 2 {
		if _, err := p.Start(context.Background(), "playground.sc", ""); err == nil || errors.Is(err, ErrBusy) {
			t.Fatalf("err=%v", err)
		}
	}
}

func TestCancellationUnblocksFullChannel(t *testing.T) {
	fakeCLI(t, "while :; do printf 'output\\n'; done\n")
	p := New(time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	run, err := p.Start(ctx, "playground.sc", "")
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	// Wait until the bounded stream is full, then abandon it and cancel.
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(3 * time.Second)
	for len(run.Events) < cap(run.Events) {
		select {
		case <-ticker.C:
		case <-deadline:
			cancel()
			t.Fatal("output stream did not fill")
		}
	}
	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	if err := p.Shutdown(shutdownCtx); err != nil {
		t.Fatal(err)
	}
	for range run.Events {
	}
	if _, err := p.Start(context.Background(), "playground.sc", ""); !errors.Is(err, ErrClosed) {
		t.Fatalf("shutdown accepted a new run: %v", err)
	}
}

func TestShutdownStopsActiveRun(t *testing.T) {
	fakeCLI(t, "sleep 30\n")
	p := New(time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	run, err := p.Start(ctx, "playground.sc", "")
	if err != nil {
		t.Fatal(err)
	}
	<-run.Events
	if _, err := p.Start(ctx, "playground.sc", ""); !errors.Is(err, ErrBusy) {
		t.Fatalf("busy=%v", err)
	}
	if p.Stop("wrong-id") {
		t.Fatal("stopped a different run")
	}
	if err := p.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	var result Event
	for event := range run.Events {
		result = event
	}
	if result.Status != "stopped" {
		t.Fatalf("result=%+v", result)
	}
}

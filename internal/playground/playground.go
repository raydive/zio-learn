// Package playground executes Scala CLI scripts independently of HTTP.
package playground

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var (
	// ErrBusy means another execution is active.
	ErrBusy = errors.New("別のコードを実行中です。終了後に再実行してください。")
	// ErrClosed means the playground is shutting down.
	ErrClosed      = errors.New("playground is shutting down")
	errStopped     = errors.New("stopped")
	errTimeout     = errors.New("timeout")
	errOutputLimit = errors.New("output_limit")
)

// Event reports process lifecycle or output. It has no HTTP or JSON dependency.
type Event struct {
	Type     string
	ID       string
	Filename string
	Text     string
	Status   string
	ExitCode int
	Duration time.Duration
}

// Execution exposes a read-only event stream and cancellation for its caller.
type Execution struct {
	ID     string
	Events <-chan Event
	cancel context.CancelCauseFunc
	done   chan struct{}
}

// Stop cancels this execution without affecting later executions.
func (e *Execution) Stop() { e.cancel(errStopped) }

// Playground owns process execution and enforces a single active run.
type Playground struct {
	timeout time.Duration
	mu      sync.Mutex
	active  *Execution
	closed  bool
}

// New configures the compile-and-run timeout.
func New(timeout time.Duration) *Playground { return &Playground{timeout: timeout} }

// Start prepares and starts Scala CLI; failures are returned before any events.
// The caller must consume Events or cancel ctx when it stops consuming.
func (p *Playground) Start(ctx context.Context, filename, source string) (*Execution, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrClosed
	}
	if p.active != nil {
		p.mu.Unlock()
		return nil, ErrBusy
	}
	runCtx, cancel := context.WithCancelCause(ctx)
	events := make(chan Event, 16)
	run := &Execution{ID: fmt.Sprintf("%d", time.Now().UnixNano()), Events: events, cancel: cancel, done: make(chan struct{})}
	p.active = run
	p.mu.Unlock()
	started := false
	defer func() {
		if !started {
			cancel(nil)
			p.release(run)
			close(run.done)
		}
	}()
	dir, err := os.MkdirTemp("", "zio-learn-")
	if err != nil {
		return nil, fmt.Errorf("create execution directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			removeDirectory(dir)
		}
	}()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(source), 0600); err != nil {
		return nil, fmt.Errorf("write script: %w", err)
	}
	ready := make(chan struct{})
	output := &outputWriter{ctx: runCtx, cancel: cancel, events: events, ready: ready}
	cmd := scalaCommand(runCtx, filename, dir)
	cmd.Stdout, cmd.Stderr = output, output
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start scala-cli: %w", err)
	}
	start := time.Now()
	timer := time.AfterFunc(p.timeout, func() { cancel(errTimeout) })
	events <- Event{Type: "started", ID: run.ID, Filename: filename}
	close(ready)
	started = true
	cleanup = false
	go func() {
		defer close(events)
		err := cmd.Wait()
		timer.Stop()
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		output.finish()
		status := "success"
		if err != nil {
			status = "error"
		}
		if cause := context.Cause(runCtx); cause != nil {
			status = cause.Error()
		}
		result := Event{Type: "finished", Status: status, ExitCode: cmd.ProcessState.ExitCode(), Duration: time.Since(start)}
		removeDirectory(dir)
		p.release(run)
		close(run.done)
		select {
		case events <- result:
		case <-ctx.Done():
		}
		cancel(nil)
	}()
	return run, nil
}

func (p *Playground) release(run *Execution) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active == run {
		p.active = nil
	}
}

// Stop cancels the active execution matching id.
func (p *Playground) Stop(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active == nil || p.active.ID != id {
		return false
	}
	p.active.Stop()
	return true
}

// Shutdown rejects new runs and waits for the active process to be reaped.
func (p *Playground) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	p.closed = true
	run := p.active
	if run != nil {
		run.Stop()
	}
	p.mu.Unlock()
	if run == nil {
		return nil
	}
	select {
	case <-run.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

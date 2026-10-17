// Package handler exposes transpilation and execution through HTTP.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"zio-learn/internal/assets"
	"zio-learn/internal/playground"
	"zio-learn/internal/transpiler"
)

// Config is the public editor configuration served to the browser.
type Config struct {
	ScalaVersion  string            `json:"scalaVersion"`
	ScalaVersions []string          `json:"scalaVersions"`
	Dependencies  []string          `json:"dependencies"`
	Examples      map[string]string `json:"examples"`
}

// Handler connects HTTP requests to the transpiler and execution service.
type Handler struct {
	config     Config
	transpiler *transpiler.Transpiler
	playground *playground.Playground
}

// New constructs the HTTP adapter using explicitly supplied services.
func New(config Config, t *transpiler.Transpiler, p *playground.Playground) (http.Handler, error) {
	h := &Handler{config: config, transpiler: t, playground: p}
	static, err := fs.Sub(assets.Files(), "web")
	if err != nil {
		return nil, fmt.Errorf("open web assets: %w", err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(static)))
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, _ *http.Request) { respond(w, http.StatusOK, h.config) })
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/script", h.script)
	mux.HandleFunc("POST /api/run", h.run)
	mux.HandleFunc("POST /api/stop", h.stop)
	return localAccess(mux), nil
}

type runRequest struct {
	ScalaVersion string `json:"scalaVersion"`
	Code         string `json:"code"`
}

func (h *Handler) readScript(w http.ResponseWriter, r *http.Request) (transpiler.Script, bool) {
	var req runRequest
	if !readJSON(w, r, &req) {
		return transpiler.Script{}, false
	}
	input, err := h.transpiler.Validate(transpiler.Input{Code: req.Code, ScalaVersion: req.ScalaVersion})
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return transpiler.Script{}, false
	}
	return h.transpiler.Transpile(input), true
}

func (h *Handler) script(w http.ResponseWriter, r *http.Request) {
	script, ok := h.readScript(w, r)
	if !ok {
		return
	}
	respond(w, http.StatusOK, script)
}

func (h *Handler) run(w http.ResponseWriter, r *http.Request) {
	script, ok := h.readScript(w, r)
	if !ok {
		return
	}
	run, err := h.playground.Start(r.Context(), script.Filename, script.Source)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, playground.ErrBusy) {
			status = http.StatusConflict
		}
		if errors.Is(err, playground.ErrClosed) {
			status = http.StatusServiceUnavailable
		}
		respond(w, status, map[string]string{"error": err.Error()})
		return
	}
	defer run.Stop()
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	controller := http.NewResponseController(w)
	defer func() { _ = controller.SetWriteDeadline(time.Time{}) }()
	for event := range run.Events {
		// Bound slow-client writes so cancellation can release the request handler.
		if err := controller.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil && !errors.Is(err, http.ErrNotSupported) {
			return
		}
		if err := json.NewEncoder(w).Encode(wireEvent(event)); err != nil {
			return
		}
		if err := controller.Flush(); err != nil {
			return
		}
	}
}

// wireEvent keeps JSON/HTTP presentation out of the execution service.
func wireEvent(event playground.Event) map[string]any {
	switch event.Type {
	case "started":
		return map[string]any{"type": event.Type, "id": event.ID, "filename": event.Filename}
	case "output":
		return map[string]any{"type": event.Type, "text": event.Text}
	default:
		return map[string]any{"type": event.Type, "status": event.Status, "exitCode": event.ExitCode, "durationMs": event.Duration.Milliseconds()}
	}
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !h.playground.Stop(req.ID) {
		respond(w, http.StatusNotFound, map[string]string{"error": "実行はすでに終了しています。"})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "stopping"})
}

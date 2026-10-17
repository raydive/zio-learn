package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Debug("write JSON response", "error", err)
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Debug("close request body", "error", err)
		}
	}()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "JSON が不正、または入力が 256 KiB を超えています。"})
		return false
	}
	if decoder.Decode(new(any)) != io.EOF {
		respond(w, http.StatusBadRequest, map[string]string{"error": "JSON は1つだけ送信してください。"})
		return false
	}
	return true
}

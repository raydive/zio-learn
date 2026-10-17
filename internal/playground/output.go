package playground

import (
	"context"
	"sync"
	"unicode/utf8"
)

// outputWriter bounds output and preserves UTF-8 across process writes.
type outputWriter struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelCauseFunc
	events  chan<- Event
	ready   <-chan struct{}
	bytes   int
	pending []byte
}

func (w *outputWriter) emit(text string) error {
	select {
	case w.events <- Event{Type: "output", Text: text}:
		return nil
	case <-w.ctx.Done():
		return context.Cause(w.ctx)
	}
}

func (w *outputWriter) Write(p []byte) (int, error) {
	<-w.ready
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bytes+len(p) > 1024*1024 {
		w.cancel(errOutputLimit)
		return 0, errOutputLimit
	}
	w.bytes += len(p)
	w.pending = append(w.pending, p...)
	data := w.pending
	end := 0
	for end < len(data) && utf8.FullRune(data[end:]) {
		_, size := utf8.DecodeRune(data[end:])
		end += size
	}
	w.pending = append([]byte(nil), data[end:]...)
	if end != 0 {
		if err := w.emit(string(data[:end])); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (w *outputWriter) finish() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.pending) > 0 {
		_ = w.emit(string(w.pending))
		w.pending = nil
	}
}

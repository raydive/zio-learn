package playground

import (
	"context"
	"testing"
)

func TestSplitUTF8(t *testing.T) {
	ready := make(chan struct{})
	close(ready)
	events := make(chan Event, 8)
	output := &outputWriter{ctx: context.Background(), cancel: func(error) {}, events: events, ready: ready}
	for _, b := range []byte("日本語") {
		if _, err := output.Write([]byte{b}); err != nil {
			t.Fatal(err)
		}
	}
	output.finish()
	close(events)
	var text string
	for event := range events {
		text += event.Text
	}
	if text != "日本語" {
		t.Fatalf("output=%q", text)
	}
}

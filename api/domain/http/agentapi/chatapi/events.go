package chatapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// sseWriter wraps an http.ResponseWriter + http.Flusher for sending
// Server-Sent Events.
type sseWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

// newSSEWriter prepares the response for SSE streaming.
// Returns nil if the ResponseWriter does not support flushing.
func newSSEWriter(w http.ResponseWriter) *sseWriter {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	f.Flush()

	return &sseWriter{w: w, f: f}
}

// send writes one SSE event. data is JSON-marshalled if non-nil.
func (s *sseWriter) send(event string, data any) {
	if data != nil {
		b, _ := json.Marshal(data)
		fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, b)
	} else {
		fmt.Fprintf(s.w, "event: %s\ndata: {}\n\n", event)
	}
	s.f.Flush()
}

// sendError is a convenience for the "error" event.
func (s *sseWriter) sendError(msg string) {
	s.send("error", map[string]string{"message": msg})
}

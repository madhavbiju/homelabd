package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/events"
)

type EventsHandler struct {
	broker *events.Broker
}

func NewEventsHandler(broker *events.Broker) *EventsHandler {
	return &EventsHandler{broker: broker}
}

func (h *EventsHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Allow CORS if necessary (usually good for SSE if UI is separate, but we can rely on proxy)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.broker.Subscribe()
	defer h.broker.Unsubscribe(ch)

	// Send an initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case evt := <-ch:
			data, err := json.Marshal(evt.Payload)
			if err != nil {
				continue
			}
			// Write the SSE format
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, string(data))
			
			// Flush the buffer to ensure it sends immediately
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	slog.Debug("Client connected")

	// Create a channel for client disconnection
	clientGone := r.Context().Done()

	rc := http.NewResponseController(w)
	t := stateStore.Subscribe()

	for {
		select {
		case <-clientGone:
			slog.Debug("Client disconnected")
			stateStore.Unsubscribe(t)
			return
		case event := <-t:
			jsonData, _ := json.Marshal(event)
			_, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
			if err != nil {
				return
			}
			err = rc.Flush()
			if err != nil {
				return
			}
		}
	}
}

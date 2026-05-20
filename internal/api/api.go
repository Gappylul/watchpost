package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gappylul/watchpost/internal/hub"
	"github.com/gappylul/watchpost/templates"
)

type Handler struct {
	hub *hub.Hub
}

func New(h *hub.Hub) *Handler {
	return &Handler{hub: h}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("GET /api/services", h.handleGetServices)
	mux.HandleFunc("POST /api/services/{name}/recheck", h.handleRecheck)
	mux.HandleFunc("GET /api/stream", h.handleSSE)
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	services := h.hub.Latest()
	templates.Index(services).Render(r.Context(), w)
}

func (h *Handler) handleGetServices(w http.ResponseWriter, r *http.Request) {
	services := h.hub.Latest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *Handler) handleRecheck(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"message":"recheck triggered for %s"}`, name)
}

func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.hub.Subscribe()
	defer h.hub.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case s := <-ch:
			var buf bytes.Buffer
			templates.Card(s).Render(r.Context(), &buf)
			fmt.Fprintf(w, "event: status_update\ndata: %s\n\n", buf.String())
			w.(http.Flusher).Flush()
		}
	}
}

package hub

import (
	"sync"
	"time"
)

type Status struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[chan Status]struct{}
	latest  map[string]Status
}

func New() *Hub {
	return &Hub{
		clients: make(map[chan Status]struct{}),
		latest:  make(map[string]Status),
	}
}

func (h *Hub) Subscribe() chan Status {
	ch := make(chan Status, 16)

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, s := range h.latest {
		ch <- s
	}

	h.clients[ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(ch chan Status) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ch)
	close(ch)
}

func (h *Hub) Publish(s Status) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.latest[s.Service] = s

	for ch := range h.clients {
		select {
		case ch <- s:
		default:
		}
	}
}

func (h *Hub) Latest() []Status {
	h.mu.RLock()
	defer h.mu.RUnlock()
	statuses := make([]Status, 0, len(h.latest))
	for _, s := range h.latest {
		statuses = append(statuses, s)
	}
	return statuses
}

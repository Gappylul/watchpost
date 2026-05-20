package checker

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gappylul/watchpost/internal/config"
	"github.com/gappylul/watchpost/internal/hub"
)

func NewHTTPChecker(svc config.Service, h *hub.Hub) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		doHTTPCheck(ctx, svc, h)

		ticker := time.NewTicker(svc.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				doHTTPCheck(ctx, svc, h)
			}
		}
	}
}

func NewTCPChecker(svc config.Service, h *hub.Hub) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		doTCPCheck(ctx, svc, h)

		ticker := time.NewTicker(svc.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				doTCPCheck(ctx, svc, h)
			}
		}
	}
}

func doHTTPCheck(ctx context.Context, svc config.Service, h *hub.Hub) {
	start := time.Now()
	resp, err := http.Get(svc.Target)
	latency := time.Since(start).Milliseconds()

	s := hub.Status{
		Service:   svc.Name,
		LatencyMs: latency,
		CheckedAt: time.Now(),
	}

	if err != nil {
		s.Status = "down"
		s.Error = err.Error()
	} else {
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			s.Status = "down"
			s.Error = resp.Status
		} else {
			s.Status = "up"
		}
	}

	h.Publish(s)
}

func doTCPCheck(ctx context.Context, svc config.Service, h *hub.Hub) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", svc.Target, 5*time.Second)
	latency := time.Since(start).Milliseconds()

	s := hub.Status{
		Service:   svc.Name,
		LatencyMs: latency,
		CheckedAt: time.Now(),
	}

	if err != nil {
		s.Status = "down"
		s.Error = err.Error()
	} else {
		conn.Close()
		s.Status = "up"
	}

	h.Publish(s)
}

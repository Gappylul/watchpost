package hub_test

import (
	"testing"

	"github.com/gappylul/watchpost/internal/hub"
)

func TestSubscribeGetsLatest(t *testing.T) {
	h := hub.New()

	h.Publish(hub.Status{
		Service: "postgres",
		Status:  "up",
	})

	ch := h.Subscribe()

	select {
	case s := <-ch:
		if s.Service != "postgres" {
			t.Fatalf("expected postgres, got %s", s.Service)
		}
		if s.Status != "up" {
			t.Fatalf("expected up, got %s", s.Status)
		}
	default:
		t.Fatal("expected a status on subscribe, got nothing")
	}
}

func TestPublishFansOut(t *testing.T) {
	h := hub.New()

	ch1 := h.Subscribe()
	ch2 := h.Subscribe()

	h.Publish(hub.Status{
		Service: "redis",
		Status:  "down",
	})

	for _, ch := range []chan hub.Status{ch1, ch2} {
		select {
		case s := <-ch:
			if s.Service != "redis" {
				t.Fatalf("expected redis, got %s", s.Service)
			}
		default:
			t.Fatal("expected status, got nothing")
		}
	}
}

func TestUnsubscribeCleanup(t *testing.T) {
	h := hub.New()

	ch := h.Subscribe()
	h.Unsubscribe(ch)

	h.Publish(hub.Status{
		Service: "postgres",
		Status:  "up",
	})

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/gappylul/watchpost/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	content := `
services:
  - name: "postgres"
    check: tcp
    target: "postgres:5432"
    interval: "10s"
  - name: "api"
    check: http
    target: "http://api:3000/health"
    interval: "5s"
`
	f, err := os.CreateTemp("", "watchpost-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}
	if cfg.Services[0].Name != "postgres" {
		t.Fatalf("expected postgres, got %s", cfg.Services[0].Name)
	}
	if cfg.Services[0].Interval != 10*time.Second {
		t.Fatalf("expected 10s, got %s", cfg.Services[0].Interval)
	}
	if cfg.Services[1].Interval != 5*time.Second {
		t.Fatalf("expected 5s, got %s", cfg.Services[1].Interval)
	}
}

func TestLoadBadInterval(t *testing.T) {
	content := `
services:
  - name: "postgres"
    check: tcp
    target: "postgres:5432"
    interval: "banana"
`
	f, err := os.CreateTemp("", "watchpost-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	_, err = config.Load(f.Name())
	if err == nil {
		t.Fatal("expected error for bad interval, got nil")
	}
}

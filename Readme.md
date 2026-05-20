# watchpost

A self-hosted service health monitor. Define services in a YAML file, get a live dashboard that updates in real time as things go up and down.

Built with Go, HTMX, SSE, and [goverseer](https://github.com/Gappylul/goverseer) for fault-tolerant worker supervision.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- HTTP and TCP health checks
- Live dashboard via SSE — no polling, no page refresh
- Each checker is a supervised goverseer worker — crashes restart automatically with backoff
- gRPC sidecar on `:9090` for machine-to-machine querying
- Single static binary, ~6MB scratch Docker image

## Usage

Define your services in `watchpost.yml`:

```yaml
services:
  - name: "postgres"
    check: tcp
    target: "postgres:5432"
    interval: "10s"

  - name: "api"
    check: http
    target: "http://api:3000/health"
    interval: "5s"

  - name: "redis"
    check: tcp
    target: "redis:6379"
    interval: "10s"
```

Run with Docker Compose:

```bash
docker compose up
```

Open `http://localhost:8080`.

## API

| Method | Path                           | Description                    |
|--------|--------------------------------|--------------------------------|
| GET    | `/api/services`                | Current status of all services |
| GET    | `/api/stream`                  | SSE stream of status updates   |
| POST   | `/api/services/{name}/recheck` | Manually trigger a recheck     |

## gRPC

The gRPC server runs on `:9090` and exposes two RPCs:

```
GetServices  — returns current status of all services
WatchStatus  — streaming RPC, pushes updates as they happen
```

Query it with grpcurl:

```bash
grpcurl -plaintext localhost:9090 watchpost.Watchpost/GetServices
```

## Development

```bash
# generate templ templates
templ generate

# run locally
go run ./cmd/watchpost/main.go --config watchpost.yml

# run tests
go test ./...

# build docker image
docker compose up --build
```

## CI

GitHub Actions runs on every PR and push to main — lint, test, Docker build. Merges to main push to GHCR automatically.

```
ghcr.io/gappylul/watchpost:latest
```
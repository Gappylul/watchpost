FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o watchpost ./cmd/watchpost

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/watchpost /watchpost

ENTRYPOINT ["/watchpost"]
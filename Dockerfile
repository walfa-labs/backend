# --- Build stage ---
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache deps: copy go.mod/go.sum first, download, then copy source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary (CGO disabled — Sonic works on amd64/arm64 without CGO).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api

# --- Runtime stage ---
FROM scratch

# CA certs for TLS calls to object storage, tzdata for time zones.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /bin/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]

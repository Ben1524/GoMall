# Multi-stage Dockerfile for paymentApi service
FROM golang:1.25.3-alpine3.22 AS builder
RUN apk add --no-cache git build-base
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN set -eux; \
    for i in 1 2 3 4 5; do \
      go mod download && break || (echo "go mod download failed, retrying ($i)"; sleep 5); \
    done

# 下面继续你的编译步骤
COPY . .
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED} GOOS=linux GOARCH=amd64
RUN go build -ldflags "-s -w" -o /app/main ./main.go

# Runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary and default config from builder
COPY --from=builder /app/main .
COPY --from=builder /src/config.example.yaml /app/paymentApi/config.example.yaml

# Default config path for the service (can be overridden by env)
ENV CONFIG_PATH=/app/paymentApi/config.example.yaml

# Create non-root user and switch
RUN addgroup -S app && adduser -S -G app app || true
USER app

# Expose the HTTP port (update if your service listens on a different port)
EXPOSE 8082

ENTRYPOINT ["/app/main"]


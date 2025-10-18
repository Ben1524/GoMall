# Multi-stage Dockerfile for web service
FROM golang:1.25.3-alpine3.22 AS builder
RUN apk add --no-cache git build-base
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy sources and build
COPY . .
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED} GOOS=linux GOARCH=amd64
RUN go build -ldflags "-s -w" -o /app/main ./main.go

# Final runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary and default config into image
COPY --from=builder /app/main .
COPY --from=builder /src/config.example.yaml /app/web/config.example.yaml

# Default location for the service to load config from
ENV CONFIG_PATH=/app/web/config.example.yaml

# Create a non-root user and switch to it
RUN addgroup -S app && adduser -S -G app app || true
USER app

# Expose the HTTP port used by the service
EXPOSE 8080

ENTRYPOINT ["/app/main"]

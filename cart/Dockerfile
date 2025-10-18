FROM golang:1.25.3-alpine3.22 AS builder
RUN apk add --no-cache git
WORKDIR /src
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

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
RUN addgroup -S app && adduser -S -G app app
USER app
ENTRYPOINT ["/app/main"]


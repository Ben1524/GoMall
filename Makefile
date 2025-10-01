.PHONY: help build run stop test clean

help:
	@echo "GoMall - Microservice E-commerce System"
	@echo ""
	@echo "Available commands:"
	@echo "  make build         - Build all services"
	@echo "  make run           - Run all services with docker-compose"
	@echo "  make stop          - Stop all services"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make run-local     - Run services locally (without Docker)"
	@echo ""

build:
	@echo "Building all services..."
	go mod tidy
	go build -o bin/api-gateway ./services/api-gateway/main.go
	go build -o bin/user-service ./services/user-service/main.go
	go build -o bin/product-service ./services/product-service/main.go
	go build -o bin/order-service ./services/order-service/main.go
	@echo "Build complete!"

run:
	@echo "Starting all services with Docker Compose..."
	docker-compose up --build

stop:
	@echo "Stopping all services..."
	docker-compose down

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	docker-compose down -v
	@echo "Clean complete!"

run-local: build
	@echo "Starting services locally..."
	@echo "Starting User Service on port 8081..."
	PORT=8081 ./bin/user-service &
	@echo "Starting Product Service on port 8082..."
	PORT=8082 ./bin/product-service &
	@echo "Starting Order Service on port 8083..."
	PORT=8083 ./bin/order-service &
	@sleep 2
	@echo "Starting API Gateway on port 8080..."
	PORT=8080 ./bin/api-gateway

# Quick Start Guide

## Start the System

### Option 1: Using Make (Recommended for local development)
```bash
make build
make run-local
```

### Option 2: Using Docker Compose
```bash
make run
```

## Example API Requests

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. User Service Examples

#### Register a new user
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "password123"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "password123"
  }'
```

#### Get user by ID
```bash
curl http://localhost:8080/api/user?id=1
```

#### List all users
```bash
curl http://localhost:8080/api/users
```

### 3. Product Service Examples

#### List all products
```bash
curl http://localhost:8080/api/products
```

#### Get product by ID
```bash
curl http://localhost:8080/api/product?id=1
```

#### Create a new product
```bash
curl -X POST http://localhost:8080/api/product/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tablet",
    "description": "10-inch tablet",
    "price": 299.99,
    "stock": 50,
    "category": "Electronics"
  }'
```

#### Update product stock
```bash
curl -X PUT http://localhost:8080/api/product/stock \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "stock": 100
  }'
```

### 4. Order Service Examples

#### Create an order
```bash
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      },
      {
        "product_id": 2,
        "quantity": 1
      }
    ]
  }'
```

#### Get order by ID
```bash
curl http://localhost:8080/api/order?id=1
```

#### List all orders
```bash
curl http://localhost:8080/api/orders
```

#### List orders by user
```bash
curl http://localhost:8080/api/orders?user_id=1
```

#### Update order status
```bash
curl -X PUT http://localhost:8080/api/order/status \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "status": "completed"
  }'
```

## Running Tests

Run the comprehensive test suite:
```bash
./test_api.sh
```

## Stopping Services

### If using make run-local
Press `Ctrl+C` to stop the API Gateway, then:
```bash
pkill -f user-service
pkill -f product-service
pkill -f order-service
```

### If using Docker Compose
```bash
make stop
```

## Clean Up

Remove build artifacts and containers:
```bash
make clean
```

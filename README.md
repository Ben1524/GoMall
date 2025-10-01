# GoMall
微服务电商系统 (Microservice E-commerce System)

A lightweight microservice-based e-commerce system built with Go, demonstrating modern architecture patterns and best practices.

## Architecture

GoMall consists of four main microservices:

```
┌─────────────┐
│ API Gateway │  (Port 8080)
└──────┬──────┘
       │
       ├──────────┬──────────┬──────────┐
       │          │          │          │
┌──────▼───┐ ┌───▼──────┐ ┌─▼────────┐
│   User   │ │ Product  │ │  Order   │
│ Service  │ │ Service  │ │ Service  │
│  :8081   │ │  :8082   │ │  :8083   │
└──────────┘ └──────────┘ └──────────┘
```

### Services

1. **API Gateway (Port 8080)**: 
   - Entry point for all client requests
   - Routes requests to appropriate microservices
   - Handles CORS and request proxying

2. **User Service (Port 8081)**:
   - User registration and authentication
   - User profile management
   - Endpoints: `/register`, `/login`, `/user`, `/users`

3. **Product Service (Port 8082)**:
   - Product catalog management
   - Inventory management
   - Endpoints: `/products`, `/product`, `/product/create`, `/product/stock`

4. **Order Service (Port 8083)**:
   - Order creation and management
   - Order status tracking
   - Endpoints: `/orders`, `/order`, `/order/create`, `/order/status`

## Project Structure

```
GoMall/
├── services/
│   ├── api-gateway/        # API Gateway service
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── user-service/       # User management service
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── product-service/    # Product catalog service
│   │   ├── main.go
│   │   └── Dockerfile
│   └── order-service/      # Order processing service
│       ├── main.go
│       └── Dockerfile
├── common/
│   ├── models/             # Shared data models
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── order.go
│   │   └── response.go
│   ├── utils/              # Shared utilities
│   │   └── http.go
│   └── config/             # Configuration management
│       └── config.go
├── docker-compose.yml      # Docker Compose configuration
├── Makefile               # Build and run commands
├── go.mod                 # Go module dependencies
└── README.md              # This file
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for containerized deployment)

## Getting Started

### Option 1: Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/Ben1524/GoMall.git
cd GoMall
```

2. Start all services:
```bash
make run
```

This will build and start all services in containers.

### Option 2: Running Locally

1. Clone the repository:
```bash
git clone https://github.com/Ben1524/GoMall.git
cd GoMall
```

2. Build all services:
```bash
make build
```

3. Run services locally:
```bash
make run-local
```

## API Endpoints

All requests should be made through the API Gateway (http://localhost:8080)

### User Service

- **POST** `/api/register` - Register a new user
  ```json
  {
    "username": "john",
    "email": "john@example.com",
    "password": "password123"
  }
  ```

- **POST** `/api/login` - User login
  ```json
  {
    "username": "john",
    "password": "password123"
  }
  ```

- **GET** `/api/user?id=1` - Get user by ID

- **GET** `/api/users` - List all users

### Product Service

- **GET** `/api/products` - List all products

- **GET** `/api/product?id=1` - Get product by ID

- **POST** `/api/product/create` - Create a new product
  ```json
  {
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "stock": 10,
    "category": "Electronics"
  }
  ```

- **PUT** `/api/product/stock` - Update product stock
  ```json
  {
    "product_id": 1,
    "stock": 20
  }
  ```

### Order Service

- **GET** `/api/orders` - List all orders

- **GET** `/api/orders?user_id=1` - List orders by user ID

- **GET** `/api/order?id=1` - Get order by ID

- **POST** `/api/order/create` - Create a new order
  ```json
  {
    "user_id": 1,
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      }
    ]
  }
  ```

- **PUT** `/api/order/status` - Update order status
  ```json
  {
    "order_id": 1,
    "status": "completed"
  }
  ```

### Health Check

- **GET** `/health` - Check API Gateway health and service availability

## Example Usage

1. Register a user:
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"pass123"}'
```

2. List products:
```bash
curl http://localhost:8080/api/products
```

3. Create an order:
```bash
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"items":[{"product_id":1,"quantity":2}]}'
```

## Development

### Building Services

```bash
make build
```

### Running Tests

```bash
make test
```

### Cleaning Build Artifacts

```bash
make clean
```

## Technology Stack

- **Language**: Go 1.21+
- **Architecture**: Microservices
- **Communication**: REST API / HTTP
- **Containerization**: Docker
- **Orchestration**: Docker Compose

## Features

- ✅ Microservice architecture
- ✅ API Gateway pattern
- ✅ RESTful API design
- ✅ In-memory data storage (easily replaceable with database)
- ✅ Docker support
- ✅ Health check endpoint
- ✅ CORS support
- ✅ Modular and scalable structure

## Future Enhancements

- [ ] Database integration (PostgreSQL/MySQL)
- [ ] Authentication with JWT tokens
- [ ] Service discovery (Consul/Etcd)
- [ ] Message queue (RabbitMQ/Kafka)
- [ ] API rate limiting
- [ ] Logging and monitoring
- [ ] Unit and integration tests
- [ ] CI/CD pipeline
- [ ] Kubernetes deployment

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

# Developer Guide

## Project Overview

GoMall is a microservice-based e-commerce system demonstrating:
- Microservice architecture patterns
- API Gateway pattern
- RESTful API design
- Go best practices

## Code Organization

### Services (`services/`)

Each service is self-contained with its own `main.go` and `Dockerfile`:

- **api-gateway**: Entry point, routes requests to appropriate services
- **user-service**: User management and authentication
- **product-service**: Product catalog and inventory
- **order-service**: Order processing and management

### Common Code (`common/`)

Shared code used across all services:

- **models/**: Data structures (User, Product, Order, Response)
- **utils/**: HTTP utilities for JSON handling
- **config/**: Configuration management

## Development Workflow

### 1. Make Changes

Edit code in the appropriate service or common directory.

### 2. Build

```bash
make build
```

Or build individual services:
```bash
go build -o bin/api-gateway ./services/api-gateway/main.go
go build -o bin/user-service ./services/user-service/main.go
go build -o bin/product-service ./services/product-service/main.go
go build -o bin/order-service ./services/order-service/main.go
```

### 3. Run Locally

```bash
make run-local
```

Or start services individually:
```bash
# Terminal 1
PORT=8081 ./bin/user-service

# Terminal 2
PORT=8082 ./bin/product-service

# Terminal 3
PORT=8083 ./bin/order-service

# Terminal 4
PORT=8080 ./bin/api-gateway
```

### 4. Test

Run the automated test suite:
```bash
./test_api.sh
```

Or test individual endpoints:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/products
```

### 5. Docker Build

Build Docker images:
```bash
docker-compose build
```

Run with Docker Compose:
```bash
docker-compose up
```

## Adding New Features

### Adding a New Endpoint to an Existing Service

1. Add handler method in the service's `main.go`
2. Register route in the service's `main()` function
3. Add corresponding route in API Gateway's routing logic
4. Update README with new endpoint documentation
5. Add test cases to `test_api.sh`

Example:
```go
// In user-service/main.go
func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

func main() {
    // ...
    http.HandleFunc("/user/delete", service.DeleteUser)
    // ...
}
```

### Adding a New Service

1. Create new directory under `services/`
2. Create `main.go` with service logic
3. Create `Dockerfile` for the service
4. Add service to `docker-compose.yml`
5. Add routing logic to API Gateway
6. Update Makefile with build commands
7. Document in README

## Data Storage

Currently, all services use in-memory storage (maps). To add database support:

1. Choose database (PostgreSQL, MySQL, MongoDB, etc.)
2. Add database driver to `go.mod`
3. Create database connection utility in `common/utils/`
4. Update service structs to use database instead of maps
5. Add database to `docker-compose.yml`

Example database integration:
```go
import "database/sql"
_ "github.com/lib/pq"

type UserService struct {
    db *sql.DB
}
```

## Code Style

- Use `gofmt` for formatting
- Follow standard Go naming conventions
- Keep services independent and loosely coupled
- Use common models for data structures
- Handle errors appropriately
- Return proper HTTP status codes

## Testing

### Manual Testing
```bash
./test_api.sh
```

### Unit Tests (to be added)
```bash
go test ./...
```

### Integration Tests (to be added)
```bash
go test -tags=integration ./...
```

## Debugging

### View Service Logs

When running locally, logs appear in the terminal.

With Docker Compose:
```bash
docker-compose logs -f [service-name]
```

### Common Issues

1. **Port already in use**: Kill processes on ports 8080-8083
   ```bash
   pkill -f user-service
   pkill -f product-service
   pkill -f order-service
   pkill -f api-gateway
   ```

2. **Service not responding**: Check if all services are running
   ```bash
   curl http://localhost:8080/health
   ```

3. **Build errors**: Ensure Go modules are up to date
   ```bash
   go mod tidy
   ```

## Performance Considerations

Current implementation uses in-memory storage and is suitable for:
- Development
- Demos
- Learning microservices

For production, consider:
- Database integration
- Caching (Redis)
- Load balancing
- Service discovery
- Circuit breakers
- Rate limiting
- Logging and monitoring

## Security Notes

⚠️ **This is a demo application. Do not use in production without:**

1. **Authentication**: Add JWT or OAuth
2. **Password hashing**: Use bcrypt or similar
3. **HTTPS**: Enable TLS
4. **Input validation**: Validate all inputs
5. **Rate limiting**: Prevent abuse
6. **SQL injection protection**: Use parameterized queries when adding database
7. **CORS configuration**: Restrict to known origins

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Microservices Patterns](https://microservices.io/patterns/)
- [REST API Design](https://restfulapi.net/)
- [Docker Documentation](https://docs.docker.com/)

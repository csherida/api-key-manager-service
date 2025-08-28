# API Key Manager Service

A secure API key management service built with Go, following Clean Architecture principles. This service provides functionality for generating, validating, listing, and expiring API keys with comprehensive usage tracking.

## 🏗️ Architecture

This project is structured based on **Clean Architecture** principles, ensuring separation of concerns, testability, and maintainability.

### Directory Structure

```
api-key-manager-service/
├── internal/
│   ├── api-key-manager-service/
│   │   ├── domain/           # Enterprise Business Rules
│   │   │   ├── api_key.go
│   │   │   ├── api_usage.go
│   │   │   └── ...
│   │   ├── usecase/           # Application Business Rules
│   │   │   ├── api_key_generation.go
│   │   │   ├── api_key_validation.go
│   │   │   ├── api_key_deletion.go
│   │   │   ├── api_key_listing.go
│   │   │   └── repository.go
│   │   ├── api/               # Interface Adapters
│   │   │   ├── api_key_generation_handler.go
│   │   │   ├── api_key_validation_handler.go
│   │   │   ├── api_key_deletion_handler.go
│   │   │   ├── api_key_list_handler.go
│   │   │   └── usecase_interfaces.go
│   │   └── infra/             # Frameworks & Drivers
│   │       └── datastore.go
│   └── service/
│       └── di/                # Dependency Injection
│           ├── application.go
│           ├── providers.go
│           └── wire_gen.go
├── test/
│   └── e2e_test.go
└── cmd/
    └── main.go
```

### Clean Architecture Layers

#### 1. **Domain Layer** (`internal/api-key-manager-service/domain/`)
- Contains enterprise business rules and entities
- Pure Go structs with no external dependencies
- Defines core models: `ApiKey`, `ApiUsage`, and DTOs

#### 2. **Use Case Layer** (`internal/api-key-manager-service/usecase/`)
- Contains application business logic
- Implements specific business operations
- Defines repository interfaces (ports)
- Independent of external frameworks

#### 3. **Interface Adapters Layer** (`internal/api-key-manager-service/api/`)
- HTTP handlers that adapt use cases to web interface
- Converts data between use cases and external formats
- Handles HTTP-specific concerns (status codes, headers)

#### 4. **Infrastructure Layer** (`internal/api-key-manager-service/infra/`)
- Implements repository interfaces
- In-memory data store implementation
- Can be easily replaced with database implementations

### Dependency Flow

```
External Request → Handler → Use Case → Repository Interface
                     ↓           ↓              ↓
                  (api/)    (usecase/)   (Implemented by infra/)
```

Dependencies always point inward: Infrastructure depends on Use Cases, Use Cases depend on Domain, Domain depends on nothing.

## 🚀 Getting Started

### Prerequisites

- Go 1.23.1 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/csherida/api-key-manager-service.git
cd api-key-manager-service
```

2. Install dependencies:
```bash
go mod download
```

3. Generate Wire dependencies:
```bash
go generate ./...
```

### Running the Service

Start the server:
```bash
go run cmd/main.go
```

The service will start on `http://localhost:8080`

### Running Tests

Run end-to-end tests:
```bash
go test -tags=e2e ./test/...
```

## 📚 API Documentation

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/keys` | Generate a new API key |
| GET | `/keys` | List all API keys with usage stats |
| POST | `/keys/validate` | Validate an API key |
| DELETE | `/keys/{keyId}` | Expire an API key |

### API Examples with cURL

#### 1. Generate API Key

```bash
curl -X POST http://localhost:8080/keys \
  -H "Content-Type: application/json" \
  -d '{"organization_name": "ACME Corp"}' | jq
```

Response:
```json
{
   "api_id": "550e8400-e29b-41d4-a716-446655440000",
   "api_key": "5fb2d3e4a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8"
}
```

#### 2. List All API Keys

```bash
curl -X GET http://localhost:8080/keys | jq
```

Response:
```json
{
   "api_keys": [
      {
         "api_id": "550e8400-e29b-41d4-a716-446655440000",
         "organization_name": "ACME Corp",
         "expiration_date": null,
         "is_expired": false,
         "usage_stats": {
            "total_requests": 5,
            "last_used": "2025-08-28T10:30:00Z",
            "unique_ip_count": 2,
            "most_recent_ip": "192.168.1.100"
         }
      }
   ],
   "total": 1
}
```

#### 3. Validate API Key

```bash
# Replace <API_KEY> with the actual key returned from generation
curl -X POST http://localhost:8080/keys/validate \
  -H "Authorization: Bearer <API_KEY>" | jq
```

Response:
```json
{
   "valid": true,
   "api_id": "550e8400-e29b-41d4-a716-446655440000",
   "organization_name": "ACME Corp",
   "message": "API key is valid"
}
```

#### 4. Delete/Expire API Key

```bash
# Replace <API_ID> with the actual API ID
curl -X DELETE http://localhost:8080/keys/<API_ID> | jq
```

Response:
```json
{
   "success": true,
   "message": "API key successfully expired",
   "api_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 🔑 Key Features

- **Secure Key Generation**: Uses Ethereum's ECDSA key generation for cryptographic security
- **Usage Tracking**: Automatically tracks API key usage including request counts, IP addresses, and timestamps
- **Concurrent Safe**: Thread-safe operations using read/write mutexes
- **Clean Architecture**: Modular design allows easy replacement of components
- **Dependency Injection**: Uses Google Wire for compile-time dependency injection
- **Comprehensive Testing**: End-to-end tests covering all API endpoints

## 🛠️ Technology Stack

- **Language**: Go 1.23.1
- **Web Framework**: Gorilla Mux
- **Dependency Injection**: Google Wire
- **Cryptography**: Ethereum's go-ethereum library
- **Testing**: Stretchr Testify
- **Utilities**: Samber Lo (functional programming utilities)

## 📂 Configuration

The service uses default configuration values:
- **Port**: 8080
- **CORS**: Configured for localhost access
- **Storage**: In-memory (can be replaced with persistent storage)

## 🔄 Development Workflow

1. **Adding New Features**: Follow the Clean Architecture layers
   - Define domain models if needed
   - Create use case in `usecase/`
   - Add repository methods if needed
   - Create HTTP handler in `api/`
   - Wire dependencies in `di/`

2. **Regenerate Dependencies**: After modifying dependency injection:
   ```bash
   go generate ./...
   ```

3. **Testing**: Add tests in `test/` directory
   ```bash
   go test -tags=e2e ./test/...
   ```

## 📝 License

This project is part of a service architecture demonstration.

## 🤝 Contributing

1. Follow Clean Architecture principles
2. Maintain separation of concerns
3. Add appropriate tests
4. Update documentation as needed

## 📧 Contact

For questions or issues, please open a GitHub issue.
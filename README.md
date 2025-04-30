# Bill Aggregation Service

A scalable backend system for aggregating utility bills from multiple providers. This service allows users to link their accounts from various utility providers and view all their bills in one place.

## Architecture

This project is implemented using **Hexagonal Architecture** (also known as Ports and Adapters), which provides several benefits:

- **Separation of concerns:** The core business logic is isolated from external dependencies.
- **Testability:** Components can be easily tested by swapping out adapters.
- **Flexibility:** Support for multiple interfaces to the same core functionality.
- **Maintainability:** External systems can be changed without affecting business logic.

### Key Components

1. **Domain Layer (Core)**
   - Domain models (Bill, Account, Provider)
   - Core business logic and services

2. **Ports**
   - Primary/inbound ports (service interfaces)
   - Secondary/outbound ports (repository and API client interfaces)

3. **Adapters**
   - Primary/inbound adapters (REST API)
   - Secondary/outbound adapters (PostgreSQL repositories, HTTP API clients)

## Features

- Link accounts from multiple utility providers
- Aggregate bills from different providers
- Calculate total amount due
- View bills by provider
- Refresh bills automatically or on-demand
- Fault-tolerant integration with third-party providers

## Requirements

- Go 1.19+
- PostgreSQL 14+
- Docker and Docker Compose (for containerized deployment)

## Getting Started

### Run with Docker Compose

```bash
# Clone the repository
git clone https://github.com/bete7512/bill-aggregator.git
cd bill-aggregator

# Start the application
docker-compose -f docker/docker-compose.yml up -d
```

The service will be available at http://localhost:8080

### Local Development

```bash
# Clone the repository
git clone https://github.com/bete7512/bill-aggregator.git
cd bill-aggregator

# Install dependencies
go mod download

# Run migrations on your PostgreSQL instance
psql -U postgres -d billaggregator -f migrations/001_initial_schema.sql

# Start the application
go run cmd/server/main.go
```

## API Endpoints

### Link Utility Account
```
POST /api/v1/accounts/link
```
Request:
```json
{
  "provider_id": "electricity-provider",
  "account_details": {
    "account_number": "12345",
    "api_key": "secret-key"
  }
}
```

### Get Linked Accounts
```
GET /api/v1/accounts
```

### Delete Linked Account
```
DELETE /api/v1/accounts/{account_id}
```

### Get Aggregated Bills
```
GET /api/v1/bills
```

### Get Bills by Provider
```
GET /api/v1/bills/{provider_id}
```

### Refresh Bills
```
POST /api/v1/bills/refresh
```

## Design Decisions

### Concurrency

The system handles concurrent requests using Go's goroutines and channels. When refreshing bills from multiple providers, the requests are made in parallel to improve performance.

### Fault Tolerance

To ensure fault tolerance, the service implements:

1. **Retry with Exponential Backoff**: When third-party APIs fail, the system retries with increasing delays.
2. **Circuit Breaker Pattern**: Prevents cascading failures by stopping requests to failing services.
3. **Graceful Degradation**: Falls back to cached data when external services are unavailable.

### Database Design

The database schema is designed to efficiently support the following operations:

1. Storing user account credentials securely
2. Linking multiple provider accounts to a user
3. Storing and retrieving bills by user, provider, or status
4. Supporting fast aggregation queries

### Security Considerations

1. Credentials are stored securely as JSON objects (in a real production system, these should be encrypted)
2. JWT-based authentication for API endpoints
3. Input validation to prevent SQL injection and other attacks

## Monitoring and Observability

The service includes:

1. **Prometheus Metrics**: Exposed on `/metrics` endpoint
2. **Grafana Dashboards**: For visualizing system performance
3. **Logging**: Structured logging for easier analysis

## Future Improvements

1. Add support for more utility providers
2. Implement rate limiting to prevent API abuse
3. Add notification system for upcoming bills
4. Improve error handling and reporting
5. Add support for multi-currency bill aggregation
6. Implement user account management

## License

MIT
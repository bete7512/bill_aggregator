```
bill-aggregator/
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── internal/
│   ├── core/                        # Domain Layer (Business Logic)
│   │   ├── domain/                  # Domain models and core business logic
│   │   │   ├── model/               # Domain models
│   │   │   │   ├── account.go       # User account model
│   │   │   │   ├── bill.go          # Bill model
│   │   │   │   └── provider.go      # Provider model
│   │   │   └── service/             # Domain services
│   │   │       ├── account.go       # Account service
│   │   │       ├── bill.go          # Bill service
│   │   │       └── aggregation.go   # Bill aggregation service
│   │   └── port/                    # Ports (interfaces)
│   │       ├── inbound/             # Primary/Driving ports
│   │       │   ├── account.go       # Account service port
│   │       │   └── bill.go          # Bill service port
│   │       └── outbound/            # Secondary/Driven ports
│   │           ├── repository/      # Repository ports
│   │           │   ├── account.go   # Account repository port
│   │           │   └── bill.go      # Bill repository port
│   │           └── provider/        # Provider API ports
│   │               └── client.go    # Provider client port
│   ├── adapter/                     # Adapters implementation
│   │   ├── inbound/                 # Primary/Driving adapters
│   │   │   └── rest/                # REST API handlers
│   │   │       ├── handler/         # HTTP request handlers
│   │   │       │   ├── account.go   # Account handlers
│   │   │       │   └── bill.go      # Bill handlers
│   │   │       ├── dto/             # Data Transfer Objects
│   │   │       │   ├── account.go   # Account DTOs
│   │   │       │   └── bill.go      # Bill DTOs
│   │   │       └── middleware/      # HTTP middlewares
│   │   │           ├── auth.go      # Authentication middleware
│   │   │           └── logging.go   # Logging middleware
│   │   └── outbound/                # Secondary/Driven adapters
│   │       ├── repository/          # Repository implementations
│   │       │   ├── postgres/        # PostgreSQL implementation
│   │       │   │   ├── account.go   # Account repository
│   │       │   │   └── bill.go      # Bill repository
│   │       │   └── cache/           # Cache implementation
│   │       │       └── redis.go     # Redis cache
│   │       └── provider/            # Provider API implementations
│   │           ├── client/          # HTTP client for providers
│   │           │   └── http.go      # HTTP client
│   │           └── mock/            # Mock provider for testing
│   │               └── mock.go      # Mock provider
│   └── config/                      # Application configuration
│       └── config.go                # Configuration loader
├── pkg/                             # Public packages
│   ├── api/                         # API documentation
│   │   └── swagger.yaml            # Swagger/OpenAPI spec
│   └── util/                        # Utility functions
│       ├── logger/                  # Logging utility
│       │   └── logger.go            # Logger implementation
│       └── resilience/              # Resilience patterns
│           ├── circuitbreaker.go    # Circuit breaker
│           └── retry.go             # Retry mechanism
├── migrations/                      # Database migrations
│   └── 001_initial_schema.sql       # Initial schema
├── docker/                          # Docker configuration
│   ├── Dockerfile                   # Application Dockerfile
│   └── docker-compose.yml           # Docker Compose file
└── go.mod                           # Go module file
```
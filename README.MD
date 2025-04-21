# Go Project Template

A clean and structured Go project template using Fiber framework.

## 📁 Project Structure

```bash
.
├── app/ # Application bootstrapping
│ ├── application.go # Main application setup
│ ├── controller.go # Controller initialization
│ ├── database.go # Databases setup
│ ├── framework.go # Framework (Fiber) setup
│ ├── repository.go # Repository initialization
│ ├── router.go # Router setup
│ └── service.go # Service initialization
│
├── cmd/ # Command line interface
│ ├── arango_makemigration.go # Make migration for arangodb command
│ ├── arango_migrate.go # Migrate the json files for arangodb command
│ ├── arango_rollback.go # Rollback migration/migrations for arangodb command
│ ├── postgres_makemigration.go # Make migration for postgres command
│ ├── postgres_migrate.go # Migrate the migrations for postgres command
│ ├── postgres_rollback.go # Rollback migration/migrations for postgres command
│ ├── root.go # Root cobra command
│ └── serve.go # Server start command
│
├── docs/ # Documentation
│
├── config/ # Configuration
│ ├── config.example.yml # Configuration file example
│ └── config.yml # Configuration file
│
├── handler/ # HTTP Layer
│ ├── controllers/ # Request handlers
│ ├── errors/ # Handler layer errors
│ ├── middlewares/ # HTTP middlewares
│ └── routers/ # Route definitions
│
├── internal/ # Private application code
│ ├── config/ # Configuration
│ ├── database/ # Databases
│ ├── ent/ # Ent schema files
│ ├── mocks/ # Mock files
│ ├── producers/ # Producers
│ ├── repositories/ # Data access layer
│ ├── services/ # Business logic
│ ├── utils/ # Utility functions
│ └── helper/ # Helper functions
│
├── pkg/ # Shared packages
│
└── main.go # Application entry point
```

## 🔧 Technology Stack

- [Fiber](https://gofiber.io/) - Web framework
- [Uber FX](https://uber-go.github.io/fx/) - Dependency injection
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Cobra](https://github.com/spf13/cobra) - CLI commands

## 🏗️ Architecture

This template follows clean architecture principles:

1. **Domain Layer** (`internal/domain/`)

   - Business entities
   - Repository interfaces
   - Service interfaces

2. **Application Layer** (`internal/services/`)

   - Business logic implementation
   - Use case orchestration
   - Domain service implementation

3. **Infrastructure Layer** (`internal/repositories/`)

   - Database implementations
   - External service integrations
   - Repository implementations

4. **Interface Layer** (`handler/`)
   - HTTP handlers
   - Middleware
   - Route definitions

## 🚀 Getting Started

### installation and run

1. install dependencies

   ```bash
   go mod tidy
   ```

2. configure environment variables in config.yml

   ```bash
   nano config/config.yml.example
   nano config/config.yml
   ```

3. run the server

   ```bash
   go run main.go serve
   ```

4. Also you can set environment variable mappings by writing the following command before executing command `go run main.go serve`:

   ```bash
   APP_SERVER_PORT=8004 APP_SERVER_HOST=0.0.0.0 go run main.go serve
   ```

## 📚 ArangoDB

To see the instructions on how to work with ArangoDB in this application : [ArangoDB Documentation](./docs/arangodb.md)

# ADR 0003: Use simple layered architecture

## Status

Accepted

## Context

The project is a Junior Golang test assignment for a REST service that manages user subscription records and calculates total subscription cost for a selected period.

The service must include HTTP endpoints, PostgreSQL persistence, migrations, logging, configuration, Swagger documentation, and Docker Compose support.

The implementation should be easy to read, review, run locally, and explain during a technical interview. A heavy enterprise architecture would add complexity without improving the assignment result.

The main responsibilities are HTTP request handling, subscription business rules, and PostgreSQL access. These responsibilities should not be mixed in a single handler or package.

## Decision

Use a simple layered architecture:

- `cmd/api` starts the application and wires dependencies.
- `internal/config` loads configuration.
- `internal/db` initializes the PostgreSQL connection pool.
- `internal/httpserver` configures the HTTP server, routes, and middleware.
- `internal/subscription` contains subscription-specific handlers, services, DTOs, validation, and business logic.
- `db/migration` contains database migrations.
- `db/query` contains SQL queries for sqlc.
- `db/sqlc` contains generated database access code.
- `docs` contains API and architectural documentation.

The application will keep clear boundaries between HTTP handling, business rules, and database access.

The primary request flow is:

```text
HTTP request
  -> handler
  -> service
  -> repository/sqlc
  -> PostgreSQL
```

Expected responsibilities:

- Handler: HTTP routing, request binding, input shape validation, status codes, and JSON responses.
- Service: business rules, date parsing, validation, total cost calculation rules, and application-level errors.
- Repository/sqlc: database queries, persistence, filtering, and SQL execution.

The project should avoid a full Clean Architecture, Hexagonal Architecture, CQRS, event bus, dependency injection container, and domain model over-engineering.

## Consequences

Positive:

- The project remains small and understandable.
- Business logic can be tested separately from HTTP handlers.
- Database access stays isolated from request handling.
- HTTP concerns do not leak into business logic.
- SQL/database concerns stay outside handlers.
- The structure is close to common Go backend practice without overengineering.
- Codex can make focused changes with lower risk of modifying unrelated layers.

Negative:

- The architecture is less abstract than full Clean Architecture.
- Some package boundaries may need adjustment if the project grows.
- The repository layer depends on generated sqlc code.
- A repository wrapper over sqlc may look slightly redundant for a small service, but it keeps database access isolated.

## Practical mapping

Recommended package structure for the subscription feature:

```text
internal/subscription/
├── handler.go
├── service.go
├── repository.go
├── dto.go
├── model.go
└── errors.go
```

The handler should not execute SQL directly.

The repository should not know about HTTP status codes, Gin contexts, or JSON responses.

The service should own the main subscription rules, especially date validation and total cost calculation behavior.

## Alternatives considered

### Flat package structure

Rejected because mixing handlers, database code, configuration, and business logic would make the project harder to review, test, and extend.

### Full Clean Architecture

Rejected because it would introduce unnecessary abstractions for a small test assignment.

### Framework-style project layout

Rejected because the goal is to demonstrate clear Go backend fundamentals rather than framework-specific conventions.

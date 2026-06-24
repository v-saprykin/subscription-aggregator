# ADR 0001: Use Go, Gin, PostgreSQL, pgx, sqlc, and golang-migrate

## Status

Accepted

## Context

The project is a REST API service for managing user subscriptions and calculating total subscription cost for a selected period. The service must be small enough for a Junior Golang test assignment, but still demonstrate practical backend engineering: HTTP API design, PostgreSQL persistence, migrations, logging, configuration, Swagger documentation, and Docker Compose runtime.

The project should avoid unnecessary production complexity. It does not need authentication, user management, gRPC, Redis, background workers, Kubernetes, event-driven architecture, or a full Clean Architecture implementation.

## Decision

Use the following core stack:

- Go as the application language
- Gin as the HTTP router/framework
- PostgreSQL as the database
- pgx/v5 as the PostgreSQL driver
- sqlc for type-safe Go code generation from SQL queries
- golang-migrate for database migrations
- slog for structured logging
- Viper or equivalent environment-based configuration loading
- Swagger/OpenAPI documentation for the REST API
- Docker Compose for local runtime in a later implementation step

Use sqlc instead of an ORM. SQL queries should remain explicit and reviewable, while Go database access code should be generated and type-safe.

## Consequences

Positive:

- The stack is common and appropriate for a Go backend test assignment.
- SQL remains explicit and easy to review.
- sqlc reduces manual scanning and mapping boilerplate.
- pgx/v5 provides a modern PostgreSQL driver.
- golang-migrate gives reproducible schema initialization.
- Gin keeps HTTP routing simple and familiar.
- slog avoids an extra logging dependency.

Negative:

- sqlc adds a generation step after query or schema changes.
- Developers must understand SQL instead of relying on ORM abstractions.
- Gin introduces a framework dependency instead of using only net/http.
- The project needs clear documentation so the toolchain remains easy to run locally.

## Alternatives considered

### net/http only

Rejected for this project. It would reduce dependencies, but Gin provides faster API scaffolding, request binding, routing groups, and Swagger integration patterns that are useful for a test assignment.

### GORM

Rejected. GORM would reduce some SQL boilerplate, but it hides SQL details and can make aggregation logic less explicit. For this project, explicit SQL is preferable.

### sqlx or raw pgx queries

Rejected for the initial implementation. Both are valid options, but sqlc gives stronger compile-time guarantees and less manual row scanning.

# AGENTS.md

## Project context
This repository implements a Junior Golang Developer test assignment: a REST service for aggregating online subscription data.

## Primary goal
Build a small, correct, readable Go backend service that satisfies `TASK.md` without overengineering.

## Technology choices
- Go.
- Gin for HTTP routing.
- PostgreSQL.
- pgx/v5 for PostgreSQL access.
- sqlc for type-safe SQL code generation.
- golang-migrate for migrations.
- slog for structured logging.
- Viper for configuration from environment variables.
- Swagger documentation.
- Docker Compose for local startup.

## Architecture constraints
- Keep the architecture simple.
- Prefer explicit code over hidden framework behavior.
- Keep HTTP, business logic, and database access separated.
- Do not introduce Clean Architecture ceremony beyond what is useful for this small service.
- Use `context.Context` for request-scoped database operations.
- Return clear HTTP errors for validation and persistence failures.

## Suggested project layout
- `cmd/api/main.go`: application entry point.
- `internal/config`: configuration loading.
- `internal/httpserver`: server setup, routes, middleware.
- `internal/subscription`: handlers, service logic, DTOs, validation.
- `internal/db/sqlc`: generated sqlc package.
- `db/migration`: SQL migrations.
- `db/query`: sqlc SQL queries.
- `docs`: Swagger/OpenAPI files or generated docs.

## Do not add
- Authentication.
- Authorization.
- User management.
- User existence checks.
- gRPC.
- Redis.
- Background workers.
- Email sending.
- Kubernetes manifests.
- Cloud deployment files.
- Frontend code.

## Database rules
- Use explicit SQL.
- Store API dates from `MM-YYYY` as PostgreSQL `DATE` values using the first day of the month.
- Store price as integer rubles.
- Use UUID for subscription `id` and `user_id`.
- Keep migrations reversible.
- Do not modify existing migration files after they are considered applied; create a new migration instead.

## Total calculation rules
- Request period is inclusive.
- Subscription period is inclusive.
- Count only overlapping months.
- `end_date = null` means active indefinitely.
- Filters by `user_id` and `service_name` are optional.
- Add tests for overlapping date ranges.

## API rules
- Use `/api/v1` prefix.
- Use JSON request and response bodies.
- Validate incoming DTOs before calling the service layer.
- Keep external date format as `MM-YYYY`.
- Do not expose PostgreSQL-specific details in API responses.

## Logging rules
- Use structured logs through `log/slog`.
- Log server startup and shutdown.
- Log request method, path, status, duration, and request id if implemented.
- Log internal errors with enough context for debugging.
- Do not log secrets.

## Configuration rules
- Commit `.env.example` only.
- Do not commit real `.env` files.
- Required configuration should include server address, database URL, log level, and environment name.

## Swagger rules
- Document all implemented endpoints.
- Document request bodies, response bodies, query parameters, and path parameters.
- Keep Swagger synchronized with the actual API.

## Testing expectations
- Add unit tests for date parsing and month-overlap calculation.
- Add tests for validation rules.
- Add service tests where practical.
- Add repository or integration tests only if they remain simple and reproducible.

## Before finishing a coding task
Run:
- `gofmt ./...`
- `go test ./...`
- `go vet ./...`
- `sqlc generate`
- `docker compose config`

## Coding style
- Use idiomatic Go names.
- Return errors explicitly.
- Wrap errors where it helps debugging.
- Avoid global mutable state except initialized logger/config where appropriate.
- Keep functions short enough to read without scrolling too much.
- Prefer standard library unless a dependency has a clear purpose.

## Work process for Codex
- First read `TASK.md`, `ENDPOINTS.md`, `DB_SCHEMA.md`, and this file.
- Create or change a small set of files per iteration.
- Explain what was changed after each iteration.
- Run verification commands when possible.
- Do not silently change the chosen stack.

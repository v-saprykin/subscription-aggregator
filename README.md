# Subscription Aggregator

REST service for storing user online subscriptions and calculating total subscription cost for a selected period.

## Status
Project skeleton for the Effective Mobile Junior Golang Developer test assignment.

## Stack
- Go.
- Gin.
- PostgreSQL.
- pgx/v5.
- sqlc.
- golang-migrate.
- slog.
- Viper.
- Swagger.
- Docker Compose.

## Features
- Create subscription.
- Read subscription by id.
- List subscriptions.
- Update subscription.
- Delete subscription.
- Calculate total subscription cost for a selected period.
- Filter total calculation by `user_id` and `service_name`.

## Project structure
```text
.
├── cmd/api
├── internal/config
├── internal/httpserver
├── internal/subscription
├── internal/db/sqlc
├── db/migration
├── db/query
├── docs
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── sqlc.yaml
├── TASK.md
├── ENDPOINTS.md
├── DB_SCHEMA.md
└── AGENTS.md
```

## Configuration
Copy example env file:

```bash
cp .env.example .env
```

Required variables:
- `APP_ENV`: application environment, for example `local`.
- `HTTP_ADDR`: HTTP server address, for example `:8080`.
- `DATABASE_URL`: PostgreSQL connection string.
- `LOG_LEVEL`: log level, for example `debug`, `info`, `warn`, or `error`.

## Local startup
Start all services:

```bash
docker compose up --build
```

Stop all services:

```bash
docker compose down
```

Stop and remove database volume:

```bash
docker compose down -v
```

## Migrations
Create a new migration:

```bash
migrate create -ext sql -dir db/migration -seq migration_name
```

Apply migrations:

```bash
migrate -path db/migration -database "$DATABASE_URL" up
```

Rollback last migration:

```bash
migrate -path db/migration -database "$DATABASE_URL" down 1
```

## sqlc
Generate database code:

```bash
sqlc generate
```

## API
Base URL:

```text
http://localhost:8080/api/v1
```

Endpoints:
- `POST /subscriptions`
- `GET /subscriptions`
- `GET /subscriptions/{id}`
- `PUT /subscriptions/{id}`
- `DELETE /subscriptions/{id}`
- `GET /subscriptions/total`

Full endpoint contract is described in `ENDPOINTS.md`.

## Example requests
Create subscription:

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'
```

Calculate total:

```bash
curl 'http://localhost:8080/api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus'
```

## Swagger
Swagger documentation should be available at one of these locations after implementation:
- `docs/swagger.yaml`
- `http://localhost:8080/swagger/index.html`

## Verification
Run before submitting:

```bash
gofmt ./...
go test ./...
go vet ./...
sqlc generate
docker compose config
docker compose up --build
```

## Submission checklist
- GitHub repository is public or accessible to reviewers.
- README contains setup and API usage instructions.
- Docker Compose startup works from a clean checkout.
- Migrations initialize PostgreSQL.
- Swagger documentation matches implemented endpoints.
- No real `.env` or secrets are committed.

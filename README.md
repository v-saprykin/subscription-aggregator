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
- `DATABASE_URL`: PostgreSQL connection string. For host tools, use the `localhost` URL from `.env.example`.
- `LOG_LEVEL`: log level, for example `debug`, `info`, `warn`, or `error`.

Docker Compose supplies these variables to the API automatically and uses `postgres` as the database hostname.

## Docker Compose

The normal local workflow starts PostgreSQL, applies all migrations, and then starts the API:

```bash
docker compose up --build
```

The API is available at `http://localhost:8080` after PostgreSQL becomes healthy and the `migrate` service completes successfully. A manual host-side migration command is not required for this workflow.

Start the full environment in detached mode:

```bash
docker compose up -d --build
```

Check the running services and the completed one-shot migration service:

```bash
docker compose ps --all
```

Check API health:

```bash
curl http://localhost:8080/healthz
```

Check the subscriptions endpoint:

```bash
curl http://localhost:8080/api/v1/subscriptions
```

View service logs:

```bash
docker compose logs -f api
docker compose logs -f postgres
docker compose logs migrate
```

Stop the services while keeping the PostgreSQL volume:

```bash
docker compose down
```

Stop the services and remove the PostgreSQL volume:

```bash
docker compose down -v
```

## Local PostgreSQL
For non-Compose API development, start only PostgreSQL:

```bash
docker compose up -d postgres
```

Stop PostgreSQL and keep the database volume:

```bash
docker compose down
```

Stop and remove database volume:

```bash
docker compose down -v
```

The local PostgreSQL service uses:

```text
host: localhost
port: 5433
database: subscription_aggregator
user: subscription_aggregator
password: subscription_aggregator_password
```

## Docker image
Build the API image:

```bash
docker build -t subscription-aggregator-api .
```

Run the API container against the local PostgreSQL exposed by Docker Compose:

```bash
docker run --rm --name subscription-aggregator-api \
  -p 8080:8080 \
  -e APP_ENV=local \
  -e HTTP_ADDR=:8080 \
  -e LOG_LEVEL=debug \
  -e DATABASE_URL='postgres://subscription_aggregator:subscription_aggregator_password@host.docker.internal:5433/subscription_aggregator?sslmode=disable' \
  subscription-aggregator-api
```

## Migrations
Docker Compose runs migrations automatically during full-environment startup. The manual commands below remain useful when running PostgreSQL and the API separately for host-side development.

Create a new migration:

```bash
migrate create -ext sql -dir db/migration -seq migration_name
```

Apply migrations:

```bash
migrate -path db/migration -database "postgres://subscription_aggregator:subscription_aggregator_password@localhost:5433/subscription_aggregator?sslmode=disable" up
```

Rollback last migration:

```bash
migrate -path db/migration -database "postgres://subscription_aggregator:subscription_aggregator_password@localhost:5433/subscription_aggregator?sslmode=disable" down 1
```

Inspect the `subscriptions` table schema:

```bash
docker compose exec postgres psql -U subscription_aggregator -d subscription_aggregator -c "\d subscriptions"
```

Inspect subscription rows:

```bash
docker compose exec postgres psql -U subscription_aggregator -d subscription_aggregator -c "SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 20;"
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

Swagger documents the existing API contract and is served by the API during normal local and Docker Compose startup.

Install the Swagger generator:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Regenerate the committed files under `docs/` after changing API routes, request or response models, or Swagger annotations:

```bash
swag init -g cmd/api/main.go
```

Open Swagger UI while the API is running:

```text
http://localhost:8080/swagger/index.html
```

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

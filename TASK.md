# TASK.md

## Goal
Implement a REST service for aggregating user online subscription data.

## Source
This task is based on the Effective Mobile Junior Golang Developer test assignment.

## Required features
- CRUDL operations for subscription records.
- Endpoint for calculating the total subscription cost for a selected period.
- Optional filtering in total calculation by `user_id` and `service_name`.
- PostgreSQL as the database.
- SQL migrations for database initialization.
- Structured application logs.
- Configuration loaded from `.env` or `.yaml`.
- Swagger documentation for the implemented API.
- Service startup through Docker Compose.

## Subscription record
Each subscription contains:
- `id`: internal subscription identifier.
- `service_name`: name of the subscription service.
- `price`: monthly subscription price in integer rubles.
- `user_id`: user identifier in UUID format.
- `start_date`: subscription start date in `MM-YYYY` format at API boundary.
- `end_date`: optional subscription end date in `MM-YYYY` format at API boundary.
- `created_at`: record creation timestamp.
- `updated_at`: record update timestamp.

## Out of scope
- User management.
- User existence validation.
- Authentication and authorization.
- gRPC.
- Redis or background workers.
- Kubernetes or cloud deployment.
- Payment processing.
- Fractional ruble amounts.

## Date rules
- API date format: `MM-YYYY`.
- PostgreSQL storage format: `DATE`.
- Store subscription month as the first day of the month.
- Example: `07-2025` is stored as `2025-07-01`.
- `end_date` is nullable.
- If `end_date` is null, the subscription is considered active indefinitely.
- `end_date` must be greater than or equal to `start_date` when provided.

## Total cost calculation rules
- Request period is inclusive.
- Subscription active period is inclusive.
- Only overlapping months are counted.
- Formula: `total += price * active_month_count_inside_requested_period`.
- If there is no overlap, the subscription contributes `0`.
- If `service_name` filter is provided, only matching service names are counted.
- If `user_id` filter is provided, only matching user subscriptions are counted.

## Total cost examples
- Subscription `07-2025`, no `end_date`, request `07-2025..07-2025`: 1 month.
- Subscription `07-2025..12-2025`, request `01-2025..12-2025`: 6 months.
- Subscription `07-2025..12-2025`, request `08-2025..10-2025`: 3 months.
- Subscription `07-2025`, no `end_date`, request `07-2025..09-2025`: 3 months.
- Subscription `07-2025..08-2025`, request `09-2025..12-2025`: 0 months.

## API endpoints
See `ENDPOINTS.md`.

## Database schema
See `DB_SCHEMA.md` and `db/migration/000001_create_subscriptions.up.sql`.

## Verification checklist
- `gofmt -w .`
- `go test ./...`
- `go vet ./...`
- `sqlc generate`
- `docker compose config`
- `docker compose up --build`
- Manual CRUD requests with `curl`.
- Manual total calculation requests with known test data.

## Acceptance criteria
- Project starts with Docker Compose.
- PostgreSQL is initialized through migrations.
- All required REST endpoints work.
- CRUDL operations persist data in PostgreSQL.
- Total calculation handles overlapping periods correctly.
- API validates UUID, price, and `MM-YYYY` dates.
- Swagger documentation is available or generated.
- README explains setup and usage.
- No secrets are committed.

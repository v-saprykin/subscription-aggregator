# ADR 0004: Use Docker Compose for local environment

## Status

Accepted

## Context

The project needs PostgreSQL for local development and for running database migrations. The application is not connected to PostgreSQL yet, and there is no application Dockerfile in the current implementation step.

Local setup should stay small and reproducible. Developers should be able to start the database, run migrations from the host with golang-migrate, inspect the schema, and stop the database without installing PostgreSQL directly.

## Decision

Use Docker Compose for local infrastructure services.

The initial Compose file contains only PostgreSQL:

- service name: `postgres`
- database: `subscription_aggregator`
- user: `subscription_aggregator`
- password: `subscription_aggregator_password`
- host port: `127.0.0.1:5433`
- named volume: `subscription_aggregator_postgres_data`

Migrations continue to live in `db/migration` and are run with the golang-migrate CLI against the local PostgreSQL URL.

The application service will be added to Docker Compose only after the application has PostgreSQL connection wiring and an application Dockerfile.

## Consequences

Positive:

- Local PostgreSQL startup is reproducible.
- The project does not require a system PostgreSQL installation.
- Migrations can be tested against the same database image used by local development.
- Compose remains focused on infrastructure until the application runtime is ready.

Negative:

- Developers need Docker and Docker Compose installed.
- PostgreSQL listens on port `5432` inside the container.
- The service is exposed on the host as `127.0.0.1:5433`.
- The host migration URL uses `localhost`, while future containers should use the Compose service hostname `postgres`.

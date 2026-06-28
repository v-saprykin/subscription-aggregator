# ADR 0005: Use swaggo for OpenAPI documentation

## Status

Accepted

## Context

The project exposes a Gin REST API whose routes, request models, response models, validation failures, and status codes need an accurate, reviewable contract. OpenAPI documentation makes the implemented API easier to inspect and try locally, and it satisfies the assignment requirement for Swagger documentation.

The documentation approach should fit the existing Go and Gin code without introducing a separate API implementation or changing endpoint behavior.

## Decision

Use swaggo to generate a Swagger 2.0/OpenAPI description from annotations near the Gin handlers and their request and response models. Swaggo is designed to integrate directly with Gin, so the documentation can remain close to the code that defines the API contract.

Use `github.com/swaggo/gin-swagger` and `github.com/swaggo/files` to expose Swagger UI locally at:

```text
http://localhost:8080/swagger/index.html
```

The generated `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` files are committed to the repository. Any API change requires updating the relevant annotations and regenerating the files with:

```bash
swag init -g cmd/api/main.go
```

## Alternatives considered

- Maintain a manual OpenAPI YAML file. This avoids code annotations, but duplicates the contract by hand and makes handler/model changes easier to miss.
- Use `oapi-codegen`. A specification-first, generated-server approach is useful for larger APIs, but it would add unnecessary generation and integration complexity to this already implemented, small Gin service.
- Commit only a static OpenAPI specification without Swagger UI. This would provide a machine-readable contract, but would make local discovery and interactive endpoint inspection less convenient.

## Consequences

Positive:

- Swagger UI is available during normal API startup without a separate documentation service.
- The generated specification documents the current routes and JSON field names.
- Annotations live near the handlers and models they describe, making contract changes visible during code review.

Negative:

- The application gains new swaggo runtime dependencies.
- Generated files are stored under `docs/` and create repository diffs when the contract changes.
- Handler and model code includes documentation annotations.
- Documentation can drift from the implementation when annotations are not updated or the generated files are not regenerated.

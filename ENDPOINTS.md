# ENDPOINTS.md

## General rules
- Base path: `/api/v1`.
- Request format: JSON.
- Response format: JSON.
- API date format: `MM-YYYY`.
- UUID format: canonical UUID string.
- Price format: integer rubles.

## Error response format
```json
{
  "error": "validation_error",
  "message": "start_date must use MM-YYYY format"
}
```

## Subscription response format
```json
{
  "id": "2f8d9b27-5b9e-4d6f-83ef-cf4fef0c9fc2",
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": null,
  "created_at": "2026-06-23T10:00:00Z",
  "updated_at": "2026-06-23T10:00:00Z"
}
```

## Create subscription
```http
POST /api/v1/subscriptions
```

Request body:
```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

Validation:
- `service_name` is required and must not be blank.
- `price` is required and must be greater than `0`.
- `user_id` is required and must be UUID.
- `start_date` is required and must use `MM-YYYY` format.
- `end_date` is optional and must use `MM-YYYY` format when provided.
- `end_date` must not be earlier than `start_date`.

Responses:
- `201 Created`: created subscription.
- `400 Bad Request`: validation error.
- `500 Internal Server Error`: unexpected server error.

## Get subscription by id
```http
GET /api/v1/subscriptions/{id}
```

Path parameters:
- `id`: subscription UUID.

Responses:
- `200 OK`: subscription.
- `400 Bad Request`: invalid id.
- `404 Not Found`: subscription not found.
- `500 Internal Server Error`: unexpected server error.

## List subscriptions
```http
GET /api/v1/subscriptions
```

Query parameters:
- `limit`: optional integer, default `20`, max `100`.
- `offset`: optional integer, default `0`.
- `user_id`: optional UUID filter.
- `service_name`: optional exact service name filter.

Example:
```http
GET /api/v1/subscriptions?limit=20&offset=0&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus
```

Response body:
```json
{
  "items": [],
  "limit": 20,
  "offset": 0
}
```

Responses:
- `200 OK`: subscription list.
- `400 Bad Request`: invalid query parameter.
- `500 Internal Server Error`: unexpected server error.

## Update subscription
```http
PUT /api/v1/subscriptions/{id}
```

Path parameters:
- `id`: subscription UUID.

Request body:
```json
{
  "service_name": "Yandex Plus",
  "price": 500,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

Responses:
- `200 OK`: updated subscription.
- `400 Bad Request`: validation error.
- `404 Not Found`: subscription not found.
- `500 Internal Server Error`: unexpected server error.

## Delete subscription
```http
DELETE /api/v1/subscriptions/{id}
```

Path parameters:
- `id`: subscription UUID.

Responses:
- `204 No Content`: subscription deleted.
- `400 Bad Request`: invalid id.
- `404 Not Found`: subscription not found.
- `500 Internal Server Error`: unexpected server error.

## Calculate total subscription cost
```http
GET /api/v1/subscriptions/total
```

Query parameters:
- `from`: required period start in `MM-YYYY` format.
- `to`: required period end in `MM-YYYY` format.
- `user_id`: optional UUID filter.
- `service_name`: optional exact service name filter.

Example:
```http
GET /api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus
```

Response body:
```json
{
  "total_price": 2400,
  "period_from": "07-2025",
  "period_to": "12-2025",
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "service_name": "Yandex Plus"
}
```

Validation:
- `from` is required.
- `to` is required.
- `from` and `to` must use `MM-YYYY` format.
- `to` must not be earlier than `from`.
- `user_id` must be UUID when provided.

Responses:
- `200 OK`: calculated total.
- `400 Bad Request`: validation error.
- `500 Internal Server Error`: unexpected server error.

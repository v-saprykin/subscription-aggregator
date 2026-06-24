-- name: CreateSubscription :one
INSERT INTO subscriptions (
    service_name,
    price,
    user_id,
    start_date,
    end_date
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at;

-- name: GetSubscription :one
SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE id = $1;

-- name: ListSubscriptions :many
SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE (sqlc.narg(user_id_filter)::uuid IS NULL OR user_id = sqlc.narg(user_id_filter)::uuid)
  AND (sqlc.narg(service_name_filter)::text IS NULL OR service_name = sqlc.narg(service_name_filter)::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_value)::int
OFFSET sqlc.arg(offset_value)::int;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET service_name = $2,
    price = $3,
    user_id = $4,
    start_date = $5,
    end_date = $6
WHERE id = $1
RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE id = $1;

-- name: CalculateTotalPrice :one
SELECT COALESCE(SUM(
    price * (
        EXTRACT(YEAR FROM age(
            LEAST(COALESCE(end_date, sqlc.arg(period_to)::date), sqlc.arg(period_to)::date),
            GREATEST(start_date, sqlc.arg(period_from)::date)
        ))::int * 12
        + EXTRACT(MONTH FROM age(
            LEAST(COALESCE(end_date, sqlc.arg(period_to)::date), sqlc.arg(period_to)::date),
            GREATEST(start_date, sqlc.arg(period_from)::date)
        ))::int
        + 1
    )
), 0)::bigint AS total_price
FROM subscriptions
WHERE start_date <= sqlc.arg(period_to)::date
  AND COALESCE(end_date, sqlc.arg(period_to)::date) >= sqlc.arg(period_from)::date
  AND (sqlc.narg(user_id_filter)::uuid IS NULL OR user_id = sqlc.narg(user_id_filter)::uuid)
  AND (sqlc.narg(service_name_filter)::text IS NULL OR service_name = sqlc.narg(service_name_filter)::text);

# DB_SCHEMA.md

## Database
PostgreSQL.

## Extensions
- `pgcrypto`: used for `gen_random_uuid()`.

## Table: subscriptions
Stores online subscription records.

Columns:
- `id uuid primary key`: internal subscription identifier.
- `service_name text not null`: subscription service name.
- `price integer not null`: monthly price in integer rubles.
- `user_id uuid not null`: user identifier from an external user service.
- `start_date date not null`: first day of the subscription start month.
- `end_date date null`: first day of the subscription end month, nullable.
- `created_at timestamptz not null`: creation timestamp.
- `updated_at timestamptz not null`: last update timestamp.

Constraints:
- `service_name` must not be blank.
- `price` must be greater than `0`.
- `end_date` must be null or greater than or equal to `start_date`.

Indexes:
- `idx_subscriptions_user_id` on `user_id`.
- `idx_subscriptions_service_name` on `service_name`.
- `idx_subscriptions_period` on `start_date, end_date`.

## Date storage rule
API receives dates as `MM-YYYY` strings, but database stores them as `DATE` values.

Examples:
- `07-2025` becomes `2025-07-01`.
- `12-2025` becomes `2025-12-01`.

## Total calculation rule
For each subscription:
- Calculate overlap between requested period and subscription active period.
- Count months inclusively.
- Add `price * active_months` to total.

## Null end date rule
If `end_date` is null, the subscription is treated as active until the end of the requested period for calculation purposes.

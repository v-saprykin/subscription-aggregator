# ADR 0002: Store subscription months as PostgreSQL date values

## Status

Accepted

## Context

The API receives subscription dates as month-year values in `MM-YYYY` format. A subscription has a required start date and an optional end date. The service also has to calculate total subscription cost for a selected period.

The business domain is month-based, not day-based. Prices are monthly integer ruble amounts. The total calculation must count how many months a subscription was active inside a requested period.

Storing dates as plain text would preserve the input format, but it would make validation, sorting, filtering, comparison, and aggregation less reliable.

## Decision

The API will accept dates in `MM-YYYY` format.

The application will parse these values and store them in PostgreSQL as `DATE` values using the first day of the month.

Examples:

- `07-2025` is stored as `2025-07-01`
- `12-2025` is stored as `2025-12-01`

The API will return dates back to clients in `MM-YYYY` format.

Billing periods are inclusive by month:

- `07-2025` to `07-2025` means 1 active month
- `07-2025` to `12-2025` means 6 active months
- a null `end_date` means the subscription is active indefinitely

Total cost is calculated from the intersection between the subscription period and the requested period.

## Consequences

Positive:

- PostgreSQL can compare and filter dates correctly.
- The schema keeps temporal data in a native date type instead of text.
- Month-based calculations can be implemented consistently.
- API formatting remains user-friendly while storage remains database-friendly.
- Invalid dates can be rejected at the application boundary.

Negative:

- The application must convert between `MM-YYYY` and `DATE`.
- Developers must remember that stored dates always represent the first day of a month.
- Total calculation logic must be tested carefully to avoid off-by-one month errors.

## Business rule

For total cost calculation:

```text
requested period:     from_month ... to_month
subscription period:  start_month ... end_month/null
active months:        intersection(requested period, subscription period)
total contribution:   subscription price * active months
```

If there is no overlap between the requested period and the subscription period, the subscription contributes `0`.

## Examples

- Subscription `07-2025..12-2025`, query `01-2025..12-2025` contributes 6 months.
- Subscription `07-2025..12-2025`, query `08-2025..10-2025` contributes 3 months.
- Subscription `07-2025..null`, query `07-2025..09-2025` contributes 3 months.
- Subscription `07-2025..08-2025`, query `09-2025..12-2025` contributes 0 months.

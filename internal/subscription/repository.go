package subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sqldb "github.com/v-saprykin/subscription-aggregator/internal/db/sqlc"
)

type Repository struct {
	queries sqldb.Querier
}

func NewRepository(queries sqldb.Querier) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Create(ctx context.Context, input UpsertSubscription) (Subscription, error) {
	row, err := r.queries.CreateSubscription(ctx, sqldb.CreateSubscriptionParams{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   pgDate(input.StartDate),
		EndDate:     nullablePGDate(input.EndDate),
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	subscription, err := subscriptionFromRow(row)
	if err != nil {
		return Subscription{}, fmt.Errorf("map created subscription: %w", err)
	}

	return subscription, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Subscription, error) {
	row, err := r.queries.GetSubscription(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, fmt.Errorf("get subscription: %w", err)
	}

	subscription, err := subscriptionFromRow(row)
	if err != nil {
		return Subscription{}, fmt.Errorf("map subscription: %w", err)
	}

	return subscription, nil
}

func (r *Repository) List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, error) {
	rows, err := r.queries.ListSubscriptions(ctx, sqldb.ListSubscriptionsParams{
		UserIDFilter:      uuidFilter(filter.UserID),
		ServiceNameFilter: textFilter(filter.ServiceName),
		OffsetValue:       filter.Offset,
		LimitValue:        filter.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	items := make([]Subscription, 0, len(rows))
	for _, row := range rows {
		subscription, err := subscriptionFromRow(row)
		if err != nil {
			return nil, fmt.Errorf("map subscription: %w", err)
		}
		items = append(items, subscription)
	}

	return items, nil
}

func (r *Repository) CalculateTotalPrice(ctx context.Context, filter TotalPriceFilter) (int64, error) {
	total, err := r.queries.CalculateTotalPrice(ctx, sqldb.CalculateTotalPriceParams{
		PeriodTo:          pgDate(filter.PeriodTo),
		PeriodFrom:        pgDate(filter.PeriodFrom),
		UserIDFilter:      uuidFilter(filter.UserID),
		ServiceNameFilter: textFilter(filter.ServiceName),
	})
	if err != nil {
		return 0, fmt.Errorf("calculate total price: %w", err)
	}

	return total, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpsertSubscription) (Subscription, error) {
	row, err := r.queries.UpdateSubscription(ctx, sqldb.UpdateSubscriptionParams{
		ID:          id,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   pgDate(input.StartDate),
		EndDate:     nullablePGDate(input.EndDate),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, fmt.Errorf("update subscription: %w", err)
	}

	subscription, err := subscriptionFromRow(row)
	if err != nil {
		return Subscription{}, fmt.Errorf("map updated subscription: %w", err)
	}

	return subscription, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.Get(ctx, id); err != nil {
		return err
	}

	if err := r.queries.DeleteSubscription(ctx, id); err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	return nil
}

func subscriptionFromRow(row sqldb.Subscription) (Subscription, error) {
	startDate, ok := timeFromPGDate(row.StartDate)
	if !ok {
		return Subscription{}, fmt.Errorf("invalid start_date")
	}

	createdAt, ok := timeFromPGTimestamptz(row.CreatedAt)
	if !ok {
		return Subscription{}, fmt.Errorf("invalid created_at")
	}

	updatedAt, ok := timeFromPGTimestamptz(row.UpdatedAt)
	if !ok {
		return Subscription{}, fmt.Errorf("invalid updated_at")
	}

	subscription := Subscription{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		UserID:      row.UserID,
		StartDate:   startDate,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if endDate, ok := timeFromPGDate(row.EndDate); ok {
		subscription.EndDate = &endDate
	}

	return subscription, nil
}

func uuidFilter(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{
		Bytes: [16]byte(*value),
		Valid: true,
	}
}

func textFilter(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

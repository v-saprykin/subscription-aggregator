package subscription

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqldb "github.com/v-saprykin/subscription-aggregator/internal/db/sqlc"
)

const integrationTestTimeout = 5 * time.Second

var (
	integrationUserOne = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	integrationUserTwo = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestRepositoryCalculateTotalPriceIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	pool := newIntegrationTestPool(ctx, t)
	defer pool.Close()

	requireSubscriptionsTable(ctx, t, pool)

	tests := []struct {
		name     string
		fixtures []integrationSubscriptionFixture
		filter   TotalPriceFilter
		want     int64
	}{
		{
			name: "full overlap",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					ServiceName: "Full Overlap",
					Price:       400,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     monthPtr(t, "12-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "07-2025"),
				PeriodTo:   mustAPIMonth(t, "12-2025"),
			},
			want: 2400,
		},
		{
			name: "partial overlap at beginning",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					ServiceName: "Partial Beginning",
					Price:       1000,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "03-2025"),
					EndDate:     monthPtr(t, "08-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "06-2025"),
				PeriodTo:   mustAPIMonth(t, "12-2025"),
			},
			want: 3000,
		},
		{
			name: "partial overlap at end",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					ServiceName: "Partial End",
					Price:       700,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "10-2025"),
					EndDate:     monthPtr(t, "03-2026"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "07-2025"),
				PeriodTo:   mustAPIMonth(t, "12-2025"),
			},
			want: 2100,
		},
		{
			name: "single-month overlap",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					ServiceName: "Single Month",
					Price:       500,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "06-2025"),
					EndDate:     monthPtr(t, "06-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "06-2025"),
				PeriodTo:   mustAPIMonth(t, "06-2025"),
			},
			want: 500,
		},
		{
			name: "open-ended subscription",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					ServiceName: "Open Ended",
					Price:       250,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     nil,
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "07-2025"),
				PeriodTo:   mustAPIMonth(t, "09-2025"),
			},
			want: 750,
		},
		{
			name: "no overlap",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000006"),
					ServiceName: "No Overlap",
					Price:       900,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "01-2025"),
					EndDate:     monthPtr(t, "03-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "07-2025"),
				PeriodTo:   mustAPIMonth(t, "12-2025"),
			},
			want: 0,
		},
		{
			name: "filter by user_id",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000007"),
					ServiceName: "User Filter",
					Price:       400,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     monthPtr(t, "12-2025"),
				},
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000008"),
					ServiceName: "User Filter",
					Price:       1000,
					UserID:      integrationUserTwo,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     monthPtr(t, "12-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom: mustAPIMonth(t, "07-2025"),
				PeriodTo:   mustAPIMonth(t, "12-2025"),
				UserID:     &integrationUserOne,
			},
			want: 2400,
		},
		{
			name: "filter by service_name",
			fixtures: []integrationSubscriptionFixture{
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000009"),
					ServiceName: "Matched Service",
					Price:       300,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     monthPtr(t, "12-2025"),
				},
				{
					ID:          uuid.MustParse("00000000-0000-0000-0000-000000000010"),
					ServiceName: "Ignored Service",
					Price:       1000,
					UserID:      integrationUserOne,
					StartDate:   mustAPIMonth(t, "07-2025"),
					EndDate:     monthPtr(t, "12-2025"),
				},
			},
			filter: TotalPriceFilter{
				PeriodFrom:  mustAPIMonth(t, "07-2025"),
				PeriodTo:    mustAPIMonth(t, "12-2025"),
				ServiceName: stringPtr("Matched Service"),
			},
			want: 1800,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
			defer cancel()

			tx := beginIntegrationTestTx(ctx, t, pool)
			defer rollbackIntegrationTestTx(context.Background(), t, tx)

			clearIntegrationSubscriptions(ctx, t, tx)
			for _, fixture := range tt.fixtures {
				insertIntegrationSubscription(ctx, t, tx, fixture)
			}

			queries := sqldb.New(tx)
			repository := NewRepository(queries)
			got, err := repository.CalculateTotalPrice(ctx, tt.filter)
			if err != nil {
				t.Fatalf("CalculateTotalPrice() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("CalculateTotalPrice() = %d, want %d", got, tt.want)
			}
		})
	}
}

type integrationSubscriptionFixture struct {
	ID          uuid.UUID
	ServiceName string
	Price       int32
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

func newIntegrationTestPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration tests")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to TEST_DATABASE_URL: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping TEST_DATABASE_URL: %v", err)
	}

	return pool
}

func requireSubscriptionsTable(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var exists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.subscriptions') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatalf("check subscriptions table in TEST_DATABASE_URL: %v", err)
	}
	if !exists {
		t.Fatal("subscriptions table does not exist in TEST_DATABASE_URL; apply migrations before running integration tests")
	}
}

func beginIntegrationTestTx(ctx context.Context, t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}

	return tx
}

func clearIntegrationSubscriptions(ctx context.Context, t *testing.T, tx pgx.Tx) {
	t.Helper()

	if _, err := tx.Exec(ctx, `DELETE FROM subscriptions`); err != nil {
		t.Fatalf("clear subscriptions inside test transaction: %v", err)
	}
}

func rollbackIntegrationTestTx(ctx context.Context, t *testing.T, tx pgx.Tx) {
	t.Helper()

	if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
		t.Fatalf("rollback transaction: %v", err)
	}
}

func insertIntegrationSubscription(ctx context.Context, t *testing.T, tx pgx.Tx, fixture integrationSubscriptionFixture) {
	t.Helper()

	_, err := tx.Exec(ctx, `
		INSERT INTO subscriptions (
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`,
		fixture.ID,
		fixture.ServiceName,
		fixture.Price,
		fixture.UserID,
		pgDate(fixture.StartDate),
		nullablePGDate(fixture.EndDate),
	)
	if err != nil {
		t.Fatalf("insert integration subscription fixture: %v", err)
	}
}

func mustAPIMonth(t *testing.T, value string) time.Time {
	t.Helper()

	month, err := parseAPIMonth(value)
	if err != nil {
		t.Fatalf("parse test month %q: %v", value, err)
	}

	return month
}

func monthPtr(t *testing.T, value string) *time.Time {
	t.Helper()

	month := mustAPIMonth(t, value)
	return &month
}

func stringPtr(value string) *string {
	return &value
}

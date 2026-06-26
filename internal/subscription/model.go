package subscription

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("subscription not found")

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int32
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpsertSubscription struct {
	ServiceName string
	Price       int32
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type ListSubscriptionsFilter struct {
	Limit       int32
	Offset      int32
	UserID      *uuid.UUID
	ServiceName *string
}

package subscription

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input UpsertSubscription) (Subscription, error) {
	return s.repository.Create(ctx, input)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Subscription, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, error) {
	return s.repository.List(ctx, filter)
}

func (s *Service) CalculateTotalPrice(ctx context.Context, filter TotalPriceFilter) (int64, error) {
	return s.repository.CalculateTotalPrice(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpsertSubscription) (Subscription, error) {
	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

package bridge

import (
	"context"
	"errors"
	"example.com/bridge-health-monitor-service/domain"
	"example.com/bridge-health-monitor-service/store"
	"example.com/bridge-health-monitor-service/validation"
	"fmt"
	"sort"
)

var ErrBridgeNotFound = errors.New("bridge not found")

type Service struct {
	store *store.Store
}

func New(s *store.Store) *Service { return &Service{store: s} }

func (svc *Service) List(ctx context.Context) []domain.Bridge {
	items := svc.store.List()
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (svc *Service) Get(ctx context.Context, id string) (domain.Bridge, error) {
	b, err := svc.store.Get(id)
	if err != nil {
		return domain.Bridge{}, fmt.Errorf("get bridge %s: %w", id, store.ErrNotFound)
	}
	return b, nil
}

func (svc *Service) UpdateCondition(ctx context.Context, id, condition string, expected int) (domain.Bridge, error) {
	b, err := svc.store.Get(id)
	if err != nil {
		return domain.Bridge{}, fmt.Errorf("update bridge %s: %w", id, store.ErrNotFound)
	}
	if err := validation.Transition(b.Condition, condition); err != nil {
		return domain.Bridge{}, err
	}
	return svc.store.UpdateCondition(id, condition, expected)
}

// RiskLevel buckets a bridge risk score into a level name.
func (svc *Service) RiskLevel(risk int) string {
	switch {
	case risk >= 80:
		return "high"
	case risk >= 50:
		return "medium"
	default:
		return "low"
	}
}

func (svc *Service) History(ctx context.Context, id string) ([]string, error) {
	if _, err := svc.store.Get(id); err != nil {
		return nil, fmt.Errorf("history bridge %s: %w", id, store.ErrNotFound)
	}
	return svc.store.History(id), nil
}

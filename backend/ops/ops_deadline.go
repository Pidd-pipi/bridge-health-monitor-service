package ops

import (
	"context"
	"sort"
	"time"
)

// PriorityWindow returns how long a record at the given priority may stay
// open before it is considered overdue.
func PriorityWindow(p OpsPriority) time.Duration {
	switch p {
	case OpsPriorityCritical:
		return 24 * time.Hour
	case OpsPriorityHigh:
		return 72 * time.Hour
	case OpsPriorityNormal:
		return 168 * time.Hour
	default:
		return 336 * time.Hour
	}
}

func (s *OpsService) Deadline(ctx context.Context, id string) (time.Time, error) {
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return time.Time{}, err
	}
	updated, err := opsParseStamp(record.UpdatedAt)
	if err != nil {
		return time.Time{}, wrapOps("deadline", "parse.updated_at", err)
	}
	return updated.Add(PriorityWindow(record.Priority)), nil
}

func (s *OpsService) Overdue(ctx context.Context) ([]OpsRecord, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	out := make([]OpsRecord, 0)
	for _, item := range items {
		if item.Status == OpsStatusClosed || item.Status == OpsStatusQueued {
			continue
		}
		due, err := s.Deadline(ctx, item.ID)
		if err != nil {
			continue
		}
		if now.After(due) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

package ops

import "context"

type BatchItemResult struct {
	ID  string `json:"id"`
	OK  bool   `json:"ok"`
	Err error  `json:"error,omitempty"`
}

// batchLease models a fixed pool of per-item resources (e.g. worker slots)
// that must be released as soon as an item finishes processing.
type batchLease struct {
	slots chan struct{}
}

func newBatchLease(limit int) *batchLease {
	if limit < 1 {
		limit = 1
	}
	return &batchLease{slots: make(chan struct{}, limit)}
}

func (l *batchLease) acquire(ctx context.Context) error {
	select {
	case l.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *batchLease) release() {
	select {
	case <-l.slots:
	default:
	}
}

// BatchClose closes several records. Each item is processed independently: a
// failure on one item must not abort the rest, and every acquired lease must
// be released on all paths.
func (s *OpsService) BatchClose(ctx context.Context, ids []string, actor string) []BatchItemResult {
	results := make([]BatchItemResult, 0, len(ids))
	lease := newBatchLease(2)
	for _, id := range ids {
		if err := lease.acquire(ctx); err != nil {
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		rec, err := s.store.Get(ctx, id)
		if err != nil {
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		rec.Status = OpsStatusClosed
		if err := s.store.Update(ctx, rec, rec.Revision); err != nil {
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		s.audit.Add(rec.ID, "status_changed", actor)
		lease.release()
		results = append(results, BatchItemResult{ID: id, OK: true})
	}
	return results
}

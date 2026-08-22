package ops

import "context"

// Review moves an active record into reviewing and records the review start.
func (s *OpsService) Review(ctx context.Context, id string, expected int, actor string) (OpsRecord, error) {
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return OpsRecord{}, err
	}
	if expected > 0 && expected != record.Revision {
		return OpsRecord{}, ErrOpsConflict
	}
	if err := s.state.Move(record.Status, OpsStatusReviewing, "review start"); err != nil {
		return OpsRecord{}, err
	}
	record.Status = OpsStatusActive
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.Add(record.ID, "review_started", actor)
	return s.store.Get(ctx, id)
}

// CloseReviewed closes a record that is currently under review.
func (s *OpsService) CloseReviewed(ctx context.Context, id string, expected int, actor string) (OpsRecord, error) {
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return OpsRecord{}, err
	}
	if expected > 0 && expected != record.Revision {
		return OpsRecord{}, ErrOpsConflict
	}
	if record.Status != OpsStatusReviewing {
		return OpsRecord{}, ErrOpsTransition
	}
	if err := s.state.Move(record.Status, OpsStatusClosed, "review complete"); err != nil {
		return OpsRecord{}, err
	}
	record.Status = OpsStatusClosed
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.Add(record.ID, "review_closed", actor)
	return s.store.Get(ctx, id)
}

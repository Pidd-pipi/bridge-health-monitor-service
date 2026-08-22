package ops

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type OpsPolicy struct {
	RequireOwner  bool
	RequiredLabel string
	MaxActive     int
}

type OpsService struct {
	store  *OpsStore
	audit  *OpsAudit
	state  *OpsStateMachine
	policy OpsPolicy
	clock  OpsClock
}

func NewService(seed []OpsRecord) *OpsService {
	return &OpsService{store: newOpsStore(seed), audit: newOpsAudit(), state: newOpsStateMachine(), policy: OpsPolicy{RequireOwner: true, RequiredLabel: "site", MaxActive: 1000}, clock: newOpsClock()}
}

// SeedRecords returns the initial monitoring-event ledger.
func SeedRecords() []OpsRecord { return seedOpsRecords() }

func seedOpsRecords() []OpsRecord {
	base := "2026-08-20T08:00:00Z"
	return []OpsRecord{
		{
			ID: "evt-1001", Subject: "主梁裂缝宽度超限", Owner: "张工", Status: OpsStatusActive,
			Priority: OpsPriorityCritical, Revision: 3, Labels: map[string]string{"site": "东江大桥", "operator": "张工", "evidence": "img-001", "reviewed": "yes"},
			CreatedAt: base, UpdatedAt: "2026-08-20T08:00:00Z",
		},
		{
			ID: "evt-1002", Subject: "桥墩沉降监测点漂移", Owner: "李工", Status: OpsStatusQueued,
			Priority: OpsPriorityHigh, Revision: 1, Labels: map[string]string{"site": "港湾跨线桥", "operator": "李工", "evidence": "csv-002"},
			CreatedAt: base, UpdatedAt: "2026-08-20T08:00:00Z",
		},
		{
			ID: "evt-1003", Subject: "伸缩缝异响排查", Owner: "王工", Status: OpsStatusPaused,
			Priority: OpsPriorityNormal, Revision: 2, Labels: map[string]string{"site": "东江大桥", "operator": "王工", "evidence": "rec-003"},
			CreatedAt: base, UpdatedAt: "2026-08-20T08:00:00Z",
		},
	}
}

func (p OpsPolicy) Check(record OpsRecord) error {
	if p.RequireOwner && strings.TrimSpace(record.Owner) == "" {
		return fmt.Errorf("%w: owner required", ErrOpsPolicy)
	}
	if p.RequiredLabel != "" && record.LabelValue(p.RequiredLabel) == "" {
		return fmt.Errorf("%w: %s label required", ErrOpsPolicy, p.RequiredLabel)
	}
	if record.Priority == "" {
		return fmt.Errorf("%w: priority required", ErrOpsPolicy)
	}
	return nil
}

func (s *OpsService) Create(ctx context.Context, record OpsRecord) (OpsRecord, error) {
	record = normalizeOpsRecord(record)
	if err := s.policy.Check(record); err != nil {
		return OpsRecord{}, err
	}
	record.Status = OpsStatusQueued
	record.Revision = 1
	record.CreatedAt = s.clock.Stamp()
	record.UpdatedAt = record.CreatedAt
	if err := s.store.Put(ctx, record); err != nil {
		return OpsRecord{}, wrapOps("create", "store.put", err)
	}
	s.audit.Add(record.ID, "created", record.Owner)
	return record, nil
}

func (s *OpsService) Get(ctx context.Context, id string) (OpsRecord, error) {
	return s.store.Get(ctx, id)
}

func (s *OpsService) Search(ctx context.Context, q OpsQuery) (OpsPage, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return OpsPage{}, err
	}
	filtered := FilterRecords(items, q)
	sortOpsRecords(filtered)
	q = opsQueryDefaults(q)
	pageItems, total := Paginate(filtered, q.Page, q.PageSize)
	_, end := opsBounds(len(filtered), q.Page, q.PageSize)
	return OpsPage{Items: pageItems, Page: q.Page, PageSize: q.PageSize, Total: total, HasNext: end < len(filtered)}, nil
}

func (s *OpsService) Transition(ctx context.Context, id string, expected int, target OpsStatus, actor string) (OpsRecord, error) {
	ctx, cancel := opsContext(ctx, 3*time.Second)
	defer cancel()
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return OpsRecord{}, err
	}
	if expected > 0 && expected != record.Revision {
		return OpsRecord{}, ErrOpsConflict
	}
	if err := s.state.Move(record.Status, target, "operator update"); err != nil {
		return OpsRecord{}, err
	}
	if target == OpsStatusClosed {
		if err := RequireReview(record); err != nil {
			return OpsRecord{}, err
		}
	}
	record.Status = target
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.Add(record.ID, "status_changed", actor)
	return s.store.Get(ctx, id)
}

func (s *OpsService) Audit(id string) []OpsEvent { return s.audit.For(id) }

func (s *OpsService) Snapshot() OpsSnapshot {
	items, _ := s.store.List(context.Background())
	out := OpsSnapshot{
		Domain:      opsDomainName,
		GeneratedAt: s.clock.Stamp(),
		ByStatus:    map[OpsStatus]int{},
		ByPriority:  map[OpsPriority]int{},
	}
	for _, i := range items {
		out.Records++
		out.ByStatus[i.Status]++
		out.ByPriority[i.Priority]++
		if i.Status == OpsStatusActive {
			out.Active++
		}
	}
	return out
}

func (s *OpsService) Domain() string { return opsDomainName }
func (s *OpsService) Count() int     { return s.store.Count() }

func timeNowOps() string { return time.Now().UTC().Format(time.RFC3339Nano) }

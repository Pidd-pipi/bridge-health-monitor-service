#!/usr/bin/env python3
"""Apply the 10 bug transformations to each record's env/ (gold = baseline)."""
import sys

ROOT = "/Users/yu/work/bridge-health-monitor-batch-20260822-s01a02862"
DATE = "2026-08-22"


def env(rec):
    return f"{ROOT}/{DATE}/bridge-health-monitor-service__{rec}/env/backend/"


def apply(path, old, new):
    with open(path) as f:
        s = f.read()
    if old not in s:
        raise SystemExit(f"PATTERN NOT FOUND in {path}:\n{old[:200]}")
    if s.count(old) != 1:
        raise SystemExit(f"PATTERN AMBIGUOUS ({s.count(old)}x) in {path}:\n{old[:200]}")
    s = s.replace(old, new)
    with open(path, "w") as f:
        f.write(s)


# ---------------- 001 concurrency ----------------
b = env("001")
apply(b + "ops/ops_store.go", '''	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return OpsRecord{}, fmt.Errorf("ops store get %s: %w", id, ErrOpsNotFound)
	}
	return item.Clone(), nil
}''', '''	item, ok := s.items[id]
	if !ok {
		return OpsRecord{}, fmt.Errorf("ops store get %s: %w", id, ErrOpsNotFound)
	}
	return item, nil
}''')
apply(b + "ops/ops_store.go", '''	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OpsRecord, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}''', '''	out := make([]OpsRecord, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}''')
apply(b + "ops/ops_store.go", '''	item.Revision = current.Revision + 1
	item.UpdatedAt = timeNowOps()
	s.items[item.ID] = item.Clone()
	return nil
}''', '''	item.Revision = current.Revision + 1
	item.UpdatedAt = timeNowOps()
	s.items[item.ID] = item
	return nil
}''')
apply(b + "ops/ops_service.go", '''func (s *OpsService) Snapshot() OpsSnapshot {
	items, _ := s.store.List(context.Background())
	out := OpsSnapshot{Domain: opsDomainName, GeneratedAt: s.clock.Stamp(), ByStatus: map[OpsStatus]int{}, ByPriority: map[OpsPriority]int{}}
	for _, i := range items {
		out.Records++
		out.ByStatus[i.Status]++
		out.ByPriority[i.Priority]++
		if i.Status == OpsStatusActive {
			out.Active++
		}
	}
	return out
}''', '''var snapshotStatusCounts = map[OpsStatus]int{}
var snapshotPriorityCounts = map[OpsPriority]int{}

func (s *OpsService) Snapshot() OpsSnapshot {
	items, _ := s.store.List(context.Background())
	out := OpsSnapshot{Domain: opsDomainName, GeneratedAt: s.clock.Stamp(), ByStatus: snapshotStatusCounts, ByPriority: snapshotPriorityCounts}
	for k := range snapshotStatusCounts {
		delete(snapshotStatusCounts, k)
	}
	for k := range snapshotPriorityCounts {
		delete(snapshotPriorityCounts, k)
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
}''')
print("001 ok")

# ---------------- 002 slice ----------------
b = env("002")
apply(b + "ops/ops_filter.go", '''func FilterRecords(items []OpsRecord, q OpsQuery) []OpsRecord {
	out := make([]OpsRecord, 0, len(items))
	for _, item := range items {
		if opsMatch(item, q) && opsMatchDate(item, q.From, q.To) {
			out = append(out, item)
		}
	}
	return out
}''', '''func FilterRecords(items []OpsRecord, q OpsQuery) []OpsRecord {
	out := items[:0]
	for _, item := range items {
		if opsMatch(item, q) && opsMatchDate(item, q.From, q.To) {
			out = append(out, item)
		}
	}
	return out
}''')
apply(b + "ops/ops_filter.go", '''	q := opsQueryDefaults(OpsQuery{Page: page, PageSize: pageSize})
	start, end := opsBounds(len(items), q.Page, q.PageSize)
	pageItems := make([]OpsRecord, end-start)
	copy(pageItems, items[start:end])
	return pageItems, len(items)
}''', '''	q := opsQueryDefaults(OpsQuery{Page: page, PageSize: pageSize})
	start, end := opsBounds(len(items), q.Page, q.PageSize)
	pageItems := items[start:end]
	return pageItems, len(items)
}''')
apply(b + "ops/ops_filter.go", '''func MergeLabels(base map[string]string, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}''', '''func MergeLabels(base map[string]string, extra map[string]string) map[string]string {
	for k, v := range extra {
		base[k] = v
	}
	return base
}''')
apply(b + "ops/ops_query.go", '''func opsClonePage(p OpsPage) OpsPage {
	p.Items = append([]OpsRecord(nil), p.Items...)
	return p
}''', '''func opsClonePage(p OpsPage) OpsPage {
	return p
}''')
print("002 ok")

# ---------------- 003 error ----------------
b = env("003")
apply(b + "ops/ops_policy.go", '''	record, err := s.store.Get(ctx, id)
	if err != nil {
		return PolicyResult{}, err
	}''', '''	record, err := s.store.Get(ctx, id)
	if err != nil {
		return PolicyResult{}, fmt.Errorf("policy: %v", err)
	}''')
apply(b + "ops/ops_policy.go", '''func RequireReview(record OpsRecord) error {
	if record.Priority == OpsPriorityCritical && record.Status != OpsStatusClosed && record.LabelValue("reviewed") == "" {
		return wrapOps("policy", "review required", ErrOpsPolicy)
	}
	return nil
}''', '''func RequireReview(record OpsRecord) error {
	return nil
}''')
apply(b + "ops/ops_errors.go", '''	switch {
	case errors.Is(err, ErrOpsNotFound):
		return "not_found"
	case errors.Is(err, ErrOpsConflict):
		return "conflict"
	case errors.Is(err, ErrOpsInvalid):
		return "invalid"
	case errors.Is(err, ErrOpsTransition):
		return "transition"
	case errors.Is(err, ErrOpsPolicy):
		return "policy"
	default:
		return "internal"
	}''', '''	switch {
	case err == ErrOpsNotFound:
		return "not_found"
	case err == ErrOpsConflict:
		return "conflict"
	case err == ErrOpsInvalid:
		return "invalid"
	case err == ErrOpsTransition:
		return "transition"
	case err == ErrOpsPolicy:
		return "policy"
	default:
		return "internal"
	}''')
apply(b + "ops/ops_http.go", '''	case "policy":
		return http.StatusForbidden''', '''	case "policy":
		return http.StatusInternalServerError''')
print("003 ok")

# ---------------- 004 nil ----------------
b = env("004")
apply(b + "ops/ops_clock.go", '''func (c OpsClock) Now() time.Time {
	if c.NowFunc == nil {
		return time.Now().UTC()
	}
	return c.NowFunc().UTC()
}''', '''func (c OpsClock) Now() time.Time {
	return c.NowFunc().UTC()
}''')
apply(b + "ops/ops_deadline.go", '''	updated, err := opsParseStamp(record.UpdatedAt)
	if err != nil {
		return time.Time{}, wrapOps("deadline", "parse.updated_at", err)
	}
	return updated.Add(PriorityWindow(record.Priority)), nil''', '''	updated, err := opsParseStamp(record.UpdatedAt)
	if err != nil {
		return time.Time{}, nil
	}
	return updated.Add(PriorityWindow(record.Priority)), nil''')
apply(b + "ops/ops_deadline.go", '''		due, err := s.Deadline(ctx, item.ID)
		if err != nil {
			continue
		}
		if now.After(due) {
			out = append(out, item)
		}''', '''		due, _ := s.Deadline(ctx, item.ID)
		if now.After(due) {
			out = append(out, item)
		}''')
apply(b + "ops/ops_report.go", '''	report := OpsReport{
		GeneratedAt: s.clock.Stamp(),
		ByStatus:    map[OpsStatus]int{},
		ByPriority:  map[OpsPriority]int{},
	}''', '''	report := OpsReport{
		GeneratedAt: s.clock.Stamp(),
	}
	if len(items) > 0 {
		report.ByStatus = map[OpsStatus]int{}
		report.ByPriority = map[OpsPriority]int{}
	}''')
print("004 ok")

# ---------------- 005 context ----------------
b = env("005")
apply(b + "runtime.go", '''		Handler:           ops.EnterpriseMiddleware(requestIDMiddleware(recoveryMiddleware(deadlineMiddleware(handler)))),''', '''		Handler:           ops.EnterpriseMiddleware(requestIDMiddleware(recoveryMiddleware(handler))),''')
apply(b + "httpapi/handler.go", '''type Handler struct {
	bridges *bridge.Service
	events  *ops.OpsService
}

func New(b *bridge.Service, e *ops.OpsService) *Handler { return &Handler{bridges: b, events: e} }''', '''type Handler struct {
	bridges *bridge.Service
	events  *ops.OpsService
	baseCtx context.Context
}

func New(b *bridge.Service, e *ops.OpsService) *Handler { return &Handler{bridges: b, events: e, baseCtx: context.Background()} }''')
apply(b + "httpapi/handler.go", '''func requestCtx(r *http.Request) context.Context { return r.Context() }''', '''func (h *Handler) requestCtx(r *http.Request) context.Context { return h.baseCtx }''')
# update call sites requestCtx(r) -> h.requestCtx(r) in env handler
import re
p = b + "httpapi/handler.go"
s = open(p).read()
s2 = re.sub(r"requestCtx\(r\)", "h.requestCtx(r)", s)
open(p, "w").write(s2)
apply(b + "httpapi/handler.go", '''func writeOpsError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		http.Error(w, http.StatusText(http.StatusGatewayTimeout), 504)
		return
	}
	http.Error(w, err.Error(), ops.StatusForError(err))
}''', '''func writeOpsError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), ops.StatusForError(err))
}''')
print("005 ok")

# ---------------- 006 defer ----------------
b = env("006")
apply(b + "ops/ops_audit.go", '''func (a *OpsAudit) Since(start time.Time) []OpsEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()''', '''func (a *OpsAudit) Since(start time.Time) []OpsEvent {
	a.mu.Lock()
	defer a.mu.RUnlock()''')
apply(b + "ops/ops_audit.go", '''func (a *OpsAudit) Count() int { a.mu.RLock(); defer a.mu.RUnlock(); return len(a.events) }''', '''func (a *OpsAudit) Count() int { a.mu.Lock(); defer a.mu.RUnlock(); return len(a.events) }''')
apply(b + "ops/ops_batch.go", '''		rec, err := s.store.Get(ctx, id)
		if err != nil {
			lease.release()
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		rec.Status = OpsStatusClosed
		if err := s.store.Update(ctx, rec, rec.Revision); err != nil {
			lease.release()
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		s.audit.Add(rec.ID, "batch_closed", actor)
		lease.release()
		results = append(results, BatchItemResult{ID: id, OK: true})''', '''		rec, err := s.store.Get(ctx, id)
		if err != nil {
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		rec.Status = OpsStatusClosed
		if err := s.store.Update(ctx, rec, rec.Revision); err != nil {
			results = append(results, BatchItemResult{ID: id, Err: err})
			continue
		}
		lease.release()
		results = append(results, BatchItemResult{ID: id, OK: true})''')
print("006 ok")

# ---------------- 007 other P8 ----------------
b = env("007")
apply(b + "ops/ops_model.go", '''func opsStatusValid(value OpsStatus) bool {
	return value == OpsStatusQueued || value == OpsStatusActive || value == OpsStatusReviewing ||
		value == OpsStatusPaused || value == OpsStatusClosed
}''', '''func opsStatusValid(value OpsStatus) bool {
	return value == OpsStatusQueued || value == OpsStatusActive ||
		value == OpsStatusPaused || value == OpsStatusClosed
}''')
apply(b + "ops/ops_model.go", '''func (r OpsRecord) Terminal() bool               { return r.Status == OpsStatusClosed }''', '''func (r OpsRecord) Terminal() bool {
	return r.Status == OpsStatusClosed || r.Status == OpsStatusReviewing
}''')
apply(b + "ops/ops_state.go", '''var opsTransitionTable = map[OpsStatus]map[OpsStatus]bool{
	OpsStatusQueued:    {OpsStatusActive: true, OpsStatusClosed: true},
	OpsStatusActive:    {OpsStatusPaused: true, OpsStatusClosed: true, OpsStatusReviewing: true},
	OpsStatusReviewing: {OpsStatusActive: true, OpsStatusClosed: true},
	OpsStatusPaused:    {OpsStatusActive: true, OpsStatusClosed: true},
	OpsStatusClosed:    {},
}''', '''var opsTransitionTable = map[OpsStatus]map[OpsStatus]bool{
	OpsStatusQueued:    {OpsStatusActive: true, OpsStatusClosed: true},
	OpsStatusActive:    {OpsStatusPaused: true, OpsStatusClosed: true},
	OpsStatusReviewing: {OpsStatusActive: true},
	OpsStatusPaused:    {OpsStatusActive: true, OpsStatusClosed: true},
	OpsStatusClosed:    {},
}''')
apply(b + "ops/ops_review.go", '''	record.Status = OpsStatusReviewing
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.Add(record.ID, "review_started", actor)''', '''	record.Status = OpsStatusActive
	if err := s.store.Update(ctx, record, expected); err != nil {
		return OpsRecord{}, err
	}
	s.audit.Add(record.ID, "review_started", actor)''')
print("007 ok")

# ---------------- 008 other P8 ----------------
b = env("008")
apply(b + "domain/models.go", '''var conditionTransitions = map[string]map[string]bool{
	"monitored":  {"watch": true, "restricted": true},
	"watch":      {"restricted": true, "cleared": true},
	"restricted": {"cleared": true, "watch": true},
	"cleared":    {"watch": true},
}''', '''var conditionTransitions = map[string]map[string]bool{
	"monitored":  {"restricted": true},
	"watch":      {"restricted": true},
	"restricted": {"watch": true, "monitored": true},
	"cleared":    {"watch": true},
}''')
apply(b + "validation/validation.go", '''func Transition(current, next string) error {
	if err := Condition(next); err != nil {
		return err
	}
	if !domain.ConditionTransition(current, next) {
		return fmt.Errorf("condition transition %q -> %q is not allowed", current, next)
	}
	return nil
}''', '''func Transition(current, next string) error {
	if err := Condition(next); err != nil {
		return err
	}
	if current == next {
		return fmt.Errorf("condition %q already set", next)
	}
	if !domain.ConditionTransition(current, next) {
		return fmt.Errorf("condition transition %q -> %q is not allowed", current, next)
	}
	return nil
}''')
print("008 ok")

# ---------------- 009 error P3 ----------------
b = env("009")
apply(b + "bridge/report.go", '''		b, err := svc.store.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				r.Missing++
				continue
			}
			return Summary{}, fmt.Errorf("bridge report: %w", err)
		}''', '''		b, err := svc.store.Get(id)
		if err != nil {
			if err == store.ErrNotFound {
				r.Missing++
				continue
			}
			return Summary{}, fmt.Errorf("bridge report: %w", err)
		}''')
apply(b + "bridge/report.go", '''	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true''', '''	for _, id := range ids {''')
apply(b + "bridge/export.go", '''		br, err := svc.store.Get(id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				b.WriteString(csvBridgeRow(id, "", "", "missing", 0, 0))
				continue
			}
			return "", fmt.Errorf("bridge export: %w", err)
		}''', '''		br, err := svc.store.Get(id)
		if err != nil {
			if err == store.ErrNotFound {
				continue
			}
			return "", fmt.Errorf("bridge export: %w", err)
		}''')
print("009 ok")

# ---------------- 010 nil P4 ----------------
b = env("010")
apply(b + "store/store.go", '''	b, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, ErrNotFound
	}
	return b, nil
}''', '''	b, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, nil
	}
	return b, nil
}''')
apply(b + "store/store.go", '''	current, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, ErrNotFound
	}
	if expected > 0 && current.Revision != expected {''', '''	current, ok := s.items[id]
	if !ok {
		return domain.Bridge{}, nil
	}
	if expected > 0 && current.Revision != expected {''')
apply(b + "bridge/bridge.go", '''func (svc *Service) History(ctx context.Context, id string) ([]string, error) {
	if _, err := svc.store.Get(id); err != nil {
		return nil, fmt.Errorf("history bridge %s: %w", id, store.ErrNotFound)
	}
	return svc.store.History(id), nil
}''', '''func (svc *Service) History(ctx context.Context, id string) ([]string, error) {
	return svc.store.History(id), nil
}''')
apply(b + "bridge/bridge.go", '''func (svc *Service) RiskLevel(risk int) string {
	switch {
	case risk >= 80:
		return "high"
	case risk >= 50:
		return "medium"
	default:
		return "low"
	}
}''', '''func (svc *Service) RiskLevel(risk int) string {
	levels := []string{"low", "medium", "high"}
	return levels[risk/40]
}''')
print("010 ok")

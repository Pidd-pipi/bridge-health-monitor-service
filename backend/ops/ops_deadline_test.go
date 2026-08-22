package ops

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPriorityWindowDefaults(t *testing.T) {
	if got := PriorityWindow(OpsPriorityCritical); got != 24*time.Hour {
		t.Fatalf("critical window=%v", got)
	}
	if got := PriorityWindow(OpsPriorityLow); got != 336*time.Hour {
		t.Fatalf("low window=%v want 336h", got)
	}
}

func TestDeadlineBasedOnPriority(t *testing.T) {
	svc := serviceWithSeed(t)
	due, err := svc.Deadline(context.Background(), "evt-1001")
	if err != nil {
		t.Fatalf("deadline: %v", err)
	}
	base := time.Date(2026, 8, 20, 8, 0, 0, 0, time.UTC)
	if !due.Equal(base.Add(24 * time.Hour)) {
		t.Fatalf("deadline=%v want %v", due, base.Add(24*time.Hour))
	}
}

func TestDeadlineUnknownNotFound(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.Deadline(context.Background(), "nope"); !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("deadline unknown: err=%v want not found", err)
	}
}

func TestOverdueIncludesOnlyOpen(t *testing.T) {
	svc := serviceWithSeed(t)
	items, err := svc.Overdue(context.Background())
	if err != nil {
		t.Fatalf("overdue: %v", err)
	}
	// evt-1001 active critical updated 2026-08-20 -> overdue with real clock
	for _, item := range items {
		if item.Status == OpsStatusClosed || item.Status == OpsStatusQueued {
			t.Fatalf("overdue includes closed/queued item %s", item.ID)
		}
	}
}

func TestDeadlineInvalidStampErrors(t *testing.T) {
	svc := serviceWithSeed(t)
	rec := svc.store.items["evt-1003"]
	rec.UpdatedAt = "not-a-time"
	svc.store.items["evt-1003"] = rec
	if _, err := svc.Deadline(context.Background(), "evt-1003"); err == nil {
		t.Fatal("deadline with invalid stamp should error")
	}
}

func TestOverdueSkipsUnparseableStamp(t *testing.T) {
	svc := serviceWithSeed(t)
	rec := svc.store.items["evt-1003"]
	rec.UpdatedAt = "not-a-time"
	svc.store.items["evt-1003"] = rec
	items, err := svc.Overdue(context.Background())
	if err != nil {
		t.Fatalf("overdue: %v", err)
	}
	for _, item := range items {
		if item.ID == "evt-1003" {
			t.Fatalf("overdue should skip unparseable item evt-1003")
		}
	}
}

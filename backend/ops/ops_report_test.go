package ops

import (
	"context"
	"testing"
)

func TestReportAggregates(t *testing.T) {
	svc := serviceWithSeed(t)
	report, err := svc.Report(context.Background(), 7)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if report.Total != 3 {
		t.Fatalf("total=%d want 3", report.Total)
	}
	if report.ByStatus[OpsStatusActive] != 1 || report.ByPriority[OpsPriorityCritical] != 1 {
		t.Fatalf("buckets wrong: %+v", report)
	}
}

func TestReportDoesNotMutateEvents(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.Report(context.Background(), 7); err != nil {
		t.Fatalf("report: %v", err)
	}
	rec, _ := svc.Get(context.Background(), "evt-1001")
	if _, ok := rec.Labels["reported"]; ok {
		t.Fatal("report mutated event labels")
	}
}

func TestReportEmptyStoreMapsInitialized(t *testing.T) {
	svc := NewService(nil)
	report, err := svc.Report(context.Background(), 7)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if report.ByStatus == nil || report.ByPriority == nil {
		t.Fatal("empty report maps must not be nil")
	}
	if report.Total != 0 {
		t.Fatalf("total=%d want 0", report.Total)
	}
}

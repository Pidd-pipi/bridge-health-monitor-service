package bridge

import (
	"context"
	"example.com/bridge-health-monitor-service/store"
	"testing"
)

func TestBridgeReportMissingIdsCounted(t *testing.T) {
	svc := New(store.New())
	r, err := svc.Report(context.Background(), []string{"br-201", "nope"})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if r.Total != 1 || r.Missing != 1 {
		t.Fatalf("report total=%d missing=%d want 1/1", r.Total, r.Missing)
	}
}

func TestBridgeReportDeduplicatesIds(t *testing.T) {
	svc := New(store.New())
	r, err := svc.Report(context.Background(), []string{"br-201", "br-201"})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if r.Total != 1 {
		t.Fatalf("report total=%d want 1 (duplicates counted)", r.Total)
	}
}

func TestBridgeReportByRiskLevels(t *testing.T) {
	svc := New(store.New())
	r, err := svc.Report(context.Background(), []string{"br-201", "br-202"})
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if r.ByRisk["low"] != 1 || r.ByRisk["medium"] != 1 {
		t.Fatalf("risk buckets wrong: %+v", r.ByRisk)
	}
}

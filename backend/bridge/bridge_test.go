package bridge

import (
	"context"
	"example.com/bridge-health-monitor-service/store"
	"testing"
)

func newService(t *testing.T) *Service {
	t.Helper()
	return New(store.New())
}

func TestBridgeGetExisting(t *testing.T) {
	svc := newService(t)
	b, err := svc.Get(context.Background(), "br-201")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if b.ID != "br-201" || b.Condition != "monitored" {
		t.Fatalf("bridge=%+v", b)
	}
}

func TestFetchMissingBridgeRejects(t *testing.T) {
	svc := newService(t)
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("missing bridge get should return error")
	}
}

func TestBridgeUpdateMissingReturnsNotFound(t *testing.T) {
	svc := newService(t)
	if _, err := svc.UpdateCondition(context.Background(), "nope", "cleared", 0); err == nil {
		t.Fatal("missing bridge update should return error")
	}
}

func TestBridgeRiskLevelBoundariesSafe(t *testing.T) {
	svc := newService(t)
	if got := svc.RiskLevel(0); got != "low" {
		t.Fatalf("risk 0 level=%s want low", got)
	}
	if got := svc.RiskLevel(120); got != "high" {
		t.Fatalf("risk 120 level=%s want high", got)
	}
	if got := svc.RiskLevel(50); got != "medium" {
		t.Fatalf("risk 50 level=%s want medium", got)
	}
}

func TestBridgeHistoryNoPanic(t *testing.T) {
	svc := newService(t)
	if _, err := svc.UpdateCondition(context.Background(), "br-201", "watch", 1); err != nil {
		t.Fatalf("update: %v", err)
	}
	history, err := svc.History(context.Background(), "br-201")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) != 1 || history[0] != "watch" {
		t.Fatalf("history=%v", history)
	}
}

func TestBridgeHistoryMissing(t *testing.T) {
	svc := newService(t)
	if _, err := svc.History(context.Background(), "nope"); err == nil {
		t.Fatal("history missing bridge should error")
	}
}

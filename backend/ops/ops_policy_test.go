package ops

import (
	"context"
	"errors"
	"testing"
)

func TestEventPolicySatisfiedCritical(t *testing.T) {
	svc := serviceWithSeed(t)
	result, err := svc.EvaluatePolicy(context.Background(), "evt-1001")
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if !result.Satisfied {
		t.Fatalf("evt-1001 should satisfy policy: %+v", result)
	}
	if result.Code == "" || result.Severity != OpsPriorityCritical {
		t.Fatalf("policy rule wrong: %+v", result)
	}
}

func TestEventPolicyUnknownNotFound(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.EvaluatePolicy(context.Background(), "nope"); !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("policy unknown: err=%v want not found", err)
	}
}

func TestEventPolicyMissingLabelsUnsatisfied(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, err := svc.Create(context.Background(), OpsRecord{Subject: "钢索腐蚀", Owner: "孙工", Priority: OpsPriorityCritical, Labels: map[string]string{"site": "东江大桥", "operator": "孙工", "evidence": "sd-1"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	result, err := svc.EvaluatePolicy(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if result.Satisfied {
		t.Fatalf("missing reviewed label should be unsatisfied: %+v", result)
	}
}

func TestEventPolicyRequireReviewForCriticalClose(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, err := svc.Create(context.Background(), OpsRecord{Subject: "桥面开裂", Owner: "周工", Priority: OpsPriorityCritical, Labels: map[string]string{"site": "港湾跨线桥", "operator": "周工", "evidence": "ck-2"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := RequireReview(rec); err == nil {
		t.Fatal("critical record without review label should be rejected on close")
	}
	rec.Labels["reviewed"] = "yes"
	if err := RequireReview(rec); err != nil {
		t.Fatalf("reviewed critical should pass: %v", err)
	}
}

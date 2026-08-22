package ops

import (
	"context"
	"testing"
)

func serviceWithSeed(t *testing.T) *OpsService {
	t.Helper()
	return NewService(seedOpsRecords())
}

func TestEventCreateAssignsIdentity(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, err := svc.Create(context.Background(), OpsRecord{ID: "evt-9001", Subject: "桥面渗水", Owner: "吴工", Priority: OpsPriorityHigh, Labels: map[string]string{"site": "东江大桥", "operator": "吴工", "evidence": "e1"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rec.Status != OpsStatusQueued || rec.Revision != 1 {
		t.Fatalf("create identity: status=%s rev=%d", rec.Status, rec.Revision)
	}
	got, err := svc.Get(context.Background(), "evt-9001")
	if err != nil || got.Subject != "桥面渗水" {
		t.Fatalf("get created: %v", err)
	}
}

func TestEventCreateRequiresOwnerAndSite(t *testing.T) {
	svc := serviceWithSeed(t)
	if _, err := svc.Create(context.Background(), OpsRecord{Subject: "x", Priority: OpsPriorityHigh, Labels: map[string]string{"site": "a"}}); err == nil {
		t.Fatal("create without owner should fail")
	}
	if _, err := svc.Create(context.Background(), OpsRecord{Subject: "x", Owner: "o", Priority: OpsPriorityHigh}); err == nil {
		t.Fatal("create without site label should fail")
	}
}

func TestEventSearchPagination(t *testing.T) {
	svc := serviceWithSeed(t)
	page, err := svc.Search(context.Background(), OpsQuery{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if page.Total != 3 || len(page.Items) != 2 || !page.HasNext {
		t.Fatalf("page1: total=%d len=%d hasNext=%v", page.Total, len(page.Items), page.HasNext)
	}
	page2, _ := svc.Search(context.Background(), OpsQuery{Page: 2, PageSize: 2})
	if len(page2.Items) != 1 || page2.HasNext {
		t.Fatalf("page2: len=%d hasNext=%v", len(page2.Items), page2.HasNext)
	}
}

func TestLedgerSnapshotFreshPerCall(t *testing.T) {
	svc := serviceWithSeed(t)
	first := svc.Snapshot()
	first.ByStatus[OpsStatusActive] = 999
	first.ByPriority[OpsPriorityCritical] = 999
	second := svc.Snapshot()
	if second.ByStatus[OpsStatusActive] != 1 {
		t.Fatalf("snapshot active polluted: got %d want 1", second.ByStatus[OpsStatusActive])
	}
	if second.ByPriority[OpsPriorityCritical] != 1 {
		t.Fatalf("snapshot critical polluted: got %d want 1", second.ByPriority[OpsPriorityCritical])
	}
	if second.ByStatus[OpsStatusQueued] != 1 || second.ByStatus[OpsStatusPaused] != 1 {
		t.Fatalf("snapshot buckets wrong: %+v", second.ByStatus)
	}
	if second.Records != 3 {
		t.Fatalf("records=%d want 3", second.Records)
	}
	sum := 0
	for _, v := range second.ByStatus {
		sum += v
	}
	if sum != second.Records {
		t.Fatalf("by_status sum=%d != records=%d", sum, second.Records)
	}
}

func TestEventTransitionWithExpectedRevision(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, err := svc.Transition(context.Background(), "evt-1002", 1, OpsStatusActive, "李工")
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if rec.Status != OpsStatusActive || rec.Revision != 2 {
		t.Fatalf("transition result: status=%s rev=%d", rec.Status, rec.Revision)
	}
}

func TestEventSearchResultIsolated(t *testing.T) {
	svc := serviceWithSeed(t)
	page, _ := svc.Search(context.Background(), OpsQuery{})
	page.Items[0].Labels["hack"] = "1"
	again, _ := svc.Search(context.Background(), OpsQuery{})
	for _, item := range again.Items {
		if item.Labels["hack"] == "1" {
			t.Fatal("search result shares state with store")
		}
	}
}

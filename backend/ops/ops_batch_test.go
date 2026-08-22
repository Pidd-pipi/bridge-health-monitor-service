package ops

import (
	"context"
	"testing"
	"time"
)

func TestBatchCloseAllSucceed(t *testing.T) {
	svc := serviceWithSeed(t)
	results := svc.BatchClose(context.Background(), []string{"evt-1002", "evt-1003"}, "工长")
	if len(results) != 2 {
		t.Fatalf("results=%d want 2", len(results))
	}
	for _, r := range results {
		if !r.OK {
			t.Fatalf("item %s failed: %v", r.ID, r.Err)
		}
	}
	rec, _ := svc.Get(context.Background(), "evt-1002")
	if rec.Status != OpsStatusClosed {
		t.Fatalf("evt-1002 status=%s want closed", rec.Status)
	}
}

func TestBatchClosePartialFailureKeepsWorking(t *testing.T) {
	svc := serviceWithSeed(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	results := svc.BatchClose(ctx, []string{"evt-9999", "evt-9998", "evt-1002", "evt-1003"}, "工长")
	okCount := 0
	for _, r := range results {
		if r.OK {
			okCount++
		}
	}
	if okCount != 2 {
		t.Fatalf("ok=%d want 2 (results=%+v)", okCount, results)
	}
}

func TestBatchCloseLargeBatchNoStarvation(t *testing.T) {
	svc := serviceWithSeed(t)
	// create several queued events
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		id := string(rune('a'+i)) + "batch"
		rec, err := svc.Create(context.Background(), OpsRecord{ID: id, Subject: "批量关闭", Owner: "工长", Priority: OpsPriorityNormal, Labels: map[string]string{"site": "东江大桥", "operator": "工长", "evidence": "b" + string(rune('0'+i))}})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ids = append(ids, rec.ID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	results := svc.BatchClose(ctx, ids, "工长")
	if len(results) != 5 {
		t.Fatalf("results=%d want 5", len(results))
	}
	for _, r := range results {
		if !r.OK {
			t.Fatalf("item %s failed: %v", r.ID, r.Err)
		}
	}
}

func TestBatchCloseRecordsAudit(t *testing.T) {
	svc := serviceWithSeed(t)
	_ = svc.BatchClose(context.Background(), []string{"evt-1002"}, "工长")
	events := svc.Audit("evt-1002")
	found := false
	for _, ev := range events {
		if ev.Type == "batch_closed" {
			found = true
		}
	}
	if !found {
		t.Fatal("batch close did not record audit event")
	}
}

package ops

import (
	"testing"
	"time"
)

func newAudit() *OpsAudit { return newOpsAudit() }

func TestAuditAddForSince(t *testing.T) {
	a := newAudit()
	a.Add("evt-1001", "created", "张工")
	a.Add("evt-1001", "status_changed", "李工")
	a.Add("evt-1002", "created", "王工")
	if got := len(a.For("evt-1001")); got != 2 {
		t.Fatalf("for(evt-1001)=%d want 2", got)
	}
	if got := len(a.Since(time.Now().Add(-time.Hour))); got != 3 {
		t.Fatalf("since=%d want 3", got)
	}
}

func TestAuditSinceNoPanic(t *testing.T) {
	a := newAudit()
	a.Add("evt-1001", "created", "张工")
	a.Add("evt-1001", "status_changed", "李工")
	_ = a.Since(time.Now().Add(-time.Hour))
}

func TestAuditCountNoPanic(t *testing.T) {
	a := newAudit()
	a.Add("evt-1001", "created", "张工")
	a.Add("evt-1002", "created", "王工")
	if got := a.Count(); got != 2 {
		t.Fatalf("count=%d want 2", got)
	}
}

func TestAuditLatestNoPanic(t *testing.T) {
	a := newAudit()
	a.Add("evt-1001", "created", "张工")
	if _, ok := a.Latest(); !ok {
		t.Fatal("latest should return an event")
	}
}

func TestAuditForNoPanic(t *testing.T) {
	a := newAudit()
	a.Add("evt-1001", "created", "张工")
	if got := len(a.For("evt-1001")); got != 1 {
		t.Fatalf("for=%d want 1", got)
	}
}

func TestAuditLatestAndClear(t *testing.T) {
	a := newAudit()
	if _, ok := a.Latest(); ok {
		t.Fatal("latest on empty should be false")
	}
	a.Add("evt-1001", "created", "张工")
	latest, ok := a.Latest()
	if !ok || latest.Type != "created" {
		t.Fatalf("latest=%+v ok=%v", latest, ok)
	}
	a.Clear()
	if got := a.Count(); got != 0 {
		t.Fatalf("count after clear=%d want 0", got)
	}
}

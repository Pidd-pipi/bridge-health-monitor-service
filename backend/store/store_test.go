package store

import (
	"testing"
)

func TestStoreListSeeded(t *testing.T) {
	s := New()
	items := s.List()
	if len(items) != 2 {
		t.Fatalf("items=%d want 2", len(items))
	}
}

func TestStoreGetMissingReturnsNotFound(t *testing.T) {
	s := New()
	if _, err := s.Get("nope"); err != ErrNotFound {
		t.Fatalf("missing get err=%v want not found", err)
	}
}

func TestStoreUpdateMissingReturnsNotFound(t *testing.T) {
	s := New()
	if _, err := s.UpdateCondition("nope", "cleared", 0); err != ErrNotFound {
		t.Fatalf("missing update err=%v want not found", err)
	}
}

func TestStoreUpdateRevisionConflict(t *testing.T) {
	s := New()
	if _, err := s.UpdateCondition("br-201", "watch", 99); err != ErrRevisionConflict {
		t.Fatalf("stale revision err=%v want conflict", err)
	}
}

func TestStoreUpdateBumpsRevision(t *testing.T) {
	s := New()
	b, err := s.UpdateCondition("br-201", "watch", 1)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if b.Revision != 2 || b.Condition != "watch" {
		t.Fatalf("updated=%+v", b)
	}
}

func TestStoreHistory(t *testing.T) {
	s := New()
	if _, err := s.UpdateCondition("br-201", "watch", 1); err != nil {
		t.Fatalf("update: %v", err)
	}
	h := s.History("br-201")
	if len(h) != 1 || h[0] != "watch" {
		t.Fatalf("history=%v", h)
	}
}

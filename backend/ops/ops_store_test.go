package ops

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func seedStore(t *testing.T) *OpsStore {
	t.Helper()
	return newOpsStore(seedOpsRecords())
}

func TestLedgerGetCopyIsolated(t *testing.T) {
	s := seedStore(t)
	rec, err := s.Get(context.Background(), "evt-1001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	rec.Labels["mutated"] = "yes"
	again, err := s.Get(context.Background(), "evt-1001")
	if err != nil {
		t.Fatalf("get again: %v", err)
	}
	if again.Labels["mutated"] == "yes" {
		t.Fatalf("store returned a live reference; mutation leaked into store")
	}
}

func TestLedgerListCopiesIsolated(t *testing.T) {
	s := seedStore(t)
	items, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("no items")
	}
	items[0].Labels["mutated"] = "yes"
	again, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("list again: %v", err)
	}
	for _, item := range again {
		if item.Labels["mutated"] == "yes" {
			t.Fatalf("list returned a live reference; mutation leaked into store")
		}
	}
}

func TestLedgerUpdateIsolated(t *testing.T) {
	s := seedStore(t)
	rec, err := s.Get(context.Background(), "evt-1003")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	rec.Labels["extra"] = "leak"
	if err := s.Update(context.Background(), rec, rec.Revision); err != nil {
		t.Fatalf("update: %v", err)
	}
	// mutate the caller's map again after the update completed
	rec.Labels["extra"] = "changed-after"
	stored, err := s.Get(context.Background(), "evt-1003")
	if err != nil {
		t.Fatalf("get stored: %v", err)
	}
	if stored.Labels["extra"] != "leak" {
		t.Fatalf("update stored a live reference; later caller mutation changed store value %q", stored.Labels["extra"])
	}
}

func TestLedgerConcurrentReadWriteStable(t *testing.T) {
	s := seedStore(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				_, _ = s.List(context.Background())
			}
		}()
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				rec, err := s.Get(context.Background(), "evt-1001")
				if err != nil {
					continue
				}
				rec.Status = OpsStatusPaused
				_ = s.Update(context.Background(), rec, rec.Revision)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := s.Count(); got != 3 {
		t.Fatalf("count=%d want 3", got)
	}
}

func TestEventStorePutRejectsDuplicate(t *testing.T) {
	s := seedStore(t)
	rec, _ := s.Get(context.Background(), "evt-1002")
	if err := s.Put(context.Background(), rec); err != ErrOpsConflict {
		t.Fatalf("duplicate put: err=%v want conflict", err)
	}
}

func TestEventStoreGetMissingNotFound(t *testing.T) {
	s := seedStore(t)
	if _, err := s.Get(context.Background(), "nope"); !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("missing get: err=%v want not found", err)
	}
}

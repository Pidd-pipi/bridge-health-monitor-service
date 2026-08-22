package ops

import (
	"testing"
)

func TestSearchFilterKeepsSourceIntact(t *testing.T) {
	items := []OpsRecord{
		{ID: "a", Subject: "甲", Status: OpsStatusActive},
		{ID: "b", Subject: "乙", Status: OpsStatusQueued},
		{ID: "c", Subject: "丙", Status: OpsStatusActive},
	}
	source := append([]OpsRecord(nil), items...)
	filtered := FilterRecords(items, OpsQuery{Status: OpsStatusActive})
	if len(filtered) != 2 || filtered[0].ID != "a" || filtered[1].ID != "c" {
		t.Fatalf("filtered=%+v want [a c]", filtered)
	}
	for i := range source {
		if items[i].ID != source[i].ID || items[i].Subject != source[i].Subject {
			t.Fatalf("filter mutated source backing array at %d: got %+v want %+v", i, items[i], source[i])
		}
	}
}

func TestPageDoesNotShareBackingArray(t *testing.T) {
	items := seedOpsRecords()
	page, total := Paginate(items, 1, 2)
	if total != 3 {
		t.Fatalf("total=%d want 3", total)
	}
	if len(page) != 2 {
		t.Fatalf("page len=%d want 2", len(page))
	}
	page[0].Subject = "mutated"
	if items[0].Subject == "mutated" {
		t.Fatal("page shares backing array with source")
	}
}

func TestPaginationSecondPageCorrect(t *testing.T) {
	items := seedOpsRecords()
	page, _ := Paginate(items, 2, 2)
	if len(page) != 1 {
		t.Fatalf("page2 len=%d want 1", len(page))
	}
	if page[0].ID != items[2].ID {
		t.Fatalf("page2 item=%s want %s", page[0].ID, items[2].ID)
	}
}

func TestMergeLabelsDoesNotMutateBase(t *testing.T) {
	base := map[string]string{"site": "东江大桥"}
	merged := MergeLabels(base, map[string]string{"reviewed": "yes"})
	if merged["reviewed"] != "yes" {
		t.Fatal("merged missing reviewed")
	}
	base["site"] = "polluted"
	if merged["site"] != "东江大桥" {
		t.Fatal("merge mutated base map")
	}
}

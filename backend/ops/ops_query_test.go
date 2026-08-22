package ops

import (
	"testing"
)

func TestEventQueryMatchesFilters(t *testing.T) {
	items := seedOpsRecords()
	if !opsMatch(items[0], OpsQuery{Status: OpsStatusActive}) {
		t.Fatal("active filter should match evt-1001")
	}
	if opsMatch(items[0], OpsQuery{Status: OpsStatusQueued}) {
		t.Fatal("queued filter should not match active event")
	}
	if !opsMatch(items[1], OpsQuery{Subject: "沉降"}) {
		t.Fatal("subject filter should match evt-1002")
	}
	if !opsMatch(items[2], OpsQuery{Priority: OpsPriorityNormal}) {
		t.Fatal("priority filter should match evt-1003")
	}
}

func TestEventQueryDateWindow(t *testing.T) {
	items := seedOpsRecords()
	from := "2026-08-19T00:00:00Z"
	to := "2026-08-21T00:00:00Z"
	for _, item := range items {
		if !opsMatchDate(item, from, to) {
			t.Fatalf("item %s should be inside window", item.ID)
		}
	}
	if opsMatchDate(items[0], "2026-08-21T00:00:00Z", "") {
		t.Fatal("item before from should be excluded")
	}
}

func TestPageCloneDetached(t *testing.T) {
	items := seedOpsRecords()
	page := OpsPage{Items: items}
	cloned := opsClonePage(page)
	cloned.Items[0].Subject = "mutated"
	if items[0].Subject == "mutated" {
		t.Fatal("cloned page shares backing array with source")
	}
}

func TestEventBoundsClamp(t *testing.T) {
	total := 3
	start, end := opsBounds(total, 5, 10)
	if start != 3 || end != 3 {
		t.Fatalf("bounds beyond total: %d,%d want 3,3", start, end)
	}
	start, end = opsBounds(total, 1, 2)
	if start != 0 || end != 2 {
		t.Fatalf("bounds page1: %d,%d want 0,2", start, end)
	}
}

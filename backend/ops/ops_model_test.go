package ops

import (
	"testing"
)

func TestEventModelCloneDeepCopiesLabels(t *testing.T) {
	rec := OpsRecord{Labels: map[string]string{"site": "东江大桥"}}
	cloned := rec.Clone()
	cloned.Labels["hack"] = "1"
	if rec.Labels["hack"] == "1" {
		t.Fatal("clone shares labels map")
	}
}

func TestEventModelNormalize(t *testing.T) {
	rec := normalizeOpsRecord(OpsRecord{ID: "  EVT-9 ", Subject: "  桥面  渗水  ", Owner: " 吴工 ", Revision: 0, Labels: nil})
	if rec.ID != "evt-9" || rec.Subject != "桥面 渗水" || rec.Owner != "吴工" {
		t.Fatalf("normalize wrong: %+v", rec)
	}
	if rec.Revision != 1 || rec.Labels == nil {
		t.Fatalf("normalize defaults wrong: %+v", rec)
	}
}

func TestEventModelSortByPriority(t *testing.T) {
	items := []OpsRecord{
		{ID: "a", Priority: OpsPriorityLow, UpdatedAt: "2026-08-01T00:00:00Z"},
		{ID: "b", Priority: OpsPriorityCritical, UpdatedAt: "2026-08-01T00:00:00Z"},
		{ID: "c", Priority: OpsPriorityNormal, UpdatedAt: "2026-08-03T00:00:00Z"},
	}
	sortOpsRecords(items)
	if items[0].ID != "b" || items[1].ID != "c" || items[2].ID != "a" {
		t.Fatalf("sort wrong: %+v", items)
	}
}

func TestModelTerminal(t *testing.T) {
	if (OpsRecord{Status: OpsStatusClosed}).Terminal() != true {
		t.Fatal("closed should be terminal")
	}
	if (OpsRecord{Status: OpsStatusReviewing}).Terminal() != false {
		t.Fatal("reviewing should not be terminal")
	}
}

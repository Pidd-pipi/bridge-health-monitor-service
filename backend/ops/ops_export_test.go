package ops

import (
	"context"
	"strings"
	"testing"
)

func TestEventExportCSVHeaderAndRows(t *testing.T) {
	svc := serviceWithSeed(t)
	csv, err := svc.ExportCSV(context.Background(), OpsQuery{})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.HasPrefix(csv, "id,subject,owner,status,priority,revision,created_at,updated_at\n") {
		t.Fatalf("bad header: %q", csv)
	}
	if !strings.Contains(csv, "evt-1001") || !strings.Contains(csv, "evt-1002") || !strings.Contains(csv, "evt-1003") {
		t.Fatalf("export missing rows: %q", csv)
	}
}

func TestEventExportCSVQuotesSpecialChars(t *testing.T) {
	svc := serviceWithSeed(t)
	rec, _ := svc.Create(context.Background(), OpsRecord{Subject: "含,逗号 与\"引号", Owner: "吴工", Priority: OpsPriorityLow, Labels: map[string]string{"site": "东江大桥", "operator": "吴工", "evidence": "q1"}})
	_ = rec
	csv, err := svc.ExportCSV(context.Background(), OpsQuery{})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(csv, `"含,逗号 与""引号"`) {
		t.Fatalf("quoting wrong: %q", csv)
	}
}

func TestEventExportCSVFiltered(t *testing.T) {
	svc := serviceWithSeed(t)
	csv, err := svc.ExportCSV(context.Background(), OpsQuery{Priority: OpsPriorityCritical})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if strings.Contains(csv, "evt-1002") {
		t.Fatalf("filtered export included other priority: %q", csv)
	}
	if !strings.Contains(csv, "evt-1001") {
		t.Fatalf("filtered export missing critical: %q", csv)
	}
}

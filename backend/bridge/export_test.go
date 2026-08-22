package bridge

import (
	"context"
	"example.com/bridge-health-monitor-service/store"
	"strings"
	"testing"
)

func TestBridgeExportRows(t *testing.T) {
	svc := New(store.New())
	csv, err := svc.ExportCSV(context.Background(), []string{"br-201"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(csv, "br-201,东江大桥,东江,monitored,24,1") {
		t.Fatalf("export row wrong: %q", csv)
	}
}

func TestBridgeExportMissingRowPresent(t *testing.T) {
	svc := New(store.New())
	csv, err := svc.ExportCSV(context.Background(), []string{"br-201", "nope"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(csv, "nope,,,missing,0,0") {
		t.Fatalf("missing row absent: %q", csv)
	}
}

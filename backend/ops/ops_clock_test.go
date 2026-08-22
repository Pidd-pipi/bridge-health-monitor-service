package ops

import (
	"context"
	"testing"
	"time"
)

func TestClockNilNowFuncSafe(t *testing.T) {
	c := OpsClock{}
	now := c.Now()
	if now.IsZero() {
		t.Fatalf("zero clock returned zero time")
	}
}

func TestClockStampRFC3339(t *testing.T) {
	c := newOpsClock()
	stamp := c.Stamp()
	if _, err := time.Parse(time.RFC3339Nano, stamp); err != nil {
		t.Fatalf("stamp not RFC3339: %v", err)
	}
}

func TestOpsDelayRespectsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := opsDelay(ctx, time.Second); err == nil {
		t.Fatal("canceled delay should return error")
	}
}

func TestOpsAgeParsing(t *testing.T) {
	now := time.Now().UTC()
	if got := opsAge(now, now.Add(-time.Hour).Format(time.RFC3339Nano)); got != time.Hour {
		t.Fatalf("age=%v want 1h", got)
	}
	if got := opsAge(now, "not-a-time"); got != 0 {
		t.Fatalf("bad stamp age=%v want 0", got)
	}
}

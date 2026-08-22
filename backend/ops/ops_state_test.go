package ops

import (
	"testing"
)

func TestTransitionTableLegalEdges(t *testing.T) {
	cases := []struct{ from, to OpsStatus }{
		{OpsStatusQueued, OpsStatusActive},
		{OpsStatusQueued, OpsStatusClosed},
		{OpsStatusActive, OpsStatusPaused},
		{OpsStatusActive, OpsStatusReviewing},
		{OpsStatusReviewing, OpsStatusClosed},
		{OpsStatusReviewing, OpsStatusActive},
		{OpsStatusPaused, OpsStatusActive},
	}
	for _, c := range cases {
		if !opsTransitionTable[c.from][c.to] {
			t.Fatalf("missing edge %s -> %s", c.from, c.to)
		}
	}
}

func TestTransitionIllegalEdgeRejected(t *testing.T) {
	m := newOpsStateMachine()
	if err := m.Move(OpsStatusQueued, OpsStatusPaused, "x"); err == nil {
		t.Fatal("queued -> paused should be rejected")
	}
	if err := m.Move(OpsStatusClosed, OpsStatusActive, "x"); err == nil {
		t.Fatal("closed -> active should be rejected")
	}
}

func TestIntermediateStateKnown(t *testing.T) {
	if !opsStatusValid(OpsStatusReviewing) {
		t.Fatal("reviewing should be a valid status")
	}
}

func TestReviewingNonTerminal(t *testing.T) {
	if opsStatusTerminal(OpsStatusReviewing) {
		t.Fatal("reviewing should not be terminal")
	}
	if !opsStatusTerminal(OpsStatusClosed) {
		t.Fatal("closed should be terminal")
	}
}

func TestStateMachineMoveRecordsHistory(t *testing.T) {
	m := newOpsStateMachine()
	_ = m.Move(OpsStatusQueued, OpsStatusActive, "dispatch")
	last, ok := m.Last()
	if !ok || last.From != OpsStatusQueued || last.To != OpsStatusActive {
		t.Fatalf("last=%+v ok=%v", last, ok)
	}
}

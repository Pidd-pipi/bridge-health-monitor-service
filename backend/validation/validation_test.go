package validation

import "testing"

func TestConditionAllowsKnownValues(t *testing.T) {
	for _, v := range []string{"monitored", "watch", "restricted", "cleared"} {
		if err := Condition(v); err != nil {
			t.Fatalf("condition %q rejected: %v", v, err)
		}
	}
	if err := Condition("broken"); err == nil {
		t.Fatal("unknown condition should be rejected")
	}
}

func TestTransitionAllowsSameState(t *testing.T) {
	if err := Transition("watch", "watch"); err != nil {
		t.Fatalf("same-state transition rejected: %v", err)
	}
}

func TestTransitionLegalEdges(t *testing.T) {
	for _, c := range [][2]string{
		{"monitored", "watch"},
		{"monitored", "restricted"},
		{"watch", "restricted"},
		{"restricted", "cleared"},
		{"watch", "cleared"},
	} {
		if err := Transition(c[0], c[1]); err != nil {
			t.Fatalf("legal transition %s->%s rejected: %v", c[0], c[1], err)
		}
	}
}

func TestTransitionIllegalEdgeRejected(t *testing.T) {
	if err := Transition("restricted", "monitored"); err == nil {
		t.Fatal("restricted -> monitored should be rejected")
	}
	if err := Transition("cleared", "restricted"); err == nil {
		t.Fatal("cleared -> restricted should be rejected")
	}
}

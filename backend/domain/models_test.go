package domain

import "testing"

func TestConditionValid(t *testing.T) {
	if !ConditionValid("monitored") || !ConditionValid("cleared") {
		t.Fatal("known conditions should be valid")
	}
	if ConditionValid("broken") {
		t.Fatal("unknown condition should be invalid")
	}
}

func TestConditionTransitionTable(t *testing.T) {
	for _, c := range [][2]string{
		{"monitored", "watch"},
		{"monitored", "restricted"},
		{"watch", "restricted"},
		{"watch", "cleared"},
		{"restricted", "cleared"},
	} {
		if !ConditionTransition(c[0], c[1]) {
			t.Fatalf("missing edge %s->%s", c[0], c[1])
		}
	}
	if ConditionTransition("restricted", "monitored") {
		t.Fatal("illegal edge restricted->monitored present")
	}
}

func TestConditionTransitionSameState(t *testing.T) {
	if !ConditionTransition("cleared", "cleared") {
		t.Fatal("same-state should be allowed")
	}
}

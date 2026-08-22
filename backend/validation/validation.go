package validation

import (
	"example.com/bridge-health-monitor-service/domain"
	"fmt"
)

// Condition validates a standalone condition value.
func Condition(v string) error {
	if !domain.ConditionValid(v) {
		return fmt.Errorf("unsupported condition %q", v)
	}
	return nil
}

// Transition validates moving a bridge from `current` to `next`.
func Transition(current, next string) error {
	if err := Condition(next); err != nil {
		return err
	}
	if !domain.ConditionTransition(current, next) {
		return fmt.Errorf("condition transition %q -> %q is not allowed", current, next)
	}
	return nil
}

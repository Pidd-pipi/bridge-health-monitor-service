package ops

import (
	"context"
	"fmt"
)

type PolicyResult struct {
	Code      string      `json:"code"`
	Name      string      `json:"name"`
	Severity  OpsPriority `json:"severity"`
	Satisfied bool        `json:"satisfied"`
	Missing   []string    `json:"missing"`
}

func (s *OpsService) EvaluatePolicy(ctx context.Context, id string) (PolicyResult, error) {
	record, err := s.store.Get(ctx, id)
	if err != nil {
		return PolicyResult{}, err
	}
	rule := findRule(record.Priority)
	missing := make([]string, 0)
	for _, label := range rule.RequiredLabels {
		if record.LabelValue(label) == "" {
			missing = append(missing, label)
		}
	}
	if len(missing) > 0 {
		return PolicyResult{Code: rule.Code, Name: rule.Name, Severity: rule.Severity, Satisfied: false, Missing: missing}, nil
	}
	return PolicyResult{Code: rule.Code, Name: rule.Name, Severity: rule.Severity, Satisfied: true, Missing: missing}, nil
}

// RequireReview rejects closing a critical record that has not been reviewed.
func RequireReview(record OpsRecord) error {
	if record.Priority == OpsPriorityCritical && record.LabelValue("reviewed") == "" {
		return fmt.Errorf("%w: critical record requires review before close", ErrOpsReviewRequired)
	}
	return nil
}

func findRule(priority OpsPriority) OpsRule {
	for _, rule := range opsRules() {
		if rule.Severity == priority {
			return rule
		}
	}
	return OpsRule{Code: "OPS-0000", Name: "fallback", Severity: OpsPriorityNormal, RequiredLabels: []string{"site"}}
}

package domain

// ConditionTransition reports whether a bridge condition may move from
// `from` to `to`. Same-state updates are always allowed.
func ConditionTransition(from, to string) bool {
	if from == to {
		return true
	}
	return conditionTransitions[from][to]
}

var conditionTransitions = map[string]map[string]bool{
	"monitored":  {"watch": true, "restricted": true},
	"watch":      {"restricted": true, "cleared": true},
	"restricted": {"cleared": true},
	"cleared":    {"watch": true},
}

// ConditionValid reports whether value is a known bridge condition.
func ConditionValid(value string) bool {
	_, ok := conditionTransitions[value]
	return ok
}

type Bridge struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	River      string `json:"river"`
	Condition  string `json:"condition"`
	LastSurvey string `json:"last_survey"`
	RiskScore  int    `json:"risk_score"`
	Revision   int    `json:"revision"`
}

type ConditionChange struct {
	Condition        string `json:"condition"`
	ExpectedRevision int    `json:"expected_revision"`
}

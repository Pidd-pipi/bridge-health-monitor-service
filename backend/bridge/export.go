package bridge

import (
	"context"
	"example.com/bridge-health-monitor-service/store"
	"fmt"
	"strings"
)

// ExportCSV renders a CSV row per requested bridge id, including a row for
// ids that could not be found.
func (svc *Service) ExportCSV(ctx context.Context, ids []string) (string, error) {
	var b strings.Builder
	b.WriteString("id,name,river,condition,risk_score,revision\n")
	for _, id := range ids {
		br, err := svc.store.Get(id)
		if err != nil {
			if err == store.ErrNotFound {
				continue
			}
			return "", fmt.Errorf("bridge export: %w", err)
		}
		b.WriteString(csvBridgeRow(br.ID, br.Name, br.River, br.Condition, br.RiskScore, br.Revision))
	}
	return b.String(), nil
}

func csvBridgeRow(id, name, river, condition string, risk, revision int) string {
	fields := []string{id, name, river, condition, fmt.Sprintf("%d", risk), fmt.Sprintf("%d", revision)}
	for i, f := range fields {
		fields[i] = csvBridgeQuote(f)
	}
	return strings.Join(fields, ",") + "\n"
}

func csvBridgeQuote(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

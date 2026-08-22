package bridge

import (
	"context"
	"example.com/bridge-health-monitor-service/store"
	"fmt"
	"sort"
)

type Summary struct {
	Total       int            `json:"total"`
	Missing     int            `json:"missing"`
	ByCondition map[string]int `json:"by_condition"`
	ByRisk      map[string]int `json:"by_risk"`
	HighRisk    []string       `json:"high_risk"`
}

// Report aggregates bridge state for the given ids. Unknown ids are counted
// as missing instead of failing the whole report, and duplicate ids are
// collapsed into a single entry.
func (svc *Service) Report(ctx context.Context, ids []string) (Summary, error) {
	r := Summary{ByCondition: map[string]int{}, ByRisk: map[string]int{}}
	for _, id := range ids {
		b, err := svc.Get(ctx, id)
		if err != nil {
			if err == store.ErrNotFound {
				r.Missing++
				continue
			}
			return Summary{}, fmt.Errorf("bridge report: %w", err)
		}
		r.Total++
		r.ByCondition[b.Condition]++
		level := svc.RiskLevel(b.RiskScore)
		r.ByRisk[level]++
		if level == "high" {
			r.HighRisk = append(r.HighRisk, b.ID)
		}
	}
	sort.Strings(r.HighRisk)
	return r, nil
}

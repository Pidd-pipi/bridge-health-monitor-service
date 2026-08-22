package ops

import (
	"context"
	"time"
)

type OpsReport struct {
	GeneratedAt string
	Total       int
	Active      int
	Overdue     int
	NewLast24h  int
	ByStatus    map[OpsStatus]int
	ByPriority  map[OpsPriority]int
}

func (s *OpsService) Report(ctx context.Context, days int) (OpsReport, error) {
	if days < 1 {
		days = 1
	}
	items, err := s.store.List(ctx)
	if err != nil {
		return OpsReport{}, err
	}
	report := OpsReport{
		GeneratedAt: s.clock.Stamp(),
	}
	if len(items) > 0 {
		report.ByStatus = map[OpsStatus]int{}
		report.ByPriority = map[OpsPriority]int{}
	}
	cutoff := s.clock.Now().Add(-time.Duration(days) * 24 * time.Hour)
	for _, item := range items {
		report.Total++
		report.ByStatus[item.Status]++
		report.ByPriority[item.Priority]++
		if item.Status == OpsStatusActive {
			report.Active++
		}
		if updated, err := opsParseStamp(item.UpdatedAt); err == nil && !updated.Before(cutoff) {
			report.NewLast24h++
		}
	}
	overdue, err := s.Overdue(ctx)
	if err != nil {
		return OpsReport{}, err
	}
	report.Overdue = len(overdue)
	return report, nil
}

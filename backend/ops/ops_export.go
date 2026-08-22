package ops

import (
	"context"
	"strconv"
	"strings"
)

func (s *OpsService) ExportCSV(ctx context.Context, q OpsQuery) (string, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return "", err
	}
	filtered := FilterRecords(items, q)
	sortOpsRecords(filtered)
	var b strings.Builder
	b.WriteString("id,subject,owner,status,priority,revision,created_at,updated_at\n")
	for _, item := range filtered {
		b.WriteString(csvLine(item))
	}
	return b.String(), nil
}

func csvLine(item OpsRecord) string {
	fields := []string{
		item.ID, item.Subject, item.Owner, string(item.Status), string(item.Priority),
		strconv.Itoa(item.Revision), item.CreatedAt, item.UpdatedAt,
	}
	for i, f := range fields {
		fields[i] = csvQuote(f)
	}
	return strings.Join(fields, ",") + "\n"
}

func csvQuote(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

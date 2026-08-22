package ops

import "strings"

func opsMatch(item OpsRecord, query OpsQuery) bool {
	if query.Subject != "" && !strings.Contains(strings.ToLower(item.Subject), strings.ToLower(query.Subject)) {
		return false
	}
	if query.Status != "" && item.Status != query.Status {
		return false
	}
	if query.Priority != "" && item.Priority != query.Priority {
		return false
	}
	if query.Owner != "" && item.Owner != query.Owner {
		return false
	}
	return true
}

// opsMatchDate bounds the record by its updated timestamp. Records whose
// timestamp cannot be parsed are excluded from the window.
func opsMatchDate(item OpsRecord, from, to string) bool {
	updated, err := opsParseStamp(item.UpdatedAt)
	if err != nil {
		return from == "" && to == ""
	}
	if from != "" {
		lower, err := opsParseStamp(from)
		if err == nil && updated.Before(lower) {
			return false
		}
	}
	if to != "" {
		upper, err := opsParseStamp(to)
		if err == nil && updated.After(upper) {
			return false
		}
	}
	return true
}

func opsQueryDefaults(q OpsQuery) OpsQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 25
	}
	if q.PageSize > 200 {
		q.PageSize = 200
	}
	return q
}

func opsBounds(total, page, size int) (int, int) {
	q := opsQueryDefaults(OpsQuery{Page: page, PageSize: size})
	start := (q.Page - 1) * q.PageSize
	if start > total {
		start = total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return start, end
}

func opsPageCount(total, size int) int {
	if size < 1 || total == 0 {
		return 0
	}
	return (total + size - 1) / size
}

func opsQueryKey(q OpsQuery) string {
	return strings.Join([]string{q.Subject, string(q.Status), string(q.Priority), q.Owner, q.From, q.To}, "|")
}

func opsClonePage(p OpsPage) OpsPage {
	p.Items = append([]OpsRecord(nil), p.Items...)
	return p
}

func opsHasNext(p OpsPage) bool { return p.HasNext }
func opsFirstID(p OpsPage) string {
	if len(p.Items) == 0 {
		return ""
	}
	return p.Items[0].ID
}
func opsLastID(p OpsPage) string {
	if len(p.Items) == 0 {
		return ""
	}
	return p.Items[len(p.Items)-1].ID
}

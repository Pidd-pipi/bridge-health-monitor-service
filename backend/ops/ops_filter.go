package ops

// FilterRecords returns the records matching the query filters. It must not
// mutate the input slice or share its backing array with the caller. Each
// returned record is a copy so callers cannot reach the store's label maps.
func FilterRecords(items []OpsRecord, q OpsQuery) []OpsRecord {
	out := make([]OpsRecord, 0, len(items))
	for _, item := range items {
		if opsMatch(item, q) && opsMatchDate(item, q.From, q.To) {
			out = append(out, item.Clone())
		}
	}
	return out
}

// Paginate returns the page of records and the total number of records.
// The returned page must not retain a reference to the input backing array.
func Paginate(items []OpsRecord, page, pageSize int) ([]OpsRecord, int) {
	q := opsQueryDefaults(OpsQuery{Page: page, PageSize: pageSize})
	start, end := opsBounds(len(items), q.Page, q.PageSize)
	pageItems := make([]OpsRecord, 0, end-start)
	for i := start; i < end; i++ {
		pageItems = append(pageItems, items[i].Clone())
	}
	return pageItems, len(items)
}

// MergeLabels merges extra labels into a copy of base. The base map passed by
// the caller must not be modified.
func MergeLabels(base map[string]string, extra map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return merged
}

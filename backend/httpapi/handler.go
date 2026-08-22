package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/bridge-health-monitor-service/bridge"
	"example.com/bridge-health-monitor-service/domain"
	"example.com/bridge-health-monitor-service/ops"
	"example.com/bridge-health-monitor-service/store"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	bridges *bridge.Service
	events  *ops.OpsService
}

func New(b *bridge.Service, e *ops.OpsService) *Handler { return &Handler{bridges: b, events: e} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	segs := splitPath(path)
	if len(segs) < 3 || segs[0] != "api" || segs[1] != "v1" {
		http.Error(w, "route not found", 404)
		return
	}
	switch segs[2] {
	case "bridges":
		h.serveBridges(w, r, segs)
	case "events":
		h.serveEvents(w, r, segs)
	default:
		http.Error(w, "route not found", 404)
	}
}

var staleRequestCtx = context.Background()

func requestCtx(r *http.Request) context.Context { return staleRequestCtx }

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

// ---- bridges ----

func (h *Handler) serveBridges(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) == 3 {
		if r.Method == http.MethodGet {
			write(w, 200, map[string]any{"items": h.bridges.List(r.Context())})
			return
		}
		http.Error(w, "method not allowed", 405)
		return
	}
	switch segs[3] {
	case "report":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		ids := strings.Split(r.URL.Query().Get("ids"), ",")
		summary, err := h.bridges.Report(r.Context(), ids)
		if err != nil {
			http.Error(w, "bridge report failed", 500)
			return
		}
		write(w, 200, summary)
	case "export":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		ids := strings.Split(r.URL.Query().Get("ids"), ",")
		csv, err := h.bridges.ExportCSV(r.Context(), ids)
		if err != nil {
			http.Error(w, "bridge export failed", 500)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(csv))
	default:
		h.serveBridgeItem(w, r, segs)
	}
}

func (h *Handler) serveBridgeItem(w http.ResponseWriter, r *http.Request, segs []string) {
	id := segs[3]
	if len(segs) == 4 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		b, err := h.bridges.Get(r.Context(), id)
		if err != nil {
			http.Error(w, "bridge not found", 404)
			return
		}
		write(w, 200, b)
		return
	}
	switch segs[4] {
	case "history":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		history, err := h.bridges.History(r.Context(), id)
		if err != nil {
			http.Error(w, "bridge not found", 404)
			return
		}
		write(w, 200, map[string]any{"id": id, "history": history})
	case "condition":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		h.updateBridgeCondition(w, r, id)
	default:
		http.Error(w, "route not found", 404)
	}
}

func (h *Handler) updateBridgeCondition(w http.ResponseWriter, r *http.Request, id string) {
	var c domain.ConditionChange
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	b, err := h.bridges.UpdateCondition(requestCtx(r), id, c.Condition, c.ExpectedRevision)
	if err != nil {
		status := 400
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, bridge.ErrBridgeNotFound) {
			status = 404
		}
		if errors.Is(err, store.ErrRevisionConflict) {
			status = 409
		}
		http.Error(w, err.Error(), status)
		return
	}
	write(w, 200, b)
}

// ---- events ----

func (h *Handler) serveEvents(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) == 3 {
		switch r.Method {
		case http.MethodGet:
			h.listEvents(w, r)
		case http.MethodPost:
			h.createEvent(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	switch segs[3] {
	case "report":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))
		report, err := h.events.Report(requestCtx(r), days)
		if err != nil {
			http.Error(w, "events report failed", 500)
			return
		}
		write(w, 200, report)
	case "export":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		csv, err := h.events.ExportCSV(requestCtx(r), queryFromRequest(r))
		if err != nil {
			http.Error(w, "events export failed", 500)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(csv))
	case "overdue":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		items, err := h.events.Overdue(requestCtx(r))
		if err != nil {
			http.Error(w, "overdue failed", 500)
			return
		}
		write(w, 200, map[string]any{"items": items})
	case "batch-close":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		h.batchClose(w, r)
	default:
		h.serveEventItem(w, r, segs)
	}
}

func (h *Handler) serveEventItem(w http.ResponseWriter, r *http.Request, segs []string) {
	id := segs[3]
	if len(segs) == 4 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		record, err := h.events.Get(requestCtx(r), id)
		if err != nil {
			writeOpsError(w, err)
			return
		}
		write(w, 200, record)
		return
	}
	switch segs[4] {
	case "audit":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		write(w, 200, map[string]any{"id": id, "events": h.events.Audit(id)})
	case "policy":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		result, err := h.events.EvaluatePolicy(requestCtx(r), id)
		if err != nil {
			writeOpsError(w, err)
			return
		}
		write(w, 200, result)
	case "deadline":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		due, err := h.events.Deadline(requestCtx(r), id)
		if err != nil {
			writeOpsError(w, err)
			return
		}
		write(w, 200, map[string]any{"id": id, "due_at": due.UTC().Format("2006-01-02T15:04:05Z07:00")})
	case "transition":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		h.transitionEvent(w, r, id)
	case "review":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		h.reviewEvent(w, r, id)
	case "close-reviewed":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		h.closeReviewed(w, r, id)
	default:
		http.Error(w, "route not found", 404)
	}
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	page, err := h.events.Search(context.Background(), queryFromRequest(r))
	if err != nil {
		writeOpsError(w, err)
		return
	}
	write(w, 200, map[string]any{
		"items": page.Items, "page": page.Page, "page_size": page.PageSize,
		"total": page.Total, "has_next": page.HasNext,
	})
}

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Subject  string            `json:"subject"`
		Owner    string            `json:"owner"`
		Priority ops.OpsPriority   `json:"priority"`
		Labels   map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	record, err := h.events.Create(requestCtx(r), ops.OpsRecord{Subject: body.Subject, Owner: body.Owner, Priority: body.Priority, Labels: body.Labels})
	if err != nil {
		writeOpsError(w, err)
		return
	}
	write(w, 201, record)
}

func (h *Handler) transitionEvent(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		To       ops.OpsStatus `json:"to"`
		Revision int           `json:"revision"`
		Actor    string        `json:"actor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	actor := strings.TrimSpace(body.Actor)
	if actor == "" {
		actor = "web"
	}
	record, err := h.events.Transition(requestCtx(r), id, body.Revision, body.To, actor)
	if err != nil {
		writeOpsError(w, err)
		return
	}
	write(w, 200, record)
}

func (h *Handler) reviewEvent(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Revision int    `json:"revision"`
		Actor    string `json:"actor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	record, err := h.events.Review(requestCtx(r), id, body.Revision, body.Actor)
	if err != nil {
		writeOpsError(w, err)
		return
	}
	write(w, 200, record)
}

func (h *Handler) closeReviewed(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Revision int    `json:"revision"`
		Actor    string `json:"actor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	record, err := h.events.CloseReviewed(requestCtx(r), id, body.Revision, body.Actor)
	if err != nil {
		writeOpsError(w, err)
		return
	}
	write(w, 200, record)
}

func (h *Handler) batchClose(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs   []string `json:"ids"`
		Actor string   `json:"actor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	results := h.events.BatchClose(requestCtx(r), body.IDs, body.Actor)
	write(w, 200, map[string]any{"results": results})
}

func queryFromRequest(r *http.Request) ops.OpsQuery {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("page_size"))
	return ops.OpsQuery{
		Subject:  q.Get("subject"),
		Status:   ops.OpsStatus(q.Get("status")),
		Priority: ops.OpsPriority(q.Get("priority")),
		Owner:    q.Get("owner"),
		From:     q.Get("from"),
		To:       q.Get("to"),
		Page:     page,
		PageSize: size,
	}
}

func writeOpsError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), ops.StatusForError(err))
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

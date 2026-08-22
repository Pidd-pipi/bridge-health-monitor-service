package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/bridge-health-monitor-service/bridge"
	"example.com/bridge-health-monitor-service/ops"
	"example.com/bridge-health-monitor-service/store"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *httptest.Server {
	h := New(bridge.New(store.New()), ops.NewService(ops.SeedRecords()))
	return httptest.NewServer(h)
}

func do(t *testing.T, method, url, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return res, string(data)
}

func doCtx(t *testing.T, ctx context.Context, method, url, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return res, string(data)
}

// ---- bridges: basic contract ----

func TestBridgeAPI(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{{"GET", "/api/v1/bridges", "", 200}, {"POST", "/api/v1/bridges/br-201/condition", `{"condition":"restricted"}`, 200}, {"POST", "/api/v1/bridges/br-201/condition", `{"condition":"broken"}`, 400}, {"POST", "/api/v1/bridges/nope/condition", `{"condition":"cleared"}`, 404}} {
		res, _ := do(t, tc.method, s.URL+tc.path, tc.body)
		if res.StatusCode != tc.want {
			t.Fatalf("%s %s: status=%d want %d", tc.method, tc.path, res.StatusCode, tc.want)
		}
	}
}

func TestBridgeWatchStateUpdate(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, _ := do(t, "POST", s.URL+"/api/v1/bridges/br-201/condition", `{"condition":"watch"}`)
	if res.StatusCode != 200 {
		t.Fatalf("monitored -> watch: status=%d want 200", res.StatusCode)
	}
}

func TestBridgeClearAfterRestrict(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	if res, _ := do(t, "POST", s.URL+"/api/v1/bridges/br-202/condition", `{"condition":"restricted"}`); res.StatusCode != 200 {
		t.Fatalf("watch -> restricted: status=%d want 200", res.StatusCode)
	}
	res, body := do(t, "POST", s.URL+"/api/v1/bridges/br-202/condition", `{"condition":"cleared"}`)
	if res.StatusCode != 200 {
		t.Fatalf("restricted -> cleared: status=%d body=%s want 200", res.StatusCode, body)
	}
}

func TestBridgeIllegalConditionReject(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	if res, _ := do(t, "POST", s.URL+"/api/v1/bridges/br-201/condition", `{"condition":"restricted"}`); res.StatusCode != 200 {
		t.Fatalf("setup restricted failed: %d", res.StatusCode)
	}
	res, body := do(t, "POST", s.URL+"/api/v1/bridges/br-201/condition", `{"condition":"monitored"}`)
	if res.StatusCode != 400 {
		t.Fatalf("restricted -> monitored: status=%d body=%s want 400", res.StatusCode, body)
	}
}

func TestBridgeSameConditionAllow(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/bridges/br-202/condition", `{"condition":"watch"}`)
	if res.StatusCode != 200 {
		t.Fatalf("watch -> watch: status=%d body=%s want 200", res.StatusCode, body)
	}
}

func TestFetchMissingBridgeRejects(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/bridges/nope", "")
	if res.StatusCode != 404 {
		t.Fatalf("GET missing bridge: status=%d body=%s want 404", res.StatusCode, body)
	}
}

func TestUpdateMissingBridgeRejects(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/bridges/nope/condition", `{"condition":"cleared"}`)
	if res.StatusCode != 404 {
		t.Fatalf("POST missing bridge: status=%d body=%s want 404", res.StatusCode, body)
	}
}

func TestBridgeUpdateRevisionConflict(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, _ := do(t, "POST", s.URL+"/api/v1/bridges/br-201/condition", `{"condition":"watch","expected_revision":99}`)
	if res.StatusCode != 409 {
		t.Fatalf("stale revision: status=%d want 409", res.StatusCode)
	}
}

// ---- bridges: report / export ----

func TestSummaryCountsMissing(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/bridges/report?ids=br-201,nope", "")
	if res.StatusCode != 200 {
		t.Fatalf("report: status=%d body=%s want 200", res.StatusCode, body)
	}
	var summary struct {
		Total   int `json:"total"`
		Missing int `json:"missing"`
	}
	if err := json.Unmarshal([]byte(body), &summary); err != nil {
		t.Fatalf("decode report: %v body=%s", err, body)
	}
	if summary.Total != 1 || summary.Missing != 1 {
		t.Fatalf("report totals: total=%d missing=%d want 1/1", summary.Total, summary.Missing)
	}
}

func TestSummaryDeduplication(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/bridges/report?ids=br-201,br-201", "")
	if res.StatusCode != 200 {
		t.Fatalf("report: status=%d body=%s want 200", res.StatusCode, body)
	}
	var summary struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal([]byte(body), &summary); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if summary.Total != 1 {
		t.Fatalf("duplicate ids counted twice: total=%d want 1", summary.Total)
	}
}

func TestExportHasMissingRow(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/bridges/export?ids=br-201,nope", "")
	if res.StatusCode != 200 {
		t.Fatalf("export: status=%d want 200", res.StatusCode)
	}
	if !strings.Contains(body, "nope,,,missing,0,0") {
		t.Fatalf("export missing row absent: %q", body)
	}
}

// ---- events: create / search / transition ----

func TestEventCreateAndSearch(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events", `{"subject":"支座锈蚀复核","owner":"赵工","priority":"high","labels":{"site":"东江大桥","operator":"赵工","evidence":"ph-9"}}`)
	if res.StatusCode != 201 {
		t.Fatalf("create event: status=%d body=%s want 201", res.StatusCode, body)
	}
	res, body = do(t, "GET", s.URL+"/api/v1/events?subject=%E9%94%88%E8%9A%80", "")
	if res.StatusCode != 200 {
		t.Fatalf("search: status=%d want 200", res.StatusCode)
	}
	if !strings.Contains(body, "支座锈蚀复核") {
		t.Fatalf("search missed created event: %s", body)
	}
}

func TestEventTransitionLifecycle(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events/evt-1002/transition", `{"to":"active","revision":1}`)
	if res.StatusCode != 200 {
		t.Fatalf("queued -> active: status=%d body=%s want 200", res.StatusCode, body)
	}
	res, body = do(t, "POST", s.URL+"/api/v1/events/evt-1002/transition", `{"to":"closed","revision":2}`)
	if res.StatusCode != 200 {
		t.Fatalf("active -> closed: status=%d body=%s want 200", res.StatusCode, body)
	}
}

func TestEventTransitionStaleRevisionConflict(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events/evt-1001/transition", `{"to":"paused","revision":1}`)
	if res.StatusCode != 409 {
		t.Fatalf("stale revision: status=%d body=%s want 409", res.StatusCode, body)
	}
}

func TestEventTransitionIllegalRejected(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events/evt-1002/transition", `{"to":"paused","revision":1}`)
	if res.StatusCode != 400 {
		t.Fatalf("queued -> paused: status=%d body=%s want 400", res.StatusCode, body)
	}
}

func TestEventUnknownGetNotFound(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/events/evt-9999", "")
	if res.StatusCode != 404 {
		t.Fatalf("unknown event: status=%d body=%s want 404", res.StatusCode, body)
	}
}

// ---- events: policy ----

func TestPolicyUnknownEventReturns404(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/events/evt-9999/policy", "")
	if res.StatusCode != 404 {
		t.Fatalf("policy unknown: status=%d body=%s want 404", res.StatusCode, body)
	}
}

func TestPolicyMissingLabelsReportsUnsatisfied(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	_, body := do(t, "POST", s.URL+"/api/v1/events", `{"subject":"钢索腐蚀报警","owner":"孙工","priority":"critical","labels":{"site":"东江大桥","operator":"孙工","evidence":"sd-1"}}`)
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("create: %v body=%s", err, body)
	}
	res, body := do(t, "GET", s.URL+"/api/v1/events/"+created.ID+"/policy", "")
	if res.StatusCode != 200 {
		t.Fatalf("policy missing labels: status=%d body=%s want 200", res.StatusCode, body)
	}
	var result struct {
		Satisfied bool     `json:"satisfied"`
		Missing   []string `json:"missing"`
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decode policy: %v", err)
	}
	if result.Satisfied {
		t.Fatalf("policy should be unsatisfied: %s", body)
	}
	if len(result.Missing) == 0 {
		t.Fatalf("policy missing list empty: %s", body)
	}
}

func TestPolicySatisfiedCritical(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/events/evt-1001/policy", "")
	if res.StatusCode != 200 {
		t.Fatalf("policy satisfied: status=%d body=%s want 200", res.StatusCode, body)
	}
	if !strings.Contains(body, `"satisfied":true`) {
		t.Fatalf("policy should be satisfied: %s", body)
	}
}

func TestTransitionCriticalWithoutReviewRejected(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	_, body := do(t, "POST", s.URL+"/api/v1/events", `{"subject":"桥面开裂新增","owner":"周工","priority":"critical","labels":{"site":"港湾跨线桥","operator":"周工","evidence":"ck-2"}}`)
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("create: %v body=%s", err, body)
	}
	res, body := do(t, "POST", s.URL+"/api/v1/events/"+created.ID+"/transition", `{"to":"closed","revision":1}`)
	if res.StatusCode != 403 {
		t.Fatalf("critical close without review: status=%d body=%s want 403", res.StatusCode, body)
	}
}

// ---- events: review lifecycle ----

func TestReviewLifecycleFlow(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events/evt-1001/review", `{"revision":3}`)
	if res.StatusCode != 200 {
		t.Fatalf("active -> reviewing: status=%d body=%s want 200", res.StatusCode, body)
	}
	res, body = do(t, "POST", s.URL+"/api/v1/events/evt-1001/close-reviewed", `{"revision":4}`)
	if res.StatusCode != 200 {
		t.Fatalf("reviewing -> closed: status=%d body=%s want 200", res.StatusCode, body)
	}
}

func TestReviewStatusPersisted(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	if res, _ := do(t, "POST", s.URL+"/api/v1/events/evt-1001/review", `{"revision":3}`); res.StatusCode != 200 {
		t.Fatalf("review failed")
	}
	res, body := do(t, "GET", s.URL+"/api/v1/events/evt-1001", "")
	if res.StatusCode != 200 {
		t.Fatalf("get after review: %d", res.StatusCode)
	}
	if !strings.Contains(body, `"status":"reviewing"`) {
		t.Fatalf("status after review should be reviewing: %s", body)
	}
}

// ---- events: batch close ----

func TestBatchClosePartial(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "POST", s.URL+"/api/v1/events/batch-close", `{"ids":["evt-1002","evt-9999","evt-1003"],"actor":"工长"}`)
	if res.StatusCode != 200 {
		t.Fatalf("batch close: status=%d body=%s want 200", res.StatusCode, body)
	}
	if !strings.Contains(body, `"ok":true`) {
		t.Fatalf("batch close should contain successes: %s", body)
	}
}

// ---- events: export ----

func TestEventExportCSVRows(t *testing.T) {
	s := newTestServer()
	defer s.Close()
	res, body := do(t, "GET", s.URL+"/api/v1/events/export?priority=critical", "")
	if res.StatusCode != 200 {
		t.Fatalf("export: status=%d want 200", res.StatusCode)
	}
	if !strings.Contains(body, "evt-1001") {
		t.Fatalf("export missing critical event: %s", body)
	}
}

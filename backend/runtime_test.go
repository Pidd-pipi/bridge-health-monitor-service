package main

import (
	"context"
	"example.com/bridge-health-monitor-service/bridge"
	"example.com/bridge-health-monitor-service/httpapi"
	"example.com/bridge-health-monitor-service/ops"
	"example.com/bridge-health-monitor-service/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func enterpriseHandler() http.Handler {
	return newEnterpriseServer("127.0.0.1:0", httpapi.New(bridge.New(store.New()), ops.NewService(ops.SeedRecords()))).Handler
}

func serveOn(handler http.Handler, method, path string, header map[string]string, ctx context.Context) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://bridge.test"+path, nil)
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestSearchHonorsClientCancel(t *testing.T) {
	h := enterpriseHandler()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rec := serveOn(h, "GET", "/api/v1/events", nil, ctx)
	if rec.Code != 504 {
		t.Fatalf("canceled search: status=%d want 504", rec.Code)
	}
}

func TestFetchDetailHonorsCancel(t *testing.T) {
	h := enterpriseHandler()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rec := serveOn(h, "GET", "/api/v1/events/evt-1001", nil, ctx)
	if rec.Code != 504 {
		t.Fatalf("canceled get: status=%d want 504", rec.Code)
	}
}

func TestHeaderDeadlineEnforced(t *testing.T) {
	h := enterpriseHandler()
	rec := serveOn(h, "GET", "/api/v1/events", map[string]string{"X-Request-Deadline-Ms": "0"}, nil)
	if rec.Code != 504 {
		t.Fatalf("header deadline 0: status=%d want 504", rec.Code)
	}
}

func TestRequestDefaultDeadlineApplies(t *testing.T) {
	h := enterpriseHandler()
	rec := serveOn(h, "GET", "/api/v1/events", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("normal request: status=%d want 200", rec.Code)
	}
}

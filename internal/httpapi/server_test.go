package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/audit"
	"agri-packaging/internal/catalog"
	"agri-packaging/internal/dashboard"
	"agri-packaging/internal/grading"
	"agri-packaging/internal/line"
	"agri-packaging/internal/packing"
	"agri-packaging/internal/shift"
	"agri-packaging/internal/storage"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Unix(12, 0) }
	c := catalog.New(store, now)
	l := line.New(store, now)
	g := grading.New(store, c, now)
	p := packing.New(store, l, now)
	d := dashboard.New(store, now)
	a := audit.New(store, now)
	sh := shift.New(store, now)
	t.Cleanup(func() { store.Close() })
	return New(Services{Catalog: c, Line: l, Grading: g, Packing: p, Dashboard: d, Audit: a, Shift: sh}, nil)
}

func TestHTTPWorkflowCreatesBatchAndDashboard(t *testing.T) {
	server := testServer(t)
	body, _ := json.Marshal(map[string]any{"id": "b1", "crop": "苹果", "received_kg": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewReader(body))
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("code=%d body=%s", res.Code, res.Body.String())
	}
	res = httptest.NewRecorder()
	server.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))
	if res.Code != http.StatusOK || !bytes.Contains(res.Body.Bytes(), []byte("pending_kg")) {
		t.Fatalf("code=%d body=%s", res.Code, res.Body.String())
	}
}

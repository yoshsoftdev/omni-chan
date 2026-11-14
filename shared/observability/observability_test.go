package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestReadyReportsFailuresWithoutLeakingErrors(t *testing.T) {
	h := Handler(NewRegistry(), map[string]Check{
		"postgres": func(context.Context) error { return nil },
		"redis":    func(context.Context) error { return errors.New("dial tcp 10.0.0.7:6379: secret-host refused") },
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret-host") {
		t.Fatalf("readiness body leaks error detail: %s", rec.Body.String())
	}
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "fail" || body.Checks["redis"] != "fail" || body.Checks["postgres"] != "ok" {
		t.Fatalf("unexpected body %+v", body)
	}

	ok := Handler(NewRegistry(), map[string]Check{"postgres": func(context.Context) error { return nil }})
	rec = httptest.NewRecorder()
	ok.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("all checks pass: status = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	ok.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "go_goroutines") {
		t.Fatalf("/metrics not served: %d", rec.Code)
	}
}

func TestFiberMiddleware(t *testing.T) {
	SetupTracing(context.Background(), "test", nil)
	reg := NewRegistry()
	metrics := NewHTTPMetrics(reg)
	var logs bytes.Buffer
	log := slog.New(WrapLogHandler(slog.NewJSONHandler(&logs, nil)))

	app := fiber.New()
	app.Use(FiberMiddleware(FiberOptions{Metrics: metrics, Log: log, TeamID: func(*fiber.Ctx) string { return "team-1" }}))
	app.Get("/api/v1/conversations/:id", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/boom", func(*fiber.Ctx) error { return errors.New("boom") })

	const parent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/123?token=supersecret", nil)
	req.Header.Set("traceparent", parent)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Test(httptest.NewRequest(http.MethodGet, "/boom", nil)); err != nil {
		t.Fatal(err)
	}

	if got := testutil.ToFloat64(metrics.requests.WithLabelValues("GET", "/api/v1/conversations/:id", "200")); got != 1 {
		t.Fatalf("route-template counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(metrics.requests.WithLabelValues("GET", "/boom", "500")); got != 1 {
		t.Fatalf("5xx counter = %v, want 1", got)
	}

	out := logs.String()
	if strings.Contains(out, "supersecret") {
		t.Fatalf("query string (credential) logged: %s", out)
	}
	if !strings.Contains(out, `"trace_id":"4bf92f3577b34da6a3ce929d0e0e4736"`) {
		t.Fatalf("incoming traceparent not continued into logs: %s", out)
	}
	if !strings.Contains(out, `"team_id":"team-1"`) || !strings.Contains(out, `"level":"ERROR"`) {
		t.Fatalf("missing team_id or error level: %s", out)
	}
	_, _ = io.Copy(io.Discard, &logs)
}

func TestSnapshotCollector(t *testing.T) {
	reg := NewRegistry()
	reg.MustRegister(NewSnapshotCollector("omni_worker", func() map[string]int64 {
		return map[string]int64{"send_success_total": 7, "worker_batches_processing": 2}
	}))
	rec := httptest.NewRecorder()
	Handler(reg, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()
	for _, want := range []string{
		"# TYPE omni_worker_send_success_total counter", "omni_worker_send_success_total 7",
		"# TYPE omni_worker_worker_batches_processing gauge", "omni_worker_worker_batches_processing 2",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in:\n%s", want, body)
		}
	}
}

func TestLabeledCounterCollector(t *testing.T) {
	reg := NewRegistry()
	reg.MustRegister(NewLabeledCounterCollector("omni_broadcast_send_failures_by_code_total", "x", "code",
		func() map[string]int64 { return map[string]int64{"invalid_number": 3, "blocked": 1} }))
	rec := httptest.NewRecorder()
	Handler(reg, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, want := range []string{
		`omni_broadcast_send_failures_by_code_total{code="invalid_number"} 3`,
		`omni_broadcast_send_failures_by_code_total{code="blocked"} 1`,
	} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("missing %q", want)
		}
	}
}

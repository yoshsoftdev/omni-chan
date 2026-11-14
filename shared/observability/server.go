// Package observability gives every process the same operational surface:
// Prometheus metrics, liveness/readiness, OpenTelemetry tracing, and log
// correlation. It serves on a separate internal address so nothing is added to
// the public or frontend-facing HTTP APIs.
package observability

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Check reports whether one dependency is usable. Return nil when healthy.
type Check func(ctx context.Context) error

// NewRegistry returns a registry with Go runtime and process metrics.
func NewRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return reg
}

// MetricsAddr reads METRICS_ADDR, falling back to a loopback default so a host
// running several services never exposes the endpoint by accident. Containers
// set METRICS_ADDR=":<port>" and keep the port unpublished.
func MetricsAddr(defaultAddr string) string {
	if v := os.Getenv("METRICS_ADDR"); v != "" {
		return v
	}
	return defaultAddr
}

// Handler serves /metrics, /health (process alive) and /ready (all checks pass).
func Handler(reg *prometheus.Registry, checks map[string]Check) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		status, body := runChecks(ctx, checks)
		writeJSON(w, status, body)
	})
	return mux
}

func runChecks(ctx context.Context, checks map[string]Check) (int, map[string]any) {
	names := make([]string, 0, len(checks))
	for name := range checks {
		names = append(names, name)
	}
	sort.Strings(names)
	results := map[string]string{}
	status := http.StatusOK
	for _, name := range names {
		if err := checks[name](ctx); err != nil {
			// Report only that it failed; error text can leak hostnames or DSNs.
			results[name] = "fail"
			status = http.StatusServiceUnavailable
			continue
		}
		results[name] = "ok"
	}
	ready := "ok"
	if status != http.StatusOK {
		ready = "fail"
	}
	return status, map[string]any{"status": ready, "checks": results}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// Serve starts the internal endpoint and stops it when ctx is done.
// A listen failure is logged, never fatal: losing metrics must not stop the service.
func Serve(ctx context.Context, addr string, reg *prometheus.Registry, checks map[string]Check, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           Handler(reg, checks),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("observability endpoint not started", "addr", addr, "error", err)
		return
	}
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("observability endpoint stopped", "error", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Info("observability endpoint started", "addr", addr)
}

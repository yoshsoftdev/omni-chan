package observability

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMetrics are RED metrics (rate, errors, duration) per route template.
type HTTPMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewHTTPMetrics registers http_requests_total and http_request_duration_seconds.
func NewHTTPMetrics(reg prometheus.Registerer) *HTTPMetrics {
	m := &HTTPMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by method, route template and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency by method and route template.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "route"}),
	}
	reg.MustRegister(m.requests, m.duration)
	return m
}

// FiberOptions configures FiberMiddleware.
type FiberOptions struct {
	Metrics *HTTPMetrics
	Log     *slog.Logger
	// TeamID optionally returns the authenticated tenant for the access log.
	TeamID func(c *fiber.Ctx) string
}

// FiberMiddleware traces each request (continuing an incoming traceparent),
// records RED metrics, and writes one JSON access log line per request.
// Only the path is logged, never the query string: /ws?token=... carries a credential.
func FiberMiddleware(opts FiberOptions) fiber.Handler {
	log := opts.Log
	if log == nil {
		log = slog.Default()
	}
	return func(c *fiber.Ctx) error {
		headers := map[string][]string{}
		c.Request().Header.VisitAll(func(k, v []byte) { headers[string(k)] = []string{string(v)} })
		ctx, span := Tracer().Start(Extract(c.UserContext(), headers), c.Method(), trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		c.SetUserContext(ctx)

		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start)

		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
			}
		}
		route := c.Route().Path
		if status == fiber.StatusNotFound {
			route = "unmatched" // keeps label cardinality bounded
		}

		span.SetName(c.Method() + " " + route)
		span.SetAttributes(
			attribute.String("http.request.method", c.Method()),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
		)
		if status >= 500 {
			span.SetStatus(codes.Error, strconv.Itoa(status))
		}

		if opts.Metrics != nil {
			opts.Metrics.requests.WithLabelValues(c.Method(), route, strconv.Itoa(status)).Inc()
			opts.Metrics.duration.WithLabelValues(c.Method(), route).Observe(elapsed.Seconds())
		}

		if route != "/health" && route != "/api/v1/health" {
			attrs := []slog.Attr{
				slog.String("method", c.Method()),
				slog.String("route", route),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.Int64("duration_ms", elapsed.Milliseconds()),
			}
			if rid := c.GetRespHeader(fiber.HeaderXRequestID); rid != "" {
				attrs = append(attrs, slog.String("request_id", rid))
			}
			if opts.TeamID != nil {
				if team := opts.TeamID(c); team != "" {
					attrs = append(attrs, slog.String("team_id", team))
				}
			}
			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			}
			log.LogAttrs(ctx, level, "http request", attrs...)
		}
		return err
	}
}

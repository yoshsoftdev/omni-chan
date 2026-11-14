package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/omnichannel/omni-chan-v2/shared/observability"

// SetupTracing installs the global tracer provider and W3C trace-context
// propagator. Spans always get real trace IDs (so logs can be correlated);
// they are exported only when OTEL_EXPORTER_OTLP_ENDPOINT (or
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT) is set, e.g. http://tempo:4318.
// The returned func flushes pending spans; call it on shutdown.
func SetupTracing(ctx context.Context, service string, log *slog.Logger) func(context.Context) error {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	res, _ := resource.Merge(resource.Default(), resource.NewSchemaless(attribute.String("service.name", service)))
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" {
		exporter, err := otlptracehttp.New(ctx)
		if err != nil {
			if log != nil {
				log.Error("trace export disabled: exporter setup failed", "error", err)
			}
		} else {
			opts = append(opts, sdktrace.WithBatcher(exporter))
		}
	}
	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

// Tracer is the shared tracer for spans created by this repo's code.
func Tracer() trace.Tracer { return otel.Tracer(tracerName) }

// MapCarrier adapts a header map (e.g. nats.Header) for propagation.
type MapCarrier map[string][]string

// Get matches keys case-insensitively: HTTP stacks canonicalize "traceparent"
// to "Traceparent", while NATS headers keep whatever case was set.
func (c MapCarrier) Get(key string) string {
	if v := c[key]; len(v) > 0 {
		return v[0]
	}
	for k, v := range c {
		if len(v) > 0 && strings.EqualFold(k, key) {
			return v[0]
		}
	}
	return ""
}
func (c MapCarrier) Set(key, value string) { c[key] = []string{value} }
func (c MapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// Inject writes the current trace context into headers.
func Inject(ctx context.Context, headers map[string][]string) {
	otel.GetTextMapPropagator().Inject(ctx, MapCarrier(headers))
}

// Extract returns ctx continued from trace context found in headers.
func Extract(ctx context.Context, headers map[string][]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, MapCarrier(headers))
}

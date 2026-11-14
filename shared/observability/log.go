package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// logHandler adds trace_id and span_id to records logged with a context that
// carries a span (slog.InfoContext and friends), so logs join up with traces.
type logHandler struct{ slog.Handler }

// WrapLogHandler returns h with trace correlation fields added.
func WrapLogHandler(h slog.Handler) slog.Handler { return logHandler{h} }

func (h logHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

func (h logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return logHandler{h.Handler.WithAttrs(attrs)}
}

func (h logHandler) WithGroup(name string) slog.Handler { return logHandler{h.Handler.WithGroup(name)} }

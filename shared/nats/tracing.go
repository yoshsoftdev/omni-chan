package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/omnichannel/omni-chan-v2/shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ConsumerSpan continues the publisher's trace (traceparent header) for one message.
func ConsumerSpan(ctx context.Context, msg *nats.Msg) (context.Context, trace.Span) {
	if msg.Header != nil {
		ctx = observability.Extract(ctx, msg.Header)
	}
	return observability.Tracer().Start(ctx, "consume "+msg.Subject,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attribute.String("messaging.destination.name", msg.Subject)))
}

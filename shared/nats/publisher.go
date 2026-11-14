package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/omnichannel/omni-chan-v2/shared/observability"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

type Publisher struct {
	js            nats.JetStreamContext
	cfg           sharedconfig.NATS
	log           *slog.Logger
	internalToken string
}

func NewPublisher(conn *nats.Conn, cfg sharedconfig.NATS, log *slog.Logger) (*Publisher, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	p := &Publisher{js: js, cfg: cfg, log: log, internalToken: cfg.InternalToken}
	if err := p.ensureStream(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Publisher) ensureStream() error {
	return EnsureStream(p.js, WebhookStreamSpec(p.cfg), p.log)
}

func (p *Publisher) Publish(ctx context.Context, eventType, source string, payload any, metadata map[string]string) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	envelope := sharedevents.Envelope{
		ID:        uuid.NewString(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Payload:   raw,
		Metadata:  metadata,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	subject := sharedevents.Subject(eventType)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	if p.internalToken != "" {
		msg.Header.Set("X-Internal-Token", p.internalToken)
	}
	observability.Inject(ctx, msg.Header)
	// Server-side dedup: a retry of a publish that actually landed is dropped.
	msg.Header.Set(nats.MsgIdHdr, envelope.ID)

	var lastErr error
	retries := p.cfg.MaxPublishRetry
	if retries < 1 {
		retries = 1
	}
	for attempt := 1; attempt <= retries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, lastErr = p.js.PublishMsg(msg)
		if lastErr == nil {
			return nil
		}
		p.log.Warn("nats publish retry", "subject", subject, "attempt", attempt, "error", lastErr)
		time.Sleep(p.cfg.PublishRetryWait * time.Duration(attempt))
	}
	return fmt.Errorf("publish %s after %d attempts: %w", subject, retries, lastErr)
}

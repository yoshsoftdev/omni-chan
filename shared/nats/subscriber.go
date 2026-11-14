package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

type HandlerFunc func(ctx context.Context, envelope sharedevents.Envelope) error

type Subscriber struct {
	js            nats.JetStreamContext
	cfg           sharedconfig.NATS
	log           *slog.Logger
	internalToken string
	policy        RedeliveryPolicy
}

func NewSubscriber(conn *nats.Conn, cfg sharedconfig.NATS, log *slog.Logger) (*Subscriber, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	s := &Subscriber{js: js, cfg: cfg, log: log, internalToken: cfg.InternalToken, policy: RedeliveryPolicyFrom(cfg)}
	if err := s.ensureStream(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Subscriber) ensureStream() error {
	return EnsureStream(s.js, WebhookStreamSpec(s.cfg), s.log)
}

func (s *Subscriber) Subscribe(ctx context.Context, consumerName string, handler HandlerFunc) error {
	subject := "events.>"
	sub, err := s.js.PullSubscribe(subject, consumerName, nats.BindStream(s.cfg.StreamName))
	if err != nil {
		return fmt.Errorf("pull subscribe: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil && !errors.Is(err, nats.ErrTimeout) {
				s.log.Error("fetch messages", "error", err)
				sleepCtx(ctx, fetchErrorBackoff)
				continue
			}
			for _, msg := range msgs {
				s.handleMessage(ctx, msg, handler)
			}
		}
	}()
	s.log.Info("nats subscriber started", "consumer", consumerName, "subject", subject)
	return nil
}

func (s *Subscriber) handleMessage(ctx context.Context, msg *nats.Msg, handler HandlerFunc) {
	if s.internalToken != "" {
		token := msg.Header.Get("X-Internal-Token")
		if token != s.internalToken {
			// Bounded: retried with backoff (covers token rotation), then terminated.
			delivered, terminated := s.policy.retryOrTerm(msg)
			s.log.Warn("rejected event: invalid internal token",
				"subject", msg.Subject, "attempt", delivered, "terminated", terminated)
			return
		}
	}

	var envelope sharedevents.Envelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		s.log.Error("invalid envelope", "error", err)
		_ = msg.Term()
		return
	}

	ctx, span := ConsumerSpan(ctx, msg)
	defer span.End()
	if err := handler(ctx, envelope); err != nil {
		span.RecordError(err)
		delivered, terminated := s.policy.retryOrTerm(msg)
		if terminated {
			s.log.ErrorContext(ctx, "handler terminal failure", "type", envelope.Type, "event_id", envelope.ID, "attempt", delivered, "error", err)
			return
		}
		s.log.WarnContext(ctx, "handler retry", "type", envelope.Type, "event_id", envelope.ID, "attempt", delivered, "error", err)
		return
	}
	_ = msg.Ack()
}

func Connect(url string) (*nats.Conn, error) {
	return nats.Connect(url,
		nats.Name("omni-chan"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
}

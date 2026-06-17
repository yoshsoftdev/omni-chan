package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

type HandlerFunc func(ctx context.Context, envelope sharedevents.Envelope) error

type Subscriber struct {
	js            nats.JetStreamContext
	cfg           sharedconfig.NATS
	log           *slog.Logger
	internalToken string
}

func NewSubscriber(conn *nats.Conn, cfg sharedconfig.NATS, log *slog.Logger) (*Subscriber, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	s := &Subscriber{js: js, cfg: cfg, log: log, internalToken: cfg.InternalToken}
	if err := s.ensureStream(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Subscriber) ensureStream() error {
	_, err := s.js.StreamInfo(s.cfg.StreamName)
	if err == nil {
		return nil
	}
	_, err = s.js.AddStream(&nats.StreamConfig{
		Name:      s.cfg.StreamName,
		Subjects:  []string{"events.>"},
		Retention: nats.LimitsPolicy,
		Storage:   nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", s.cfg.StreamName, err)
	}
	s.log.Info("created jetstream", "stream", s.cfg.StreamName)
	return nil
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
			if err != nil && err != nats.ErrTimeout {
				s.log.Error("fetch messages", "error", err)
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
			s.log.Warn("dropped event: invalid internal token", "subject", msg.Subject)
			_ = msg.Nak()
			return
		}
	}

	var envelope sharedevents.Envelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		s.log.Error("invalid envelope", "error", err)
		_ = msg.Term()
		return
	}

	if err := handler(ctx, envelope); err != nil {
		s.log.Error("handler failed", "type", envelope.Type, "error", err)
		_ = msg.Nak()
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

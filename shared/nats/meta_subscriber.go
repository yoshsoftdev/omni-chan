package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

// CanonicalHandler processes canonical Meta events from JetStream.
type CanonicalHandler func(ctx context.Context, event sharedevents.CanonicalEvent) error

// MetaSubscriber consumes meta.* canonical events.
type MetaSubscriber struct {
	js            nats.JetStreamContext
	cfg           sharedconfig.NATS
	log           *slog.Logger
	internalToken string
	maxDeliver    int
}

func NewMetaSubscriber(conn *nats.Conn, cfg sharedconfig.NATS, log *slog.Logger) (*MetaSubscriber, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	s := &MetaSubscriber{
		js:            js,
		cfg:           cfg,
		log:           log,
		internalToken: cfg.InternalToken,
		maxDeliver:    5,
	}
	if err := s.ensureStream(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *MetaSubscriber) ensureStream() error {
	info, err := s.js.StreamInfo(s.cfg.StreamName)
	if err != nil {
		_, err = s.js.AddStream(&nats.StreamConfig{
			Name:      s.cfg.StreamName,
			Subjects:  []string{"events.>", "meta.>"},
			Retention: nats.LimitsPolicy,
			Storage:   nats.FileStorage,
		})
		if err != nil {
			return fmt.Errorf("create stream %s: %w", s.cfg.StreamName, err)
		}
		return nil
	}
	hasMeta := false
	for _, subj := range info.Config.Subjects {
		if subj == "meta.>" || strings.HasPrefix(subj, "meta.") {
			hasMeta = true
			break
		}
	}
	if !hasMeta {
		subjects := append(info.Config.Subjects, "meta.>")
		_, err = s.js.UpdateStream(&nats.StreamConfig{
			Name:      info.Config.Name,
			Subjects:  subjects,
			Retention: info.Config.Retention,
			Storage:   info.Config.Storage,
		})
		if err != nil {
			return fmt.Errorf("update stream subjects: %w", err)
		}
	}
	return nil
}

func (s *MetaSubscriber) Subscribe(ctx context.Context, consumerName string, handler CanonicalHandler) error {
	filter := "meta.>"
	sub, err := s.js.PullSubscribe(filter, consumerName, nats.BindStream(s.cfg.StreamName))
	if err != nil {
		return fmt.Errorf("pull subscribe meta: %w", err)
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
				s.log.Error("meta fetch messages", "error", err)
				continue
			}
			for _, msg := range msgs {
				s.handleMessage(ctx, msg, handler)
			}
		}
	}()
	s.log.Info("meta subscriber started", "consumer", consumerName, "subject", filter)
	return nil
}

func (s *MetaSubscriber) handleMessage(ctx context.Context, msg *nats.Msg, handler CanonicalHandler) {
	if s.shouldSkipSubject(msg.Subject) {
		_ = msg.Ack()
		return
	}
	if s.internalToken != "" {
		token := msg.Header.Get("X-Internal-Token")
		if token != s.internalToken {
			s.log.Warn("dropped meta event: invalid internal token", "subject", msg.Subject)
			_ = msg.Nak()
			return
		}
	}

	var event sharedevents.CanonicalEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.log.Error("invalid canonical event", "error", err, "subject", msg.Subject)
		_ = msg.Term()
		return
	}

	if err := handler(ctx, event); err != nil {
		meta, metaErr := msg.Metadata()
		deliveries := 1
		if metaErr == nil && meta != nil {
			deliveries = int(meta.NumDelivered)
		}
		if deliveries >= s.maxDeliver {
			s.log.Error("meta handler terminal failure", "type", event.Type, "event_id", event.EventID, "error", err)
			_ = msg.Term()
			return
		}
		s.log.Warn("meta handler retry", "type", event.Type, "event_id", event.EventID, "attempt", deliveries, "error", err)
		_ = msg.Nak()
		return
	}
	_ = msg.Ack()
}

func (s *MetaSubscriber) shouldSkipSubject(subject string) bool {
	return strings.HasPrefix(subject, "meta.dlq.") || strings.HasPrefix(subject, "meta.consumer.dlq.")
}

// PublishConsumerDLQ publishes a failed consumer event to meta.consumer.dlq.*.
func (s *MetaSubscriber) PublishConsumerDLQ(ctx context.Context, category string, errMsg string, event sharedevents.CanonicalEvent) error {
	payload, err := json.Marshal(struct {
		Category string                      `json:"category"`
		Error    string                      `json:"error"`
		Event    sharedevents.CanonicalEvent `json:"event"`
		At       time.Time                   `json:"at"`
	}{
		Category: category,
		Error:    errMsg,
		Event:    event,
		At:       time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	subject := sharedevents.MetaConsumerDLQSubject(category)
	msg := &nats.Msg{Subject: subject, Data: payload, Header: nats.Header{}}
	if s.internalToken != "" {
		msg.Header.Set("X-Internal-Token", s.internalToken)
	}
	_, err = s.js.PublishMsg(msg)
	return err
}

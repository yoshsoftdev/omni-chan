package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

// BroadcastPublisher publishes broadcast.* subjects to a dedicated JetStream stream.
type BroadcastPublisher struct {
	js            nats.JetStreamContext
	streamName    string
	log           *slog.Logger
	internalToken string
	maxRetry      int
	retryWait     time.Duration
}

func NewBroadcastPublisher(conn *nats.Conn, cfg sharedconfig.NATS, streamName string, log *slog.Logger) (*BroadcastPublisher, error) {
	if streamName == "" {
		streamName = "BROADCAST"
	}
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	p := &BroadcastPublisher{
		js:            js,
		streamName:    streamName,
		log:           log,
		internalToken: cfg.InternalToken,
		maxRetry:      cfg.MaxPublishRetry,
		retryWait:     cfg.PublishRetryWait,
	}
	if p.maxRetry < 1 {
		p.maxRetry = 3
	}
	if p.retryWait <= 0 {
		p.retryWait = 200 * time.Millisecond
	}
	if err := p.ensureStream(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *BroadcastPublisher) ensureStream() error {
	_, err := p.js.StreamInfo(p.streamName)
	if err == nil {
		return nil
	}
	_, err = p.js.AddStream(&nats.StreamConfig{
		Name:      p.streamName,
		Subjects:  []string{"broadcast.>"},
		Retention: nats.LimitsPolicy,
		Storage:   nats.FileStorage,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", p.streamName, err)
	}
	p.log.Info("created jetstream", "stream", p.streamName)
	return nil
}

// PublishJSON publishes a JSON payload to a broadcast subject.
func (p *BroadcastPublisher) PublishJSON(ctx context.Context, subject string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	msg := &nats.Msg{Subject: subject, Data: raw, Header: nats.Header{}}
	if p.internalToken != "" {
		msg.Header.Set("X-Internal-Token", p.internalToken)
	}
	var lastErr error
	for attempt := 1; attempt <= p.maxRetry; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, lastErr = p.js.PublishMsg(msg)
		if lastErr == nil {
			return nil
		}
		p.log.Warn("broadcast publish retry", "subject", subject, "attempt", attempt, "error", lastErr)
		time.Sleep(p.retryWait * time.Duration(attempt))
	}
	return fmt.Errorf("publish %s after %d attempts: %w", subject, p.maxRetry, lastErr)
}

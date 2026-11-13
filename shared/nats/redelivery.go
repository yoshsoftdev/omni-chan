package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

// Redelivery defaults used when config leaves a field at zero.
const (
	DefaultMaxDeliver     = 5
	DefaultRetryBaseDelay = 1 * time.Second
	DefaultRetryMaxDelay  = 30 * time.Second
	// fetchErrorBackoff pauses the pull loop after a non-timeout Fetch error
	// (e.g. disconnected) so it does not spin and flood logs.
	fetchErrorBackoff = 1 * time.Second
)

// RedeliveryPolicy decides how a failed message is retried.
type RedeliveryPolicy struct {
	MaxDeliver int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// RedeliveryPolicyFrom applies package defaults to zero config fields.
func RedeliveryPolicyFrom(cfg sharedconfig.NATS) RedeliveryPolicy {
	p := RedeliveryPolicy{
		MaxDeliver: cfg.ConsumerMaxDeliver,
		BaseDelay:  cfg.RetryBaseDelay,
		MaxDelay:   cfg.RetryMaxDelay,
	}
	if p.MaxDeliver <= 0 {
		p.MaxDeliver = DefaultMaxDeliver
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = DefaultRetryBaseDelay
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = DefaultRetryMaxDelay
	}
	if p.MaxDelay < p.BaseDelay {
		p.MaxDelay = p.BaseDelay
	}
	return p
}

// Exhausted reports whether a message on its delivered-th attempt must be terminated.
func (p RedeliveryPolicy) Exhausted(delivered int) bool {
	return delivered >= p.MaxDeliver
}

// Delay returns the backoff before redelivery attempt delivered+1:
// BaseDelay * 2^(delivered-1), capped at MaxDelay.
func (p RedeliveryPolicy) Delay(delivered int) time.Duration {
	if delivered < 1 {
		delivered = 1
	}
	d := p.BaseDelay
	for i := 1; i < delivered; i++ {
		d *= 2
		if d >= p.MaxDelay {
			return p.MaxDelay
		}
	}
	if d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}

// ackable is the subset of *nats.Msg used to settle a message.
type ackable interface {
	Ack(opts ...nats.AckOpt) error
	Term(opts ...nats.AckOpt) error
	NakWithDelay(delay time.Duration, opts ...nats.AckOpt) error
	Metadata() (*nats.MsgMetadata, error)
}

// deliveryCount returns how many times msg has been delivered (>= 1).
func deliveryCount(msg ackable) int {
	meta, err := msg.Metadata()
	if err != nil || meta == nil || meta.NumDelivered < 1 {
		return 1
	}
	return int(meta.NumDelivered)
}

// retryOrTerm settles a failed message: delayed NAK while attempts remain,
// TERM once exhausted. Returns the delivery count and whether it was terminated.
func (p RedeliveryPolicy) retryOrTerm(msg ackable) (delivered int, terminated bool) {
	delivered = deliveryCount(msg)
	if p.Exhausted(delivered) {
		_ = msg.Term()
		return delivered, true
	}
	_ = msg.NakWithDelay(p.Delay(delivered))
	return delivered, false
}

// sleepCtx waits d or until ctx is done.
func sleepCtx(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// NakWithBackoff NAKs msg with the policy delay for its delivery count and never
// terminates it. Use it where an external recovery process owns final failure
// (e.g. broadcast batches tracked in Postgres). Returns the delivery count.
func (p RedeliveryPolicy) NakWithBackoff(msg ackable) int {
	delivered := deliveryCount(msg)
	_ = msg.NakWithDelay(p.Delay(delivered))
	return delivered
}

// PauseAfterFetchError waits briefly after a non-timeout Fetch error so pull
// loops do not spin while the server is unreachable. Returns early on ctx done.
func PauseAfterFetchError(ctx context.Context) {
	sleepCtx(ctx, fetchErrorBackoff)
}

package nats

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

func TestRedeliveryPolicyDefaults(t *testing.T) {
	p := RedeliveryPolicyFrom(sharedconfig.NATS{})
	if p.MaxDeliver != DefaultMaxDeliver || p.BaseDelay != DefaultRetryBaseDelay || p.MaxDelay != DefaultRetryMaxDelay {
		t.Fatalf("defaults not applied: %+v", p)
	}
}

func TestRedeliveryPolicyDelayBackoffIsCapped(t *testing.T) {
	p := RedeliveryPolicy{MaxDeliver: 10, BaseDelay: time.Second, MaxDelay: 5 * time.Second}
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}
	for i, w := range want {
		if got := p.Delay(i + 1); got != w {
			t.Errorf("Delay(%d) = %v, want %v", i+1, got, w)
		}
	}
	if got := p.Delay(0); got != time.Second {
		t.Errorf("Delay(0) = %v, want base", got)
	}
	if got := p.Delay(1000); got != 5*time.Second {
		t.Errorf("Delay(1000) = %v, want cap (no overflow)", got)
	}
}

func TestRedeliveryPolicyExhausted(t *testing.T) {
	p := RedeliveryPolicy{MaxDeliver: 3}
	for delivered, want := range map[int]bool{1: false, 2: false, 3: true, 4: true} {
		if got := p.Exhausted(delivered); got != want {
			t.Errorf("Exhausted(%d) = %v, want %v", delivered, got, want)
		}
	}
}

func TestRedeliveryPolicyMaxBelowBase(t *testing.T) {
	p := RedeliveryPolicyFrom(sharedconfig.NATS{RetryBaseDelay: 10 * time.Second, RetryMaxDelay: time.Second})
	if p.MaxDelay != p.BaseDelay {
		t.Fatalf("MaxDelay %v should be raised to BaseDelay %v", p.MaxDelay, p.BaseDelay)
	}
}

type fakeMsg struct {
	delivered uint64
	nakDelay  time.Duration
	naks      int
	terms     int
}

func (f *fakeMsg) Ack(...nats.AckOpt) error  { return nil }
func (f *fakeMsg) Term(...nats.AckOpt) error { f.terms++; return nil }
func (f *fakeMsg) NakWithDelay(d time.Duration, _ ...nats.AckOpt) error {
	f.naks++
	f.nakDelay = d
	return nil
}
func (f *fakeMsg) Metadata() (*nats.MsgMetadata, error) {
	return &nats.MsgMetadata{NumDelivered: f.delivered}, nil
}

func TestNakWithBackoffNeverTerminates(t *testing.T) {
	p := RedeliveryPolicy{MaxDeliver: 2, BaseDelay: time.Second, MaxDelay: 4 * time.Second}
	m := &fakeMsg{delivered: 50}
	if got := p.NakWithBackoff(m); got != 50 {
		t.Fatalf("delivered = %d", got)
	}
	if m.terms != 0 || m.naks != 1 || m.nakDelay != 4*time.Second {
		t.Fatalf("terms=%d naks=%d delay=%v, want 0/1/4s", m.terms, m.naks, m.nakDelay)
	}
}

func TestRetryOrTermTerminatesWhenExhausted(t *testing.T) {
	p := RedeliveryPolicy{MaxDeliver: 3, BaseDelay: time.Second, MaxDelay: time.Minute}
	early := &fakeMsg{delivered: 2}
	if _, term := p.retryOrTerm(early); term || early.naks != 1 || early.nakDelay != 2*time.Second {
		t.Fatalf("attempt 2: term=%v naks=%d delay=%v", term, early.naks, early.nakDelay)
	}
	last := &fakeMsg{delivered: 3}
	if _, term := p.retryOrTerm(last); !term || last.terms != 1 || last.naks != 0 {
		t.Fatalf("attempt 3: term=%v terms=%d naks=%d", term, last.terms, last.naks)
	}
}

package nats

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// startJetStream runs an in-process nats-server with JetStream on a random port.
func startJetStream(t *testing.T) *nats.Conn {
	t.Helper()
	srv, err := natsserver.NewServer(&natsserver.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
		NoLog:     true,
		NoSigs:    true,
	})
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	go srv.Start()
	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("nats server not ready")
	}
	t.Cleanup(srv.Shutdown)

	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(nc.Close)
	return nc
}

func testNATSConfig(stream string) sharedconfig.NATS {
	return sharedconfig.NATS{
		StreamName:         stream,
		InternalToken:      "internal-secret",
		MaxPublishRetry:    1,
		PublishRetryWait:   time.Millisecond,
		ConsumerMaxDeliver: 3,
		RetryBaseDelay:     20 * time.Millisecond,
		RetryMaxDelay:      40 * time.Millisecond,
	}
}

func TestEnsureStream_CreatesWithLimits(t *testing.T) {
	nc := startJetStream(t)
	js, _ := nc.JetStream()
	cfg := testNATSConfig("WEBHOOK_EVENTS")

	if err := EnsureStream(js, WebhookStreamSpec(cfg), quietLog()); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo("WEBHOOK_EVENTS")
	if err != nil {
		t.Fatal(err)
	}
	if info.Config.MaxAge != DefaultStreamMaxAge {
		t.Errorf("MaxAge = %v, want %v", info.Config.MaxAge, DefaultStreamMaxAge)
	}
	if info.Config.Duplicates != DefaultDuplicateWindow {
		t.Errorf("Duplicates = %v", info.Config.Duplicates)
	}
	for _, want := range WebhookStreamSubjects {
		found := false
		for _, s := range info.Config.Subjects {
			found = found || s == want
		}
		if !found {
			t.Errorf("subject %q missing: %v", want, info.Config.Subjects)
		}
	}
}

// An existing pre-hardening stream (events.> only, no limits, custom description)
// is upgraded in place without losing operator settings or messages.
func TestEnsureStream_ReconcilesLegacyStreamWithoutClobbering(t *testing.T) {
	nc := startJetStream(t)
	js, _ := nc.JetStream()
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:        "WEBHOOK_EVENTS",
		Subjects:    []string{"events.>"},
		Retention:   nats.LimitsPolicy,
		Storage:     nats.FileStorage,
		Description: "operator note",
		MaxMsgs:     12345,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("events.message.received", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	cfg := testNATSConfig("WEBHOOK_EVENTS")
	if err := EnsureStream(js, WebhookStreamSpec(cfg), quietLog()); err != nil {
		t.Fatal(err)
	}
	info, _ := js.StreamInfo("WEBHOOK_EVENTS")
	if info.Config.Description != "operator note" || info.Config.MaxMsgs != 12345 {
		t.Errorf("operator config clobbered: %+v", info.Config)
	}
	if info.Config.MaxAge != DefaultStreamMaxAge {
		t.Errorf("MaxAge not filled in: %v", info.Config.MaxAge)
	}
	if len(info.Config.Subjects) != 2 {
		t.Errorf("subjects = %v, want events.> and meta.>", info.Config.Subjects)
	}
	if info.State.Msgs != 1 {
		t.Errorf("existing messages lost: %d", info.State.Msgs)
	}
}

// An operator-chosen MaxAge is never overwritten by the default.
func TestEnsureStream_KeepsOperatorMaxAge(t *testing.T) {
	nc := startJetStream(t)
	js, _ := nc.JetStream()
	if _, err := js.AddStream(&nats.StreamConfig{
		Name: "WEBHOOK_EVENTS", Subjects: WebhookStreamSubjects, MaxAge: time.Hour,
	}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, WebhookStreamSpec(testNATSConfig("WEBHOOK_EVENTS")), quietLog()); err != nil {
		t.Fatal(err)
	}
	info, _ := js.StreamInfo("WEBHOOK_EVENTS")
	if info.Config.MaxAge != time.Hour {
		t.Errorf("MaxAge = %v, want operator value 1h", info.Config.MaxAge)
	}
}

type failingStreamAPI struct{ addCalls int }

func (f *failingStreamAPI) StreamInfo(string, ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nats.ErrTimeout
}
func (f *failingStreamAPI) AddStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error) {
	f.addCalls++
	return nil, nil
}
func (f *failingStreamAPI) UpdateStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

// A transient StreamInfo failure must surface, not be mistaken for "missing".
func TestEnsureStream_TransientErrorDoesNotCreate(t *testing.T) {
	api := &failingStreamAPI{}
	err := EnsureStream(api, StreamSpec{Name: "WEBHOOK_EVENTS", Subjects: WebhookStreamSubjects}, quietLog())
	if !errors.Is(err, nats.ErrTimeout) {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if api.addCalls != 0 {
		t.Fatal("AddStream must not be called on a transient error")
	}
}

func TestPublisher_SetsMsgIDForDedup(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	pub, err := NewPublisher(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	if err := pub.Publish(context.Background(), "message.received", "test", map[string]string{"k": "v"}, nil); err != nil {
		t.Fatal(err)
	}
	js, _ := nc.JetStream()
	sub, err := js.PullSubscribe("events.>", "probe", nats.BindStream("WEBHOOK_EVENTS"))
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
	if err != nil || len(msgs) != 1 {
		t.Fatalf("fetch: %v (%d msgs)", err, len(msgs))
	}
	var env sharedevents.Envelope
	if err := json.Unmarshal(msgs[0].Data, &env); err != nil {
		t.Fatal(err)
	}
	if got := msgs[0].Header.Get(nats.MsgIdHdr); got == "" || got != env.ID {
		t.Fatalf("Nats-Msg-Id = %q, want envelope id %q", got, env.ID)
	}
	// Re-publishing the same id inside the window is dropped by the server.
	dup := &nats.Msg{Subject: msgs[0].Subject, Data: msgs[0].Data, Header: nats.Header{}}
	dup.Header.Set(nats.MsgIdHdr, env.ID)
	ack, err := js.PublishMsg(dup)
	if err != nil {
		t.Fatal(err)
	}
	if !ack.Duplicate {
		t.Fatal("expected server to flag duplicate publish")
	}
}

func publishEnvelope(t *testing.T, nc *nats.Conn, token string) {
	t.Helper()
	js, _ := nc.JetStream()
	data, _ := json.Marshal(sharedevents.Envelope{ID: "env-1", Type: "message.received", Timestamp: time.Now().UTC()})
	msg := &nats.Msg{Subject: "events.message.received", Data: data, Header: nats.Header{}}
	msg.Header.Set("X-Internal-Token", token)
	if _, err := js.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}
}

// waitStable waits until the counter stops changing for quiet, or timeout.
func waitStable(counter *atomic.Int64, quiet, timeout time.Duration) int64 {
	deadline := time.Now().Add(timeout)
	last := counter.Load()
	stableSince := time.Now()
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		if cur := counter.Load(); cur != last {
			last, stableSince = cur, time.Now()
		} else if time.Since(stableSince) >= quiet {
			return last
		}
	}
	return counter.Load()
}

// A permanently failing handler is attempted exactly MaxDeliver times, then terminated.
func TestSubscriber_FailingHandlerIsBounded(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	sub, err := NewSubscriber(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sub.Subscribe(ctx, "legacy-test", func(context.Context, sharedevents.Envelope) error {
		calls.Add(1)
		return errors.New("db down")
	}); err != nil {
		t.Fatal(err)
	}
	publishEnvelope(t, nc, cfg.InternalToken)

	got := waitStable(&calls, 500*time.Millisecond, 10*time.Second)
	if got != int64(cfg.ConsumerMaxDeliver) {
		t.Fatalf("handler calls = %d, want exactly MaxDeliver=%d", got, cfg.ConsumerMaxDeliver)
	}
}

// A transient failure is retried with backoff and then acknowledged once.
func TestSubscriber_TransientFailureRecovers(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	sub, err := NewSubscriber(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int64
	var mu sync.Mutex
	var times []time.Time
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sub.Subscribe(ctx, "legacy-recover", func(context.Context, sharedevents.Envelope) error {
		mu.Lock()
		times = append(times, time.Now())
		mu.Unlock()
		if calls.Add(1) == 1 {
			return errors.New("transient")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	publishEnvelope(t, nc, cfg.InternalToken)

	if got := waitStable(&calls, 400*time.Millisecond, 10*time.Second); got != 2 {
		t.Fatalf("handler calls = %d, want 2 (fail then succeed, no redelivery after ack)", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if gap := times[1].Sub(times[0]); gap < cfg.RetryBaseDelay {
		t.Fatalf("retry came after %v, want >= base delay %v (no hot loop)", gap, cfg.RetryBaseDelay)
	}
}

// A message with a wrong internal token never reaches the handler and is
// terminated after MaxDeliver attempts instead of looping forever.
func TestMetaSubscriber_InvalidTokenIsBoundedAndNeverHandled(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	sub, err := NewMetaSubscriber(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	var handled atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sub.Subscribe(ctx, "meta-token", func(context.Context, sharedevents.CanonicalEvent) error {
		handled.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	js, _ := nc.JetStream()
	data, _ := json.Marshal(sharedevents.CanonicalEvent{EventID: "e1", Type: "message.received"})
	msg := &nats.Msg{Subject: "meta.message.received", Data: data, Header: nats.Header{}}
	msg.Header.Set("X-Internal-Token", "forged")
	if _, err := js.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}

	// Let the policy run to completion, then confirm nothing is pending.
	time.Sleep(time.Second)
	if handled.Load() != 0 {
		t.Fatal("handler must never see a message with an invalid token")
	}
	ci, err := js.ConsumerInfo(cfg.StreamName, "meta-token")
	if err != nil {
		t.Fatal(err)
	}
	if ci.NumAckPending != 0 || ci.NumPending != 0 {
		t.Fatalf("message still in flight: ack_pending=%d pending=%d", ci.NumAckPending, ci.NumPending)
	}
	if int(ci.Delivered.Consumer) != cfg.ConsumerMaxDeliver {
		t.Fatalf("delivered %d times, want MaxDeliver=%d", ci.Delivered.Consumer, cfg.ConsumerMaxDeliver)
	}
}

func TestMetaSubscriber_ValidEventHandledOnce(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	sub, err := NewMetaSubscriber(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	var handled atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sub.Subscribe(ctx, "meta-ok", func(_ context.Context, ev sharedevents.CanonicalEvent) error {
		if ev.EventID != "e2" {
			t.Errorf("event id = %q", ev.EventID)
		}
		handled.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	js, _ := nc.JetStream()
	data, _ := json.Marshal(sharedevents.CanonicalEvent{EventID: "e2", Type: "message.received"})
	msg := &nats.Msg{Subject: "meta.message.received", Data: data, Header: nats.Header{}}
	msg.Header.Set("X-Internal-Token", cfg.InternalToken)
	if _, err := js.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}
	if got := waitStable(&handled, 400*time.Millisecond, 10*time.Second); got != 1 {
		t.Fatalf("handled = %d, want 1", got)
	}
}

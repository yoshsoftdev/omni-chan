package nats

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
	"github.com/omnichannel/omni-chan-v2/shared/observability"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
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

// A retention set in config (not the default) is applied to an existing stream, so it
// can change without recreating the stream (ADR 0009).
func TestEnsureStream_EnforcesConfiguredMaxAge(t *testing.T) {
	nc := startJetStream(t)
	js, _ := nc.JetStream()
	if _, err := js.AddStream(&nats.StreamConfig{
		Name: "WEBHOOK_EVENTS", Subjects: WebhookStreamSubjects, MaxAge: time.Hour,
	}); err != nil {
		t.Fatal(err)
	}
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	cfg.StreamMaxAge = 30 * 24 * time.Hour
	if err := EnsureStream(js, WebhookStreamSpec(cfg), quietLog()); err != nil {
		t.Fatal(err)
	}
	info, _ := js.StreamInfo("WEBHOOK_EVENTS")
	if info.Config.MaxAge != cfg.StreamMaxAge {
		t.Errorf("MaxAge = %v, want configured %v", info.Config.MaxAge, cfg.StreamMaxAge)
	}
}

// The broadcast stream goes through EnsureStream: created with its own retention,
// and a transient error is not mistaken for "missing" (it used to create on any error).
func TestBroadcastPublisher_EnsuresStreamFromConfig(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	cfg.BroadcastStreamMaxAge = 3 * 24 * time.Hour
	if _, err := NewBroadcastPublisher(nc, cfg, "BROADCAST", quietLog()); err != nil {
		t.Fatal(err)
	}
	js, _ := nc.JetStream()
	info, err := js.StreamInfo("BROADCAST")
	if err != nil {
		t.Fatal(err)
	}
	if info.Config.MaxAge != cfg.BroadcastStreamMaxAge || info.Config.Subjects[0] != "broadcast.>" {
		t.Errorf("broadcast stream config = %+v", info.Config)
	}
	if info.Config.Duplicates != DefaultDuplicateWindow {
		t.Errorf("Duplicates = %v", info.Config.Duplicates)
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

func TestJetStreamCollectorReportsLagAndDLQ(t *testing.T) {
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	js, _ := nc.JetStream()
	if err := EnsureStream(js, WebhookStreamSpec(cfg), quietLog()); err != nil {
		t.Fatal(err)
	}
	if _, err := js.AddConsumer("WEBHOOK_EVENTS", &nats.ConsumerConfig{Durable: "backend-meta", FilterSubject: "meta.>", AckPolicy: nats.AckExplicitPolicy}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := js.Publish("meta.message.received", []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := js.Publish("meta.consumer.dlq.message", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(NewJetStreamCollector(js, "WEBHOOK_EVENTS", "backend-meta", "not-created-yet"))
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			key := f.GetName()
			for _, l := range m.GetLabel() {
				if l.GetName() == "subject" || l.GetName() == "consumer" {
					key += "|" + l.GetValue()
				}
			}
			got[key] = m.GetGauge().GetValue()
		}
	}
	// 3 messages + the DLQ message all match the consumer filter meta.>.
	if got["jetstream_consumer_pending_messages|backend-meta"] != 4 {
		t.Fatalf("pending = %v, want 4 (%v)", got["jetstream_consumer_pending_messages|backend-meta"], got)
	}
	if got["jetstream_dlq_messages|meta.consumer.dlq.message"] != 1 {
		t.Fatalf("dlq = %v, want 1 (%v)", got["jetstream_dlq_messages|meta.consumer.dlq.message"], got)
	}
	if got["jetstream_stream_bytes"] <= 0 {
		t.Fatalf("jetstream_stream_bytes = %v, want > 0", got["jetstream_stream_bytes"])
	}
	if got["jetstream_up"] != 1 {
		t.Fatalf("jetstream_up = %v", got["jetstream_up"])
	}
}

// A message published inside a span reaches the meta subscriber handler in the same trace.
func TestTraceContextCrossesNATS(t *testing.T) {
	observability.SetupTracing(context.Background(), "test", nil)
	nc := startJetStream(t)
	cfg := testNATSConfig("WEBHOOK_EVENTS")
	sub, err := NewMetaSubscriber(nc, cfg, quietLog())
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sub.Subscribe(ctx, "trace-test", func(hctx context.Context, _ sharedevents.CanonicalEvent) error {
		got <- trace.SpanContextFromContext(hctx).TraceID().String()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	pctx, span := observability.Tracer().Start(context.Background(), "producer")
	want := span.SpanContext().TraceID().String()
	js, _ := nc.JetStream()
	data, _ := json.Marshal(sharedevents.CanonicalEvent{EventID: "trace-1", Type: "message.received"})
	msg := &nats.Msg{Subject: "meta.message.received", Data: data, Header: nats.Header{}}
	msg.Header.Set("X-Internal-Token", cfg.InternalToken)
	observability.Inject(pctx, msg.Header) // what the webhook-service publisher does
	if _, err := js.PublishMsg(msg); err != nil {
		t.Fatal(err)
	}
	span.End()
	select {
	case id := <-got:
		if id != want {
			t.Fatalf("consumer trace %s, want producer trace %s", id, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("message not consumed")
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

// 8.5: events dropped by retention before the consumer acked them are reported, and the
// gauge is zero for a consumer that is caught up.
func TestJetStreamCollectorReportsExpiredUnackedMessages(t *testing.T) {
	nc := startJetStream(t)
	js, _ := nc.JetStream()
	if _, err := js.AddStream(&nats.StreamConfig{Name: "WEBHOOK_EVENTS", Subjects: WebhookStreamSubjects, MaxMsgs: 2}); err != nil {
		t.Fatal(err)
	}
	for _, durable := range []string{"slow", "caught-up"} {
		if _, err := js.AddConsumer("WEBHOOK_EVENTS", &nats.ConsumerConfig{Durable: durable, FilterSubject: "meta.>", AckPolicy: nats.AckExplicitPolicy}); err != nil {
			t.Fatal(err)
		}
	}
	// 5 messages into a stream that keeps 2: sequences 1-3 are discarded unprocessed.
	for i := 0; i < 5; i++ {
		if _, err := js.Publish("meta.message.received", []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}
	sub, err := js.PullSubscribe("meta.>", "caught-up", nats.BindStream("WEBHOOK_EVENTS"))
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := sub.Fetch(2, nats.MaxWait(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range msgs {
		if err := m.AckSync(); err != nil {
			t.Fatal(err)
		}
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(NewJetStreamCollector(js, "WEBHOOK_EVENTS", "slow", "caught-up"))
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	expired := map[string]float64{}
	for _, f := range families {
		if f.GetName() != "jetstream_consumer_expired_unacked_messages" {
			continue
		}
		for _, m := range f.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "consumer" {
					expired[l.GetValue()] = m.GetGauge().GetValue()
				}
			}
		}
	}
	if expired["slow"] != 3 || expired["caught-up"] != 0 {
		t.Fatalf("expired unacked = %v, want slow=3 caught-up=0", expired)
	}
}

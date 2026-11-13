package nats

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/nats-io/nats.go"
	sharedconfig "github.com/omnichannel/omni-chan-v2/shared/config"
)

// Defaults for the webhook events stream when config leaves a field at zero.
const (
	DefaultStreamMaxAge = 7 * 24 * time.Hour
	// DefaultDuplicateWindow is how long JetStream remembers Nats-Msg-Id values.
	DefaultDuplicateWindow = 2 * time.Minute
)

// WebhookStreamSubjects are all subjects carried by the webhook events stream.
var WebhookStreamSubjects = []string{"events.>", "meta.>"}

// StreamSpec is the desired shape of a stream. Only fields set here are managed.
type StreamSpec struct {
	Name       string
	Subjects   []string
	MaxAge     time.Duration
	MaxBytes   int64
	Duplicates time.Duration
}

// WebhookStreamSpec returns the spec every service uses for the webhook events stream.
// Publisher and subscribers must agree on it so no process clobbers another's config.
func WebhookStreamSpec(cfg sharedconfig.NATS) StreamSpec {
	maxAge := cfg.StreamMaxAge
	if maxAge <= 0 {
		maxAge = DefaultStreamMaxAge
	}
	return StreamSpec{
		Name:       cfg.StreamName,
		Subjects:   WebhookStreamSubjects,
		MaxAge:     maxAge,
		MaxBytes:   cfg.StreamMaxBytes,
		Duplicates: DefaultDuplicateWindow,
	}
}

// streamAPI is the subset of nats.JetStreamContext used here (eases testing).
type streamAPI interface {
	StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
}

// EnsureStream creates the stream if it does not exist, otherwise reconciles it
// with spec without destroying anything an operator configured:
//   - missing subjects are added (never removed);
//   - MaxAge / MaxBytes / Duplicates are filled in only when currently unlimited/unset;
//   - every other field (retention, storage, replicas, ...) is preserved as-is.
//
// Any StreamInfo error other than "not found" is returned instead of attempting a
// create, so a transient outage never triggers a conflicting AddStream.
func EnsureStream(js streamAPI, spec StreamSpec, log *slog.Logger) error {
	if spec.Name == "" {
		return errors.New("ensure stream: empty stream name")
	}
	if log == nil {
		log = slog.Default()
	}
	info, err := js.StreamInfo(spec.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		cfg := newStreamConfig(spec)
		if _, err := js.AddStream(&cfg); err != nil {
			return fmt.Errorf("create stream %s: %w", spec.Name, err)
		}
		log.Info("created jetstream", "stream", spec.Name, "max_age", cfg.MaxAge.String(), "max_bytes", cfg.MaxBytes)
		return nil
	}
	if err != nil {
		return fmt.Errorf("stream info %s: %w", spec.Name, err)
	}

	merged, changed := mergeStreamConfig(info.Config, spec)
	if !changed {
		return nil
	}
	if _, err := js.UpdateStream(&merged); err != nil {
		return fmt.Errorf("update stream %s: %w", spec.Name, err)
	}
	log.Info("reconciled jetstream", "stream", spec.Name,
		"subjects", merged.Subjects, "max_age", merged.MaxAge.String(), "max_bytes", merged.MaxBytes)
	return nil
}

func newStreamConfig(spec StreamSpec) nats.StreamConfig {
	cfg := nats.StreamConfig{
		Name:       spec.Name,
		Subjects:   slices.Clone(spec.Subjects),
		Retention:  nats.LimitsPolicy,
		Storage:    nats.FileStorage,
		MaxAge:     spec.MaxAge,
		MaxBytes:   -1,
		Duplicates: spec.Duplicates,
	}
	if spec.MaxBytes > 0 {
		cfg.MaxBytes = spec.MaxBytes
	}
	return cfg
}

// mergeStreamConfig returns cur with spec's managed fields filled in where unset.
func mergeStreamConfig(cur nats.StreamConfig, spec StreamSpec) (nats.StreamConfig, bool) {
	out := cur
	out.Subjects = slices.Clone(cur.Subjects)
	changed := false

	for _, subj := range spec.Subjects {
		if !slices.Contains(out.Subjects, subj) {
			out.Subjects = append(out.Subjects, subj)
			changed = true
		}
	}
	if out.MaxAge == 0 && spec.MaxAge > 0 {
		out.MaxAge = spec.MaxAge
		changed = true
	}
	if out.MaxBytes <= 0 && spec.MaxBytes > 0 {
		out.MaxBytes = spec.MaxBytes
		changed = true
	}
	if out.Duplicates == 0 && spec.Duplicates > 0 {
		out.Duplicates = spec.Duplicates
		changed = true
	}
	return out, changed
}

# Phase 12.4 — Production Hardening & Broadcast Reliability Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Media handle resolution, transactional send finalization, distributed rate limiting, worker health, stuck-recipient recovery — no orchestration, webhook, or frontend changes

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Phase 12.2 / 12.3 Finding Validation

| Prior Finding | Status |
|---------------|--------|
| 12.3 #1 — Media headers require public URL only | **Resolved** — `media_id` support added |
| 12.2 #2 — Counter update non-atomic with recipient row | **Resolved** — `FinalizeRecipientSent` single transaction |
| 12.2 #3 — Per-process rate limiter multiplies throughput | **Resolved** — Redis token bucket `broadcast:rate:{channel_id}` + fallback |
| 12.3 #2 — `sample_values` fallback for custom fields | Unchanged (documented) |
| 12.3 #6 — In-memory metrics only | Unchanged — counters via `Metrics.Snapshot()` |

---

## Reliability Audit

### Media handling

- `meta.MediaReference` enforces exactly one of `link` or `media_id`
- Validation rejects missing reference, dual reference, and invalid URLs before Meta API
- Builder emits `{"image":{"id":"..."}}` or `{"image":{"link":"..."}}` as appropriate
- Metrics: `media_id_sends_total`, `media_link_sends_total`

### Counter consistency

`campaign.FinalizeRecipientSent` runs in one PostgreSQL transaction:

1. `campaign_recipients`: `sending → sent` + WAMID (conditional on `status = sending`)
2. `campaigns.sent_count += 1`
3. `campaign_daily_stats.messages_sent += 1`

Idempotent retry: if recipient already `sent`/`delivered`/`read` with WAMID, returns `applied=false` without counter increment.

### Distributed throttling

- Primary: `RedisTokenBucketLimiter` key `broadcast:rate:{channel_id}` (80/sec default)
- Fallback: `FallbackRateLimiter` → in-process limiter on Redis errors
- Metric: `rate_limit_fallback_total`

### Crash recovery

| Layer | Mechanism |
|-------|-----------|
| Batch | Phase 12.1 stale `processing` lease recovery (10m) |
| Recipient | Phase 12.4 `RecipientRecovery`: `sending` + `wamid IS NULL` + `updated_at < now()-15m` → `queued` |
| Send retry | Worker idempotency: skip `sent`/`delivered`/`read`/existing WAMID |

### Operational visibility

- `GET /health` on `BROADCAST_WORKER_HTTP_ADDR` (default `:8091`)
- Returns `{healthy, nats, postgres, redis}` — `healthy` requires Postgres + NATS
- Metrics: `worker_batches_processing`, `worker_batches_completed`, `worker_restarts_total`, `meta_send_duration_ms`, `recipient_recovery_total`

---

## Replay Safety Audit

| Entity | Idempotency key / guard | Location |
|--------|-------------------------|----------|
| Recipient send | `status` transition + `wamid IS NULL` skip; `FinalizeRecipientSent` conditional `WHERE status = sending` | `batch_processor`, `send_finalize.go` |
| Recipient WAMID | Unique partial index `idx_campaign_recipients_wamid`; webhook lookup by WAMID | migration `000012`, `FindRecipientByWAMID` |
| Batch | `status = completed` → ignore event; lease claim prevents duplicate workers | `ClaimBatchForProcessing`, `batch_processor` |
| Job | Orchestrator idempotent on `CountRecipientsByJob`; unique active job index | Phase 12.1 |
| Webhook status | `provider:status:{wamid}:{status}` via `idempotency.StatusKey` | `eventconsumer/idempotency` |
| Webhook message | `provider:message:{wamid}` via `idempotency.MessageKey` | `eventconsumer/idempotency` |
| Counter on send | Transactional — no increment unless recipient row transitions | `FinalizeRecipientSent` |
| Counter on delivery | `CanTransition` + monotonic DB trigger | Phase 11.6 `delivery` |

---

## Deliverables Checklist

| Component | Status |
|-----------|--------|
| `pkg/meta/media_reference.go` | Done |
| Media ID in template builder + validation | Done |
| `campaign/send_finalize.go` transactional finalization | Done |
| `RedisTokenBucketLimiter` + `FallbackRateLimiter` | Done |
| `worker/health.go` GET `/health` | Done |
| `worker/recovery.go` stuck recipient recovery | Done |
| Extended metrics | Done |
| Tests (media, rate limit, recovery, health, validation) | Done |

---

## Tests

| Test | Coverage |
|------|----------|
| `TestMediaReferenceValidateDualReference` | Rejects link + media_id |
| `TestBuildImageHeaderMediaID` | Meta `image.id` payload |
| `TestBuildImageHeaderDualReferenceFails` | Builder rejects dual reference |
| `TestValidateDualMediaReference` | Pre-send validation |
| `TestValidateMediaIDAccepted` | Media ID passes validation |
| `TestFallbackRateLimiterOnRedisError` | Runtime Redis fallback |
| `TestRedisTokenBucketSharedKey` | Key `broadcast:rate:{channel_id}` |
| `TestRecoverStaleSendingRecipientRequeued` | Stale `sending` → `queued` |
| `TestRecoverSkipsSentRecipient` | Sent/WAMID untouched |
| `TestHealthEndpointPostgresAndNATSDown` | Unhealthy when deps down |
| Worker integration tests | Still pass with `mockFinalizer` |

---

## Findings

### MEDIUM

1. **Failed-recipient counter not transactional** — `markFailed` updates recipient status without atomic `failed_count` / daily stat increment. Send-success path is fixed; failure counters may drift until Phase 11.6 reconciliation (same pre-existing gap for failures).

2. **No integration test for `FinalizeRecipientSent` rollback** — Transaction behavior validated by SQL design and idempotent guards; full PG rollback test requires testcontainers (not added).

3. **Redis optional for `healthy`** — Worker reports `healthy: true` when Postgres + NATS are up even if Redis is down. Rate limiting falls back to per-process limiting in that case.

### LOW

4. **`worker_batches_processing` is a counter not a gauge** — Increments per batch claim; does not decrement on completion. Useful for throughput; not a live in-flight count.

5. **Health endpoint unauthenticated** — Standard for internal worker probes; should be network-restricted in deployment.

6. **Recipient recovery 15m vs batch lease 10m** — Intentional stagger: batch recovery runs first; recipient recovery catches edge cases where batch remains `processing` with stuck `sending` rows.

---

## Recommendation

**APPROVE WITH FINDINGS**

Phase 12.4 closes the production gaps needed for large-scale broadcast rollout: Meta media ID support, atomic send counter finalization, shared Redis rate limiting with fallback, worker health probes, and stuck-recipient recovery. The broadcast path is crash-safe and operationally observable. Address failed-counter transactional finalization (MEDIUM #1) in a follow-up if failure-rate reporting must be strictly real-time.

---

## Out of Scope (confirmed unchanged)

Frontend, webhooks, delivery tracking redesign, orchestration, RBAC, analytics redesign, campaign scheduling

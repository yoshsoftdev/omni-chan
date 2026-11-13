# Phase 12.2 — Broadcast Send Worker Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Send worker only — consumes `broadcast.batch.send`, calls Meta Cloud API, writes WAMID to `campaign_recipients`

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Architecture Review

### End-to-end flow

```text
broadcast.batch.send (NATS)
        ↓
cmd/broadcast-worker
        ↓
internal/broadcast/worker (consumer → batch_processor → sender)
        ↓
pkg/meta.SendTemplateMessage
        ↓
campaign_recipients.wamid + status=sent
        ↓
Phase 11.6 delivery tracking (webhook correlation on WAMID)
```

### Repository usage

- New `campaign.SendRepository` (`send.go`, `send_repository.go`) for batch claim, lease refresh, recipient listing, batch completion
- Recipient updates reuse `campaign.Repository.UpdateRecipientDelivery` (extended with `provider_response`, `IncrementAttempt`)
- Campaign/template/channel/contact reads use narrow worker interfaces (`deps.go`) — no handler leakage
- `delivery.ReportAggregator` increments `sent_count` on `sending → sent` transition

### Meta client usage

- All sends go through `pkg/meta.Client.SendTemplateMessage`
- HTTP 4xx/5xx now return `meta.APIError` for consistent classification
- Channel tokens decrypted via `cryptoutil.TokenCipher` + `channel.FindCredentials`

### Transaction boundaries

| Operation | Transaction |
|-----------|-------------|
| Batch claim (`queued` → `processing`) | Single conditional `UPDATE … RETURNING` |
| Recipient `queued → sending → sent` | Per-recipient updates (matches async worker model) |
| Batch complete + `batches_completed` increment | Single PG tx (`CompleteBatch`) |
| Counter + daily stat on send | `ReportAggregator` (two writes, non-atomic with recipient row) |

### Idempotency

| Guard | Behavior |
|-------|----------|
| Batch `completed` | Event ignored |
| Batch held by another worker (valid lease) | Event acked, no work |
| Recipient `sent` / `delivered` / `read` | Skipped |
| Recipient with `wamid` | Skipped |
| Stale `processing` batch | Phase 12.1 recovery resets `sending → queued` and republishes |

### Delivery tracking compatibility

- Real WAMID stored in `campaign_recipients.wamid` (no mock IDs)
- `provider_response` JSONB added (`000021_broadcast_send_support`)
- Phase 11.6 `FindRecipientByWAMID` / webhook path unchanged
- Worker does **not** mark campaigns completed (owned by `CompletionMonitor`)

### Rate limiting

- Per-channel limiter interface (`RateLimiter`)
- Default 80 msg/sec (`BROADCAST_SEND_RATE_LIMIT` env)
- In-process limiter per channel when Redis unavailable; `RedisRateLimiter` when Redis connects
- Meta HTTP 429: delays batch processing, does not permanently fail recipient

### Failure classification

| Class | Recipient outcome | Retry |
|-------|-------------------|-------|
| Permanent (invalid phone, template error, blocked) | `failed` | None |
| Transient (5xx, timeout) | `queued` with `attempt_count++` | 30s / 2m / 10m, max 3 |
| Rate limited (429) | Same attempt retried after backoff | N/A |

---

## Deliverables Checklist

| Component | Status |
|-----------|--------|
| `cmd/broadcast-worker` | Done |
| `internal/broadcast/worker/consumer.go` | Done |
| `internal/broadcast/worker/sender.go` | Done |
| `internal/broadcast/worker/rate_limiter.go` | Done |
| `internal/broadcast/worker/batch_processor.go` | Done |
| `internal/broadcast/worker/failure_classifier.go` | Done |
| `internal/broadcast/worker/metrics.go` | Done |
| Migration `000021_broadcast_send_support` | Done |
| `campaign/template_send.go` (variable rendering) | Done |
| Tests (success, retry, permanent, idempotency, 429, recovery) | Done |

---

## Tests

| Test | Coverage |
|------|----------|
| `TestSendOneRecipientStoresWAMID` | Send, WAMID persist, `sent` status, batch complete |
| `TestRetry500SucceedsSecondAttempt` | Transient 500, succeeds on retry |
| `TestPermanentFailureInvalidPhone` | 400 invalid phone → `failed` |
| `TestIdempotencyDuplicateBatchEvent` | Completed batch → no Meta calls |
| `TestRateLimit429DelaysAndRetries` | 429 backoff, eventual success |
| `TestWorkerRestartBatchContinues` | Skips sent recipient, sends remainder |
| `TestClassifyPermanentTemplateError` | Template errors permanent |
| `TestClassifyTransientTimeout` | Network timeout transient |

---

## Findings

### MEDIUM

1. **Template component rendering is body-only** — `BuildTemplateComponents` maps `variable_mapping.body` only. HEADER/FOOTER/QUICK_REPLY/URL/COPY_CODE/media headers from `template.components` are not yet expanded. Matches legacy worker scope; campaigns with rich templates need a follow-up mapper.

2. **Send counter non-atomic with recipient row** — `ReportAggregator.ApplyTransition` runs after recipient update. A crash between the two can drift counters until Phase 11.6 reconciliation runs.

3. **In-process rate limiter is per worker process** — With multiple `broadcast-worker` replicas and no Redis, effective throughput = `80 × replicas` per channel. Redis limiter is wired when Redis is available.

### LOW

4. **`lease_expires_at` column not added** — Lease semantics reuse Phase 12.1 `locked_at` + 10m stale recovery. Functionally equivalent.

5. **Metrics are in-memory counters** — `Metrics.Snapshot()` exposes required names; Prometheus HTTP endpoint not added in this phase.

6. **Campaign `queued → sending` is defensive** — Orchestrator already transitions during batch planning; worker repeats the guard for safety.

---

## Recommendation

**APPROVE WITH FINDINGS**

Phase 12.2 completes the broadcast send path: orchestrator publishes batch work, the send worker claims batches, sends via Meta Cloud API, persists real WAMIDs, and hands off delivery lifecycle to Phase 11.6. Address template component expansion (MEDIUM #1) before marketing templates with media/buttons go to production.

---

## Out of Scope (confirmed unchanged)

- Frontend, websocket, inbox, webhooks, RBAC, readiness, orchestration redesign, campaign completion logic

# Retry Strategy Review — Phase 12.0

**Status:** Design review  
**Proposed recipient backoff:** 30s → 2m → 10m

## Failure classes

| Class | Signals | Retry? | Handler |
|-------|---------|--------|---------|
| **Transient** | TCP timeout, 5xx, connection reset | Yes | Recipient + batch backoff |
| **Rate limit** | HTTP 429, Meta 130429 | Yes (deferred) | Token bucket + batch retry; **does not consume attempt** |
| **Network** | DNS, TLS, dial timeout | Yes | Same as transient |
| **Meta outage** | Sustained 5xx across channel | Pause job | Circuit breaker; resume when health OK |
| **Permanent** | 131026 invalid number, template paused, 400 policy | No | `failed` immediately |
| **Auth** | 190 token expired | Once | Refresh credential, single retry |
| **Worker crash** | Batch stuck `processing` | Yes | Stale lock recovery → requeue batch |

## Recipient-level retry (transient)

| Attempt | Delay | Cumulative |
|---------|-------|------------|
| 1 | 30s | 30s |
| 2 | 2m | 2m 30s |
| 3 | 10m | 12m 30s |
| 4+ | — | → `failed` |

**Verdict:** **Approve** 30s / 2m / 10m for recipient transient errors.

Rules:

- Increment `campaign_recipients.attempt_count`
- Transition `sending → queued` (not `pending`) for retry
- Skip retry if campaign cancel flag set
- Permanent errors: no retry, set `failure_code`

## Batch-level retry (worker / NATS)

| Attempt | Delay | Trigger |
|---------|-------|---------|
| 1 | 1m | NATS NAK, worker crash |
| 2 | 5m | Stale lock recovery |
| 3 | 15m | Repeated failure |
| 4 | 30m | |
| 5 | 60m | → batch `failed` |

Stored in `broadcast_batches.attempt_count` + `next_retry_at`.

Publish to `broadcast.batch.retry` with JetStream delayed delivery or `NakWithDelay`.

## Rate limit retry (special case)

| Behavior | Detail |
|----------|--------|
| Does not increment recipient `attempt_count` | Rate limit is infrastructure, not recipient fault |
| Batch deferred | `next_retry_at = now + delay` |
| Delay source | `Retry-After` header, else 60s, else token bucket refill estimate |
| Max rate-limit deferrals | 10 per batch before batch `failed` with `rate_limited` summary |

## Meta outage

| Signal | Action |
|--------|--------|
| >50% sends return 5xx for 5 min | Set `broadcast_jobs.status = paused` (new sub-state or `sending` + flag) |
| Channel health webhook | Resume when `channel.updated` shows healthy |
| Manual | Ops API to resume job |

Workers check `pause:channel:{id}` Redis flag before send.

## Worker crash

| Step | Action |
|------|--------|
| 1 | Batch has `locked_by`, `locked_at` |
| 2 | Lease TTL: 10 minutes |
| 3 | Stale recovery cron resets `processing → queued` if `locked_at` expired |
| 4 | Recipients in `sending` without `wamid` → `queued` |
| 5 | Recipients with `wamid` → remain `sent` (idempotent) |

## Idempotency (duplicate delivery prevention)

| Check | When |
|-------|------|
| `recipient.wamid IS NOT NULL` | Skip Meta call |
| `recipient.status IN (sent, delivered, read)` | Skip |
| Unique index on `wamid` | DB enforces |

## Decision matrix

| Scenario | Max attempts | Backoff |
|----------|--------------|---------|
| Recipient transient | 3 | 30s, 2m, 10m |
| Batch worker failure | 5 | 1m, 5m, 15m, 30m, 60m |
| Rate limit (429) | 10 deferrals | 60s+ adaptive |
| Job audience resolve | 2 | 5m, 30m |
| Meta auth (190) | 1 | Immediate after token refresh |

## Decision

**Approve** proposed recipient backoff (30s / 2m / 10m) with additions:

- Separate rate-limit deferral from attempt count
- Batch-level NATS retry schedule for crash recovery
- Stale lock recovery required before Phase 12.1 ships to production

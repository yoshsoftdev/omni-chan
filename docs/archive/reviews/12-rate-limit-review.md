# Rate Limiter Review — Phase 12.0

**Status:** Design review  
**Proposed mechanism:** Redis token bucket

## Requirement summary

| Requirement | Must satisfy |
|-------------|--------------|
| Avoid Meta 429 | Proactive throttle before API call |
| Multiple send workers | Shared limit state across pods |
| Fairness | No single workspace starves others |
| Recover from 429 | Reactive backoff + token refill |

## Scope evaluation

| Scope | Key | Verdict |
|-------|-----|---------|
| **Per channel** (`channel_id` / `phone_number_id`) | `rate:channel:{channel_id}` | **Primary — required** |
| **Per WABA** | `rate:waba:{waba_id}` | **Optional** — only if one WABA maps to multiple channels |
| **Per workspace** (`team_id`) | `rate:team:{team_id}` | **Secondary fairness cap** |

### Recommendation

```
effective_tokens = min(channel_bucket, team_bucket)
```

| Bucket | Purpose | Default refill (configurable) |
|--------|---------|-------------------------------|
| Channel | Enforce Meta WABA tier | 80 tokens/sec (starter), up to tier limit |
| Team | Prevent one customer consuming platform capacity | 200 tokens/sec platform-wide default |

**WABA-level bucket:** Denormalize `waba_id` on `broadcast_batches` or resolve from `channels` at worker start. Use only when multiple `phone_number_id` share one WABA and share tier.

## Implementation: Redis token bucket

| Property | Value |
|----------|-------|
| Algorithm | Token bucket with Lua `EVAL` (atomic) |
| Keys | `rate:channel:{id}`, `rate:team:{id}` |
| TTL | Key expires after 24h idle (refill on access) |
| Acquire | `ACQUIRE 1` before each Meta send; block or defer batch on empty |

### Pseudocode

```
tokens, last = GET bucket
elapsed = now - last
tokens = min(capacity, tokens + elapsed * rate)
if tokens >= 1:
  tokens -= 1
  SET bucket
  return OK
else:
  return DEFER(delay_ms)
```

### On HTTP 429 from Meta

1. Do **not** consume token (or refund)
2. Set batch `next_retry_at`
3. Publish `broadcast.batch.retry` with `delay_ms` from `Retry-After` or 60s default
4. Increment `rate_limit_hits_total{channel_id}`

## Multi-worker safety

| Concern | Mitigation |
|---------|------------|
| Race on token acquire | Lua script — single atomic operation |
| Worker bypass | Send worker **must** acquire token; no API call without token |
| Redis down | **Fail closed** for send (defer batch); do not blast Meta |
| Hot key | One key per channel is acceptable; Meta limit is per WABA anyway |

## Existing code reuse

| Package | Use |
|---------|-----|
| `pkg/meta/retry.go` | `IsRateLimited`, retry wrapper for 429 |
| `pkg/meta/errors.go` | Classify 429 vs permanent errors |

Rate limiter is **proactive**; Meta client retry is **reactive** — both layers required.

## Configuration

| Env | Default | Description |
|-----|---------|-------------|
| `BROADCAST_RATE_CHANNEL_MPS` | 80 | Tokens/sec per channel |
| `BROADCAST_RATE_TEAM_MPS` | 200 | Fairness cap per team |
| `BROADCAST_RATE_BURST` | 20 | Bucket capacity (burst allowance) |

Tier overrides: store `messaging_limit_tier` on `channels.health_metadata` (from Meta); map tier → mps at runtime.

## Alternatives considered

| Alternative | Verdict |
|-------------|---------|
| In-process limiter only | Rejected — breaks with N workers |
| NATS consumer max ack pending | Complements but does not replace WABA limit |
| Postgres semaphore | Too slow for per-message gate |
| Fixed sleep between sends | Rejected — cannot adapt to tier changes |

## Decision

| Item | Selection |
|------|-----------|
| **Mechanism** | Redis token bucket (Lua) |
| **Primary scope** | Per `channel_id` |
| **Secondary scope** | Per `team_id` fairness |
| **WABA scope** | Phase 12.2 if multi-number WABA sharing observed |
| **429 behavior** | Defer batch + `broadcast.batch.retry` |

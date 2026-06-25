# Failure Scenarios — Phase 12.0

**Status:** Design review

## Scenario matrix

| Scenario | Impact | Detection | Response | Recovery |
|----------|--------|-----------|----------|----------|
| **Redis down** | Rate limit + cancel flags unavailable | Health check / Redis ping | Send workers **stop acquiring sends**; defer batches | Restore Redis; batches retry via `next_retry_at` |
| **NATS down** | No new work distribution | Consumer disconnect alert | Orchestrator cannot publish; API returns 503 on send if publish fails | Reconnect; JetStream redelivers unacked |
| **Meta down** | 5xx on all sends | Error rate > threshold | Pause channel/job; `broadcast.batch.retry` with long delay | Resume on health / manual |
| **Worker restart** | Mid-batch processing interrupted | `locked_at` ages | Stale recovery requeues batch | Idempotent resend skips `wamid` |
| **Duplicate delivery** | Same batch message twice | Metrics: duplicate batch claims | DB lease + `wamid` check | No duplicate Meta calls |
| **Stuck batch** | `processing` forever | `locked_at` > lease | Stale recovery cron | Requeue or fail after max attempts |

---

## Redis down

| Component | Behavior |
|-----------|----------|
| Rate limiter | Fail closed — do not send |
| Cancel flag | Fall back to Postgres `campaigns.status = cancelled` |
| Idempotency (webhook) | Consumer may use memory store fallback (existing pattern) |

**Risk:** Thundering herd when Redis returns — use jittered batch retry.

---

## NATS down

| Phase | Behavior |
|-------|----------|
| Job created in PG, publish fails | Transaction rollback OR outbox pattern — **recommend transactional outbox in Phase 12.1** |
| Send worker disconnected | Batches stay `queued`; no sends |
| Partial ack | JetStream redelivers; DB lease prevents double processing |

**Mitigation:** Postgres outbox table `broadcast_outbox` polled by orchestrator — document in Phase 12.1.

---

## Meta down

| Duration | Action |
|----------|--------|
| Single 5xx | Recipient retry (30s / 2m / 10m) |
| Sustained outage | Set `pause:channel:{id}`; stop consuming new batch messages for channel |
| 429 storm | Rate limiter defer; reduce worker count |

---

## Worker restart

| State at crash | Recovery |
|----------------|----------|
| Batch `processing`, recipients `queued` | Reprocess batch |
| Recipients `sending`, no wamid | → `queued`, retry |
| Recipients `sent` | Unchanged |
| NATS message not acked | Redelivered to another worker |

---

## Duplicate delivery

| Vector | Guard |
|--------|-------|
| NATS redelivery | DB batch lease |
| Two workers same batch | `UPDATE ... WHERE status = queued` row count |
| Same recipient twice | `wamid IS NOT NULL` skip |
| Webhook duplicate | Phase 11.6 idempotency |

---

## Stuck batch

| Symptom | Cause | Fix |
|---------|-------|-----|
| `processing` > 10 min | Worker OOM, deadlock | Stale recovery |
| `queued` + old `next_retry_at` | Retry scheduler stuck | Orchestrator republish |
| Job `sending`, batches all `completed`, recipients pending | Planner bug | Manual job fail + reconcile |

**Alert:** `broadcast_batch_stuck_total` when `processing` and `locked_at < now() - 15m`.

---

## Cascading failures

```
Redis down → sends pause → queue depth grows → NATS disk fills
```

| Safeguard |
|-----------|
| Max stream bytes on `BROADCAST` stream |
| Consumer max ack pending limit |
| API backpressure: reject new `POST /send` when queue depth > threshold |

## Decision

All scenarios have documented responses. **Phase 12.1 must implement:**

1. Stale batch recovery
2. Redis fail-closed in send worker
3. Transactional outbox or publish-after-commit for `broadcast.job.created`

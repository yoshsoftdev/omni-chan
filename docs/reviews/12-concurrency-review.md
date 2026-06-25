# Concurrency Model Review — Phase 12.0

**Status:** Design review

## Models evaluated

| Model | Description | Verdict |
|-------|-------------|---------|
| **1 worker** | Single send process | Reject — cannot meet 300k SLA |
| **N workers** | Horizontally scaled send pool | **Approve** |
| **Channel affinity** | Worker sticky to `channel_id` | Optional Phase 12.2 |
| **Workspace affinity** | Worker sticky to `team_id` | Reject — poor utilization |

## Recommended model: N workers + batch lease

```
                    ┌──────────────┐
  broadcast.batch.send ──►│ NATS queue   │
                    │ group        │
                    └──────┬───────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
    Worker A           Worker B           Worker C
         │                 │                 │
         └────────► Redis rate limit ◄──────┘
                           │
                           ▼
                      Meta API
```

### Can two workers send the same batch?

**Must not happen.** Enforcement layers:

| Layer | Mechanism |
|-------|-----------|
| **1. NATS JetStream** | Queue group — one consumer acks each message |
| **2. DB batch lease** | `UPDATE broadcast_batches SET status=processing, locked_by=$worker, locked_at=now() WHERE id=$1 AND status IN (queued) AND (locked_at IS NULL OR locked_at < expiry)` |
| **3. Idempotent sends** | Recipients with `wamid` skipped |

NATS alone is insufficient if message is redelivered after ack failure — DB lease is required.

### Lock ownership

| Field | Purpose |
|-------|---------|
| `locked_by` | Worker instance ID (hostname + uuid) |
| `locked_at` | Lease start |
| Lease TTL | 10 minutes |
| Recovery | Stale recovery cron clears expired locks |

### Recipient-level concurrency

Within a batch, **single worker processes recipients sequentially** (per batch message). Parallelism comes from **multiple batches**, not parallel sends inside one batch.

Rationale:

- Simpler rate limit accounting
- Matches token bucket per message
- Avoids goroutine storm on 250-recipient batches

## Channel affinity (optional)

| Pros | Cons |
|------|------|
| Local token bucket cache | Uneven load if one channel dominates |
| Predictable ordering per channel | Complex routing |

**v1:** No affinity — shared Redis limiter.  
**v2:** Subject `broadcast.batch.send.{channel_id}` with per-channel consumer max ack pending.

## Workspace affinity

Rejected — campaigns from large workspace would saturate dedicated workers; channel limiter already provides fairness via team cap.

## Orchestrator concurrency

| Component | Concurrency |
|-----------|-------------|
| Scheduler | Single leader (Redis lock `orchestrator:leader`) |
| Audience resolver | One job at a time per `job_id` (idempotent) |
| Batch planner | Same process as resolver |

Multiple orchestrator pods: only leader runs scheduler tick; all can consume `broadcast.job.created` with work-queue semantics.

## Postgres SKIP LOCKED fallback

If NATS unavailable, workers may claim batches:

```sql
SELECT id FROM broadcast_batches
WHERE status = 'queued' AND next_retry_at <= now()
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 1
```

Use as **degraded mode** only — not primary path.

## Decision

| Item | Selection |
|------|-----------|
| Send concurrency | **N workers** (HPA on consumer lag) |
| Batch exclusivity | NATS queue group + DB lease |
| In-batch parallelism | **Sequential** (v1) |
| Channel affinity | Defer to Phase 12.2 |
| Orchestrator | Leader-elected scheduler + horizontal job consumers |

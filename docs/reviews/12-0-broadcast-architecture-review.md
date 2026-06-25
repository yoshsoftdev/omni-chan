# Phase 12.0 — Broadcast Worker Architecture Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Design review only — no implementation

---

## Executive summary

The Phase 8 broadcast execution architecture remains valid. Phase 11.4 and 11.6 already deliver **status consumption, report aggregation, completion detection, and reconciliation**. Phase 12 should focus on the **send path** (orchestrator + send workers) without duplicating delivery infrastructure.

---

## Decisions approved

| Area | Decision |
|------|----------|
| **Worker topology** | `broadcast-orchestrator` (scheduler + audience + planner + stale recovery) + `broadcast-send-worker` pool |
| **NATS subjects** | `broadcast.job.created`, `broadcast.audience.resolved`, `broadcast.batch.send`, `broadcast.batch.retry`, `broadcast.job.completed` (optional) |
| **Batch size** | **100** default (configurable 50–250) |
| **Rate limiter** | Redis token bucket — primary per `channel_id`, fairness per `team_id` |
| **Concurrency** | N send workers; batch lease in Postgres; sequential sends per batch |
| **Recipient retry** | 30s → 2m → 10m (3 attempts) |
| **Batch retry** | 1m → 5m → 15m → 30m → 60m (5 attempts) |
| **Cancellation** | `pending` / `queued` / `sending` (no wamid) → `cancelled`; `sent`+ irreversible |
| **WAMID** | Atomic write on send; existing webhook correlation approved |

---

## Findings (non-blocking)

| # | Finding | Action |
|---|---------|--------|
| 1 | **Status/report workers already exist** in `backend-service` | Do not build parallel NATS consumers for delivery |
| 2 | **Proposed NATS subjects** (`batch.ready`, `batch.processing`, etc.) are redundant | Use minimal subject set per [`12-nats-broadcast-review.md`](12-nats-broadcast-review.md) |
| 3 | **Send worker stub** (`worker_compat.go`) | Phase 12.1 first implementation target |
| 4 | **Transactional outbox** not yet designed | Required for reliable `job.created` publish — add in Phase 12.1 |
| 5 | **Composite index** `(team_id, wamid)` | Defer until load test; monitor lookup latency |
| 6 | **Job pause state** for Meta outage | Use Redis `pause:channel:{id}` in v1; formal job sub-state optional |

---

## Known risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Meta tier limits campaign duration | High | Surface tier in UI; rate limiter fail-closed |
| Redis outage stops all sends | Medium | Fail closed; Postgres cancel fallback |
| NATS publish after DB commit failure | Medium | Transactional outbox (Phase 12.1) |
| Partial cancel (some sent) | Low | Document expected behavior |
| Audience resolve timeout at 1M | Medium | Streaming resolver; 5k chunks |

---

## Review artifacts

| Task | Document |
|------|----------|
| Worker inventory | [`12-worker-topology-review.md`](12-worker-topology-review.md) |
| NATS subjects | [`12-nats-broadcast-review.md`](12-nats-broadcast-review.md) |
| Batch size | [`12-batch-size-review.md`](12-batch-size-review.md) |
| Rate limiter | [`12-rate-limit-review.md`](12-rate-limit-review.md) |
| Retry strategy | [`12-retry-strategy-review.md`](12-retry-strategy-review.md) |
| Cancellation | [`12-cancellation-review.md`](12-cancellation-review.md) |
| WAMID correlation | [`12-wamid-review.md`](12-wamid-review.md) |
| Concurrency | [`12-concurrency-review.md`](12-concurrency-review.md) |
| Failure scenarios | [`12-failure-scenarios.md`](12-failure-scenarios.md) |
| Scaling | [`12-scaling-review.md`](12-scaling-review.md) |

---

## Success criteria checklist

| Criterion | Status |
|-----------|--------|
| Architecture approved | ✓ |
| Known risks documented | ✓ |
| Batch size selected (100) | ✓ |
| Concurrency model selected (N workers + batch lease) | ✓ |
| Rate limiter selected (Redis token bucket) | ✓ |
| Worker topology finalized | ✓ |

---

## Recommendation

**Proceed to Phase 12.1 — Scheduler + Audience Resolver** with:

1. `cmd/broadcast-orchestrator` binary
2. NATS `BROADCAST` stream with approved subjects
3. Transactional outbox for job publish
4. Reuse Phase 11.6 `delivery` package for post-send status — do not add `broadcast.recipient.status_changed` in v1

---

## Next phase

**Phase 12.1 — Scheduler + Audience Resolver**

- Implement scheduler poll + `broadcast.job.created`
- Audience resolver with snapshot + recipient bulk insert
- No send worker yet (Phase 12.2)

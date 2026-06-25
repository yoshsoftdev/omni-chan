# Phase 12.1 — Broadcast Orchestrator Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Orchestration only — no Meta send, no delivery tracking changes

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Architecture Review

### Repository usage

- Extended `campaign.PostgresRepository` with `OrchestrationRepository` methods in `orchestration_repo.go`
- Audience resolution uses narrow interfaces (`ContactLookup`, `SegmentMembers`, `TagMembers`) — no handler leakage
- Bulk recipient insert reuses existing `BulkCreateRecipients` (`pgx.Batch`)
- Batch planning runs in a single transaction (`PlanBatchesForJob`)

### Service boundaries

| Layer | Responsibility |
|-------|----------------|
| `cmd/broadcast-orchestrator` | Process bootstrap, NATS, loops |
| `internal/broadcast/orchestrator` | Scheduler, audience, planner, recovery, pipeline |
| `internal/campaign` | Persistence + orchestration SQL |
| `shared/events` + `shared/nats` | Subject contracts + JetStream publisher |

API `campaign.Service` unchanged. `POST /send` remains unavailable (Phase 12.2+).

### Transaction boundaries

| Operation | Transaction |
|-----------|-------------|
| Plan batches + assign recipients | Single PG tx |
| Audience snapshot + bulk recipients | Separate writes; idempotent on `CountRecipientsByJob` |
| Outbox insert | Per message; relay marks published |
| Scheduler queue transition | Conditional update (not full tx with outbox) — mitigated by outbox idempotency at consumer |

### NATS integration

- Stream: `BROADCAST` (`broadcast.>`)
- Subjects: `broadcast.job.created`, `broadcast.batch.send` (via outbox)
- Publisher: `shared/nats/broadcast_publisher.go`
- Consumer: durable `broadcast-orchestrator` on `broadcast.job.created`

---

## Deliverables Checklist

| Component | Status |
|-----------|--------|
| Scheduler (`scheduled` → `queued`, job outbox) | Done |
| Audience resolver (snapshot, dedupe, bulk insert) | Done |
| Batch planner (100 default, `pending` → `queued`) | Done |
| Stale recovery (10m lease, `processing` → `queued`) | Done |
| Metrics | Done |
| Migration `000020_broadcast_outbox` | Done |
| Tests | Done |

---

## Tests

| Test | Coverage |
|------|----------|
| `TestSchedulerPicksOnce` | Single dispatch, duplicate tick prevented |
| `TestAudienceResolverDeduplicates` | Snapshot + one recipient from duplicate sources |
| `TestBatchPlannerChunking` | 100 → 1 batch, 250 → 3 batches |
| `TestLeaseRecovery` | Expired requeued, fresh untouched |
| `TestOutboxJobCreatedPayload` | Outbox payload shape |

---

## Findings

### MEDIUM

| # | Finding | Recommendation |
|---|---------|----------------|
| 1 | `POST /campaigns/:id/send` not wired to orchestrator | Enable in Phase 12.2 when send worker exists |
| 2 | Campaigns created with `SendLater` stay `draft` until explicitly set `scheduled` | API should set `status=scheduled` on schedule — follow-up |
| 3 | Create flow inserts placeholder snapshot at campaign create | Orchestrator creates execution snapshot per job; placeholder remains for draft |

### LOW

| # | Finding | Recommendation |
|---|---------|----------------|
| 4 | Redis scheduler leader lock deferred | Postgres sufficient for v1; add if duplicate tick load observed |
| 5 | `broadcast.batch.send` consumed by send worker not yet implemented | Phase 12.2 |

### CRITICAL / HIGH

None.

---

## State Machine Validation

| Entity | Implemented transitions |
|--------|-------------------------|
| Campaign | `scheduled` → `queued` → `sending` |
| Job | `pending` → `resolving_audience` → `planning_batches` → `sending` |
| Recipient | `pending` → `queued` (+ `skipped` at resolve) |
| Batch | `pending` → `queued`; recovery `processing` → `queued` |

Send-path transitions (`sending` → `sent`, etc.) deferred to Phase 12.2. Delivery webhooks handled by Phase 11.6 `delivery.Tracker`.

---

## Observability

Metrics exposed via `orchestrator.Metrics.Snapshot()`:

- `scheduler_campaigns_total`
- `audience_snapshots_total`
- `recipients_created_total`
- `batches_created_total`
- `batch_recovery_total`

Structured logs include `campaign_id`, `job_id`, `batch_id`, `team_id`, `channel_id`. No phone numbers or tokens logged.

---

## Recommendation

**Proceed to Phase 12.2 — Send Worker.**

Orchestration path is ready:

```
Campaign → Job → Audience Snapshot → Recipients → Batches → NATS (broadcast.batch.send)
```

---

## Related docs

- [`docs/broadcast/scheduler-lock-decision.md`](../broadcast/scheduler-lock-decision.md)
- [`docs/reviews/12-0-broadcast-architecture-review.md`](12-0-broadcast-architecture-review.md)

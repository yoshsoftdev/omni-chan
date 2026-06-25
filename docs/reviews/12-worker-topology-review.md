# Worker Topology Review — Phase 12.0

**Status:** Design review  
**Reference:** [`docs/broadcast-execution-architecture.md`](../broadcast-execution-architecture.md)

## Proposed worker inventory

| Worker | Proposed role | Review verdict |
|--------|---------------|----------------|
| **Scheduler** | Poll `campaigns` due for send; create `broadcast_jobs`; publish `broadcast.job.created` | **Keep** — lightweight cron; can share process with orchestrator in v1 |
| **Audience Resolver** | Resolve audiences → `campaign_recipients` + snapshot | **Keep** — CPU/IO heavy; must not block send path |
| **Batch Planner** | Chunk recipients → `broadcast_batches`; enqueue send work | **Keep logically; merge deployable** — see below |
| **Send Worker** | Meta API calls; set `wamid`; recipient send lifecycle | **Keep** — horizontally scaled pool |
| **Status Consumer** | Webhook status → recipient updates | **Do not build new worker** — already implemented |
| **Report Aggregator** | Counter + daily stat updates | **Do not build new worker** — already implemented |
| **Completion Monitor** | Mark campaign/job complete | **Partially implemented** — extend, do not duplicate |

## Already implemented (Phase 11.4 / 11.6)

These run inside `backend-service` today:

```
Meta webhook → webhook-service → NATS meta.message.status.*
  → eventconsumer.StatusHandler
  → delivery.Tracker + ReportAggregator + CompletionMonitor
```

| Component | Location | Notes |
|-----------|----------|-------|
| Status Consumer | `internal/eventconsumer/meta_handlers.go` | Inbox + campaign dual path |
| Report Aggregator | `internal/campaign/delivery/aggregator.go` | Idempotent counter deltas |
| Completion Monitor | `internal/campaign/delivery/completion.go` | Triggered after status transition |
| Reconciliation | `internal/campaign/delivery/reconcile.go` | Daily counter repair |

**Recommendation:** Do not introduce parallel broadcast NATS consumers for delivery reporting. Phase 12 send workers only need to write `wamid` and `sent` status; webhooks drive the rest.

## Merge vs split recommendations

### Merge (deployment unit)

| Merge | Rationale |
|-------|-----------|
| **Scheduler + Audience Resolver + Batch Planner** → single `cmd/broadcast-orchestrator` process | Sequential pipeline per job; low message rate vs send workers; simplifies ops for v1 |
| **Completion check** into orchestrator tick (30s) + post-batch hook | Avoids extra consumer; DB query is cheap |

Keep **logical separation** (packages: `scheduler`, `audience`, `planner`) even if one binary.

### Split (required)

| Split | Rationale |
|-------|-----------|
| **Send Worker** separate from orchestrator | Different scale profile (N replicas); rate-limited; failure domain isolation |
| **Stale batch recovery** as cron in send supervisor OR orchestrator | Reclaims `locked_at` expired batches; not on hot path |

### Do not add

| Worker | Why |
|--------|-----|
| Dedicated Report Aggregator consumer | Inline in `delivery.Tracker` is simpler and idempotent |
| Per-recipient NATS publisher | 300k messages/campaign would flood JetStream; batch unit is correct |

## Recommended topology (v1)

```
┌─────────────────────────────────────┐
│  broadcast-orchestrator (1–3 pods)  │
│  Scheduler → Audience → Batch Plan  │
└──────────────┬──────────────────────┘
               │ broadcast.batch.send
               ▼
┌─────────────────────────────────────┐
│  broadcast-send-worker (N pods)     │
│  Rate limit → Meta API → wamid      │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│  backend-service (existing)         │
│  meta status consumer + delivery    │
└─────────────────────────────────────┘
```

## Additional worker: Stale Batch Recovery

Not in original list but **required**:

- Poll `broadcast_batches` where `status=processing` AND `locked_at < now() - lease`
- Reset to `queued`, republish `broadcast.batch.send` or `broadcast.batch.retry`
- Run every 5 minutes

## Open questions resolved

| Question | Answer |
|----------|--------|
| Can any be merged? | Yes — orchestration stages into one deployable |
| Should any be split? | Yes — send pool must scale independently |
| Status/report workers? | Reuse Phase 11 implementation |

## Decision

**Approved topology:**

1. `broadcast-orchestrator` — Scheduler, Audience Resolver, Batch Planner, Stale Recovery
2. `broadcast-send-worker` — Send pool only
3. No new delivery/report workers — extend existing `delivery` package if needed

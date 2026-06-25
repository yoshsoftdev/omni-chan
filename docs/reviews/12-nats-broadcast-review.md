# NATS Broadcast Subject Review — Phase 12.0

**Status:** Design review  
**Reference:** [`docs/broadcast-execution-architecture.md`](../broadcast-execution-architecture.md) §14

## Proposed subjects (Phase 12.0 brief)

| Subject | In brief |
|---------|----------|
| `broadcast.job.created` | ✓ |
| `broadcast.batch.created` | New |
| `broadcast.batch.ready` | New |
| `broadcast.batch.processing` | New |
| `broadcast.batch.completed` | New |
| `broadcast.recipient.status_changed` | New |
| `broadcast.campaign.completed` | New |

## Existing design (Phase 8)

| Subject | Purpose |
|---------|---------|
| `broadcast.job.created` | Start audience resolution |
| `broadcast.audience.resolved` | Trigger batch planning |
| `broadcast.batch.send` | Send worker work unit |
| `broadcast.batch.retry` | Delayed batch retry |
| `broadcast.job.completed` | Job finished (optional analytics) |

## Analysis

### Naming

| Rule | Recommendation |
|------|----------------|
| Prefix | `broadcast.` — consistent, separate from `meta.` webhook stream |
| Verb style | `{entity}.{action}` — e.g. `batch.send`, not `batch.ready` |
| Past tense for events | Reserve for **audit/analytics** only, not work queue |
| Version suffix | **No** — match `meta.*` stability policy ([`docs/events/nats-subjects.md`](../events/nats-subjects.md)) |

### Subject-by-subject verdict

| Subject | Verdict | Notes |
|---------|---------|-------|
| `broadcast.job.created` | **Approve** | Entry point for orchestration |
| `broadcast.batch.created` | **Reject as queue subject** | Planner creates DB rows; no consumer needs async notification |
| `broadcast.batch.ready` | **Merge into `broadcast.batch.send`** | "Ready" and "send" are the same work signal |
| `broadcast.batch.processing` | **Reject** | Processing is DB state (`locked_by`, `locked_at`), not a message |
| `broadcast.batch.completed` | **Reject as required subject** | Batch completion is DB update; optional metric event only |
| `broadcast.recipient.status_changed` | **Defer** | Phase 11.6 updates counters inline; add later for analytics fan-out |
| `broadcast.campaign.completed` | **Use `broadcast.job.completed`** | Job is execution unit; campaign may have multiple jobs over time |

### Recommended stream layout

**Stream:** `BROADCAST`  
**Retention:** 7 days  
**Storage:** file  
**Replicas:** 3 (production)

| Subject | Payload (minimal) | Publisher | Consumer |
|---------|-------------------|-----------|----------|
| `broadcast.job.created` | `{ job_id, campaign_id, team_id, channel_id }` | API, Scheduler | Orchestrator |
| `broadcast.audience.resolved` | `{ job_id, campaign_id, recipient_count }` | Audience Resolver | Batch Planner (same process OK) |
| `broadcast.batch.send` | `{ batch_id, job_id, campaign_id, channel_id, team_id }` | Batch Planner | Send Worker pool |
| `broadcast.batch.retry` | `{ batch_id, attempt, delay_ms, channel_id }` | Send Worker | Send Worker (delayed consumer) |
| `broadcast.job.completed` | `{ job_id, campaign_id, status }` | Orchestrator / Completion | Analytics (optional) |

### Separation from Meta stream

| Stream | Domain |
|--------|--------|
| `WEBHOOK_EVENTS` / `meta.>` | Inbound Meta webhooks (existing) |
| `BROADCAST` / `broadcast.>` | Internal campaign execution |

**Do not** route Meta status through `broadcast.recipient.status_changed`. Webhooks already flow:

```
meta.message.status.* → delivery.Tracker
```

## Future compatibility

| Extension | Approach |
|-----------|----------|
| Priority sends | `broadcast.batch.send.high` — new consumer binding |
| Multi-region | Stream export/import; subjects unchanged |
| v2 payload | New subject prefix `broadcast.v2.` only on breaking change |
| Per-channel fairness | Subject shard `broadcast.batch.send.{channel_id}` — optional Phase 12.2+ |

## Observability

| Metric | Source |
|--------|--------|
| `broadcast_nats_publish_total{subject}` | Publisher middleware |
| `broadcast_nats_consume_lag{consumer}` | JetStream consumer lag |
| `broadcast_batch_send_duration_seconds` | Send worker histogram |
| DLQ | `broadcast.dlq.{category}` — malformed payload, poison batch |

**Tracing:** propagate `job_id`, `batch_id`, `campaign_id` in NATS headers.

## Decision

**Approved subject set (minimal):**

```
broadcast.job.created
broadcast.audience.resolved
broadcast.batch.send
broadcast.batch.retry
broadcast.job.completed   (optional)
```

**Rejected as required queue subjects:** `batch.created`, `batch.ready`, `batch.processing`, `batch.completed`, `recipient.status_changed`, `campaign.completed`.

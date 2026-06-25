# Phase 12.6 — Multi-Node Orchestrator & Send Worker Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Horizontal scaling safety for orchestrator scheduler, job creation, audience snapshot, batch planning, and send worker batch claims

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Orchestrator Multi-Node Safety

### Scheduled campaign dispatch (`scheduled → queued`)

| Guard | Mechanism |
|-------|-----------|
| Claim due campaigns | `ClaimDueScheduled` — `FOR UPDATE SKIP LOCKED` |
| Status transition | `TransitionCampaignToQueued` — `WHERE status = 'scheduled'` |
| Second tick / replica | `RowsAffected = 0` → idempotent skip |

**Test:** `TestSchedulerPicksOnce`, `TestConcurrentSchedulerTick` (5 parallel ticks → 1 queue transition)

### Job creation (1 campaign → 1 active job)

| Guard | Mechanism |
|-------|-----------|
| Unique index | `idx_broadcast_jobs_one_active_per_campaign` — partial unique on `campaign_id` where status NOT IN (`completed`, `failed`, `cancelled`) |
| Lookup | `FindActiveJobByCampaign` before `CreateJob` |

### Audience snapshot (1 execution → 1 snapshot)

| Guard | Mechanism |
|-------|-----------|
| Job status progression | `resolving_audience` → snapshot created once per job |
| Recipient dedupe | Audience resolver deduplicates contact IDs |
| Idempotent re-entry | `CountRecipientsByJob` skip if recipients already exist |

**Test:** `TestAudienceResolverDeduplicates`

### Batch planning (250 recipients → 3 batches @ 100)

| Guard | Mechanism |
|-------|-----------|
| Plan once | `PlanBatchesForJob` checks existing batch count inside transaction |
| Chunk size | Configurable `BROADCAST_BATCH_SIZE` (default 100) |

**Test:** `TestBatchPlannerChunking`

### Outbox idempotency

| Guard | Mechanism |
|-------|-----------|
| Job created event | One outbox row per successful dispatch; relay marks `published_at` |
| NATS consumer | Durable `broadcast-orchestrator`; explicit ACK |

---

## Send Worker Multi-Node Safety

### Batch ownership (`queued → processing`)

| Guard | Mechanism |
|-------|-----------|
| Conditional claim | `ClaimBatchForProcessing` — `WHERE status IN ('queued','processing')` + lease |
| Stale lease | `locked_at < staleBefore` allows re-claim (10m default) |
| In-progress | `ErrBatchInProgress` → worker returns nil (NATS redelivery safe) |

### Recipient guarantees (1 recipient → 1 WAMID)

| Guard | Mechanism |
|-------|-----------|
| Status transition | `queued → sending → sent` conditional updates |
| WAMID skip | `skipRecipient` if WAMID already set |
| Unique index | `idx_campaign_recipients_wamid` partial unique |
| Transactional finalize | `FinalizeRecipientSent` single tx |

### Stale recovery

| Layer | Mechanism |
|-------|-----------|
| Batch | Orchestrator `Recovery` requeues `processing` batches past lease |
| Recipient | Worker `RecipientRecovery` — `sending` + `wamid IS NULL` + stale `updated_at` → `queued` |

---

## Findings

### LOW

1. **Concurrent scheduler test uses in-memory store** — Production safety relies on PostgreSQL `SKIP LOCKED` + status guards; recommend testcontainers integration test in Phase 13.

2. **JetStream consumer scaling** — Multiple orchestrator replicas share durable consumer `broadcast-orchestrator`; NATS delivers each message to one consumer. Queue group semantics validated by design; load test recommended before production scale-out.

---

## Exit Criteria

- [x] Multi-node orchestrator guards documented and unit-tested
- [x] Multi-node send worker lease + idempotency documented
- [x] No duplicate job index bypass path identified

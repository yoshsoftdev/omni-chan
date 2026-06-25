# Scheduler Lock Decision — Phase 12.1

## Decision

**Primary:** PostgreSQL row locks (`FOR UPDATE SKIP LOCKED` + conditional `UPDATE`)  
**Secondary:** Unique partial index on active `broadcast_jobs` per campaign  
**Optional:** Redis leader election for scheduler tick (deferred)

## Rationale

| Approach | Verdict |
|----------|---------|
| Postgres `SKIP LOCKED` | **Selected** — correctness without extra infra dependency |
| Redis distributed lock | Optional optimization to reduce duplicate ticks; not required for idempotency |
| NATS consumer alone | Insufficient — scheduler originates work from DB poll |

## Idempotency layers

1. `ClaimDueScheduled` uses `FOR UPDATE SKIP LOCKED` so two schedulers cannot claim the same row concurrently.
2. `TransitionCampaignToQueued` requires `status = scheduled` — second attempt returns `ErrNotFound`.
3. `idx_broadcast_jobs_one_active_per_campaign` prevents duplicate active jobs.
4. `ProcessJob` / audience / planner are idempotent on `job_id` (recipient count, batch count guards).
5. `broadcast_outbox` + relay tolerates publish retries without duplicate side effects on DB.

## Restart safety

On orchestrator restart:

- Unpublished outbox rows are relayed to NATS.
- `ListPendingJobs` sweep re-processes jobs stuck in `pending` / `resolving_audience` / `planning_batches`.
- No duplicate batches when `PlanBatchesForJob` finds existing batches for job.

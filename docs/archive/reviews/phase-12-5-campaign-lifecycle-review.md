# Phase 12.5 — Campaign Lifecycle Completion & Operational Readiness Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Campaign send/cancel APIs, lifecycle state machine, orchestrator integration, transactional failure counters, completion monitor — no frontend, webhook, or orchestration redesign

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Phase 12.4 Finding Validation

| Prior Finding | Status |
|---------------|--------|
| 12.4 #1 — `markFailed` not transactional (`failed_count` drift) | **Resolved** — `FinalizeRecipientFailed` single transaction |
| 12.4 #2 — No PG rollback integration test for finalizer | Unchanged (design + idempotent guards) |
| 12.4 #3 — Redis optional for worker `healthy` | Unchanged |

---

## API Readiness Audit

| Endpoint | Before | After |
|----------|--------|-------|
| `GET /api/v1/campaigns` | Production `CampaignService.List` | Unchanged — production |
| `GET /api/v1/campaigns/:id` | Production `CampaignService.Get` | Unchanged — production |
| `POST /api/v1/campaigns/:id/send` | `501 FeatureNotAvailable` stub | **Wired** → `Service.Send` → outbox `broadcast.job.created` |
| `POST /api/v1/campaigns/:id/cancel` | `501 FeatureNotAvailable` stub | **Wired** → `Service.Cancel` → `CancelCampaignExecution` |

### Send flow

```
POST /api/v1/campaigns/:id/send
  → CampaignHandler.Send
  → campaign.Service.Send (RBAC: campaign.send)
  → TransitionCampaignToQueuedForSend (draft|scheduled → queued)
  → FindActiveJobByCampaign / CreateJob
  → EnqueueBroadcastJobCreated → broadcast_outbox
  → orchestrator OutboxRelay → NATS broadcast.job.created
  → audience → batches → send worker
```

### Cancel flow

```
POST /api/v1/campaigns/:id/cancel
  → CampaignHandler.Cancel
  → campaign.Service.Cancel (RBAC: campaign.cancel)
  → CancelCampaignExecution (single transaction)
      campaigns.status → cancelled
      recipients pending|queued|sending(no wamid) → cancelled
      batches pending|queued|processing → cancelled
      broadcast_jobs active → cancelled
```

---

## Campaign State Machine Audit

Explicit guards in `internal/campaign/lifecycle.go`:

| Transition | Allowed |
|------------|---------|
| `draft → scheduled` | ✓ (create with `send_mode=later`) |
| `draft → queued` | ✓ (`POST /send`) |
| `draft → cancelled` | ✓ (`POST /cancel`) |
| `scheduled → queued` | ✓ (scheduler tick or `POST /send`) |
| `scheduled → cancelled` | ✓ |
| `queued → sending` | ✓ (worker `TransitionCampaignToSending`) |
| `queued → cancelled` | ✓ |
| `sending → completed` | ✓ (completion monitor) |
| `sending → failed` | ✓ (orchestrator/job failure paths) |
| `sending → cancelled` | ✓ |
| `draft → completed` | **Rejected** — no code path |
| `queued → completed` (direct) | **Rejected** — requires completion monitor |
| `completed → *` | **Rejected** — terminal |
| `cancelled → *` | **Rejected** — terminal |

Service-layer validation rejects `POST /send` on `queued`, `sending`, `completed`, `failed`, `cancelled`.  
`POST /cancel` on `completed`/`failed` returns `409 Conflict`; already `cancelled` is idempotent.

---

## Recipient Cancellation Audit (Phase 8)

| Status | May cancel? | Implementation |
|--------|-------------|----------------|
| `pending` | Yes → `cancelled` | `CancelCampaignExecution` SQL |
| `queued` | Yes → `cancelled` | ✓ |
| `sending` (no wamid) | Yes → `cancelled` | `wamid IS NULL` guard |
| `sent` / `delivered` / `read` | **No** | Not in cancel UPDATE |
| `failed` / `cancelled` / `skipped` | No (terminal) | Unchanged |

`CanCancelRecipient()` mirrors Phase 8 rules for tests and documentation.

---

## Failed Counter Consistency (12.4 closure)

`campaign.FinalizeRecipientFailed` runs in one PostgreSQL transaction:

1. `campaign_recipients`: `sending|queued → failed` + `failure_code` + `last_error` + `failed_at`
2. `campaigns.failed_count += 1`
3. `campaign_daily_stats.messages_failed += 1`

Worker `markFailed` uses finalizer when configured (production `PostgresRepository`).  
Idempotent retry: if recipient already `failed`, returns `applied=false` without counter increment.

---

## Completion Monitor Audit

`delivery.CompletionMonitor.Check`:

1. `CountNonTerminalRecipients(campaignID) == 0`
2. `CompleteCampaign(teamID, campaignID, at)` — sets `status = completed`, `completed_at = at`
3. `ErrNotFound` on second call → idempotent (already completed)

`CountNonTerminalRecipients` excludes: `read`, `failed`, `cancelled`, `skipped`.  
`sent` and `delivered` count as **non-terminal** until read webhook or reconciliation — campaigns remain `sending` until all recipients reach a terminal state.

Test: `delivery/completion_test.go` — `TestCompletionMonitorCompletesOnce`.

---

## Deliverables Checklist

| Component | Status |
|-----------|--------|
| `internal/campaign/lifecycle.go` — state machine | Done |
| `internal/campaign/lifecycle_repo.go` — send transition + cancel tx | Done |
| `internal/campaign/service_lifecycle.go` — `Send` / `Cancel` | Done |
| `internal/campaign/outbox_enqueue.go` — job.created outbox | Done |
| `internal/campaign/send_finalize.go` — `FinalizeRecipientFailed` | Done |
| `internal/api/v1/handler/campaign.go` — wire Send/Cancel | Done |
| `internal/rbac/permissions.go` — `PermCampaignCancel` | Done |
| Create sets `scheduled` when `send_mode=later` | Done |
| Tests (lifecycle, service, completion) | Done |

---

## Tests

| Test | Coverage |
|------|----------|
| `TestCanTransitionCampaignStatus` | Legal + illegal campaign transitions |
| `TestCanCancelRecipient` | Phase 8 recipient cancel rules |
| `TestServiceSendQueuesCampaign` | Send → queued + outbox enqueue |
| `TestServiceSendRejectsCompleted` | Illegal send rejected |
| `TestServiceCancelIdempotent` | Already cancelled → no-op |
| `TestServiceCancelRejectsCompleted` | Completed → conflict |
| `TestCompletionMonitorCompletesOnce` | Single completion + idempotent retry |
| Worker / orchestrator suites | Still pass with updated stubs |

---

## Findings

### MEDIUM

1. **Redis cancel flag not implemented** — Phase 8 design specifies `cancel:campaign:{id}` (TTL 24h) for fast worker polling between recipients. Current implementation is **Postgres-authoritative only**. Workers skip `queued` recipients already marked `cancelled` in DB, but in-flight `sending` recipients may still complete if Meta returns `wamid` after cancel (documented race in `12-cancellation-review.md`). Recommend Redis flag in Phase 13 operational hardening.

2. **Worker does not reload campaign status mid-batch** — Cancel updates DB; worker loads campaign once at batch start. Remaining `queued` recipients are skipped on status check; `sending` in-flight handled per recipient transition rules.

### LOW

3. **OpenAPI `v1.yaml` still documents `501` for cancel** — Handler is live; spec drift. Update OpenAPI in a docs pass.

4. **`sent`/`delivered` non-terminal for completion** — Campaigns stay `sending` until all recipients reach `read`, `failed`, `cancelled`, or `skipped`. Relies on Phase 11.6 webhooks + daily reconciliation. Expected behavior; document in runbooks.

5. **No integration test for `CancelCampaignExecution` against real PostgreSQL** — Transaction SQL validated by design; full PG test requires testcontainers.

6. **Legacy `campaign.Worker` + Redis queue still runs in `cmd/api`** — Parallel to NATS orchestration path. Not removed in 12.5; confirm deprecation timeline before Phase 13 multi-node deployment.

---

## Recommendation

**APPROVE WITH FINDINGS**

Phase 12.5 completes the campaign lifecycle API surface: send and cancel are production-wired through `CampaignService` → orchestrator outbox → broadcast job pipeline. Illegal state jumps are rejected. Cancellation matches Phase 8 recipient rules at the database layer. The Phase 12.4 failed-counter gap is closed with transactional `FinalizeRecipientFailed`. Completion monitor is idempotent.

Safe to proceed to **Phase 13 — Operational Excellence** with the understanding that Redis cancel flags, OpenAPI sync, and legacy worker deprecation should be addressed early in that phase.

---

## Out of Scope (confirmed unchanged)

Frontend, webhook consumers, delivery tracking redesign, orchestration architecture redesign, Prometheus/Grafana (Phase 13)

# Phase 12.6 — Legacy Execution Path Decommission Review

**Verdict: APPROVE**

**Date:** 2026-06-22  
**Scope:** Remove Redis-queue campaign worker and in-process scheduler from `cmd/api`; confirm NATS JetStream as sole broadcast transport

---

## Audit Summary

| Target | Before | After |
|--------|--------|-------|
| `internal/campaign/worker.go` (`//go:build legacy`) | Redis queue consumer, direct Meta send | **Removed** |
| `internal/campaign/scheduler.go` (`//go:build legacy`) | In-process scheduled dispatch | **Removed** |
| `internal/campaign/worker_compat.go` | No-op stub in `cmd/api` | **Removed** |
| `internal/campaign/scheduler_compat.go` | No-op stub in `cmd/api` | **Removed** |
| `internal/campaign/render.go`, `variables.go` | Legacy-only template render | **Removed** |
| `pkg/campaignqueue` | Redis `LPUSH`/`BRPOP` queue | **Removed** |
| `cmd/api` background jobs | `campaignWorker.Start`, `campaignScheduler.Start` | **Removed** |

---

## Active Execution Path (sole)

```text
POST /campaigns/:id/send  OR  orchestrator scheduler tick
    → broadcast_outbox (broadcast.job.created)
    → NATS JetStream
    → broadcast-orchestrator (audience → batches → batch.send)
    → NATS JetStream
    → broadcast-worker (Meta API send)
```

Dedicated processes:

- `cmd/broadcast-orchestrator` — scheduler, outbox relay, job pipeline, batch recovery
- `cmd/broadcast-worker` — batch claim, send, recipient recovery
- `cmd/api` — event consumer (webhooks → delivery tracking); **no broadcast send**

---

## Validation

| Check | Result |
|-------|--------|
| No production route invokes legacy worker | PASS — worker deleted |
| No scheduler duplication in `cmd/api` | PASS — stub removed |
| No Redis queue dependency for broadcasts | PASS — `pkg/campaignqueue` deleted |
| Duplicate recipient generation | N/A — legacy path removed; NATS path uses audience snapshot + dedupe |
| Duplicate send execution | N/A — legacy path removed; worker uses batch lease + WAMID idempotency |

---

## Exit Criteria

- [x] Legacy worker removed (not merely deprecated)
- [x] NATS JetStream is sole broadcast transport
- [x] No active code path references `campaignqueue` or legacy worker/scheduler
- [x] `go build ./...` PASS
- [x] `go test ./...` PASS

---

## Notes

- `//go:build legacy` sources were deleted outright rather than kept behind a build tag to prevent accidental re-enablement.
- `cmd/api` still runs the **event consumer** (NATS meta/legacy webhook subjects) — this is delivery tracking, not broadcast send.

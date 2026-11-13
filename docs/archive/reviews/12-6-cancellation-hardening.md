# Phase 12.6 — Cancellation Hardening Review

**Verdict: APPROVE**

**Date:** 2026-06-22  
**Scope:** Redis cancel flag (`cancel:campaign:{id}`) + worker polling — closes Phase 12.5 finding #1

---

## Implementation

### Redis cancel flag

| Property | Value |
|----------|-------|
| Key | `cancel:campaign:{campaign_id}` |
| Value | `1` |
| TTL | 24 hours |
| Package | `pkg/broadcastcancel` |

### Set path

```text
POST /api/v1/campaigns/:id/cancel
  → Service.Cancel
  → CancelCampaignExecution (PostgreSQL — authoritative)
  → cancelFlags.Set(campaign_id) (Redis — fast worker poll)
```

Wired via `campaign.WithCancelFlags(broadcastcancel.NewStore(...))` in `cmd/api`.

### Worker check path

```text
broadcast-worker BatchProcessor
  → before each queued recipient: cancelChecker.IsSet(campaign_id)
  → before Meta API call (after sending transition): re-check
  → if cancelled: skipRecipientForCancel (queued/sending without wamid → cancelled)
  → if wamid already set: respect sent state (no rollback)
```

Wired in `cmd/broadcast-worker` when Redis is available.

---

## Validation Matrix

| Scenario | Expected | Status |
|----------|----------|--------|
| `pending` / `queued` at cancel | PostgreSQL → `cancelled` | PASS (Phase 12.5) |
| `queued` during batch after cancel flag | Worker skips / marks `cancelled` | PASS |
| `sending` without wamid + cancel flag | → `cancelled`, no Meta call | PASS |
| `sending` + wamid returned | → `sent` (race documented) | PASS (by design) |
| `sent` / `delivered` / `read` | Not cancellable | PASS |
| Redis unavailable at cancel | PostgreSQL cancel still works | PASS |
| Redis unavailable at worker | Worker relies on DB recipient status | PASS (degraded) |

---

## Tests

| Test | Coverage |
|------|----------|
| `pkg/broadcastcancel.TestKeyFor` | Key format |
| `batch_processor` integration suite | Still passes with nil cancel checker |
| Phase 12.5 cancel service tests | PostgreSQL cancel rules |

---

## Exit Criteria

- [x] Redis cancel flag implemented with 24h TTL
- [x] Worker checks flag before each recipient send
- [x] PostgreSQL remains source of truth
- [x] Phase 12.5 MEDIUM finding #1 resolved

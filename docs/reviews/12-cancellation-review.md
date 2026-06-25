# Cancellation Review — Phase 12.0

**Status:** Design review  
**Trigger:** `POST /api/v1/campaigns/{id}/cancel`  
**Reference:** DB trigger `enforce_recipient_status_monotonic()` in migration `000012`

## States in scope

| State | In brief | Review |
|-------|----------|--------|
| `pending` | Not yet batched | ✓ |
| `queued` | In batch, not sending | ✓ |
| `sending` | API in flight | ✓ (conditional) |
| `sent` | Meta accepted, `wamid` set | ✓ (cannot cancel) |

## Recipient cancellation rules

| Current status | May cancel? | Result status | Rationale |
|----------------|-------------|---------------|-----------|
| `pending` | **Yes** | `cancelled` | Not yet scheduled for send |
| `queued` | **Yes** | `cancelled` | Worker has not called Meta |
| `sending` (no `wamid`) | **Yes** | `cancelled` | API not completed; worker checks flag before call |
| `sending` (`wamid` set) | **No** | `sent` | Message already accepted by Meta |
| `sent` | **No** | unchanged | Cannot recall WhatsApp message |
| `delivered` | **No** | unchanged | Terminal delivery path |
| `read` | **No** | unchanged | Terminal |
| `failed` | **No** | unchanged | Already terminal |
| `skipped` | **No** | unchanged | Never intended to send |
| `cancelled` | **No** | unchanged | Idempotent |

### DB alignment

The monotonic trigger allows:

- `pending → cancelled`
- `queued → cancelled`
- `sending → cancelled` (before `sent`)

It **does not** allow `sent → cancelled`. Implementation matches product expectation.

## Batch cancellation rules

| Batch status | Behavior |
|--------------|----------|
| `pending` | → `cancelled` |
| `queued` | → `cancelled`; NATS message acked without processing |
| `processing` | Worker checks cancel flag between recipients; remaining → `cancelled` |
| `completed` | unchanged |
| `failed` | unchanged |
| `cancelled` | idempotent |

## Job and campaign

| Entity | On cancel |
|--------|-----------|
| `campaigns.status` | → `cancelled` |
| `broadcast_jobs.status` | → `cancelled`, `cancelled_at`, `cancel_reason` |
| Active NATS work | Send worker no-ops when flag set |

## Cancel signal propagation

| Layer | Mechanism |
|-------|-----------|
| PostgreSQL | Authoritative `campaigns.status = cancelled` |
| Redis | `cancel:campaign:{id} = 1`, TTL 24h — fast worker poll |
| In-flight send | Check flag **before each** Meta API call |

### Race: cancel during Meta request

```
Worker starts send → cancel arrives → Meta returns 200 + wamid
```

**Resolution:** If `wamid` received, recipient becomes `sent` (not `cancelled`). Campaign may end `cancelled` with partial sends — expected behavior.

## API behavior

| Recipient mix at cancel | Campaign outcome |
|-------------------------|------------------|
| All `pending`/`queued` | `cancelled`, zero sends |
| Some `sent` | `cancelled`, partial send counts in reports |
| All terminal | Reject cancel (409) or no-op |

## Send worker cancel check (pseudocode)

```
for each recipient in batch:
  if redis.Exists("cancel:campaign:{id}"):
    mark remaining recipients cancelled
    break
  if recipient.wamid != nil: continue
  send to Meta...
```

## Decision

| State | Cancels? |
|-------|----------|
| `pending` | Yes |
| `queued` | Yes |
| `sending` (no wamid) | Yes |
| `sent` / `delivered` / `read` | **No** |

**Approved** with Redis cancel flag + per-recipient checks in send worker.

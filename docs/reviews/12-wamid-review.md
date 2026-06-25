# WAMID Correlation Review — Phase 12.0

**Status:** Design review  
**Reference:** [`docs/broadcast/wamid-correlation.md`](../broadcast/wamid-correlation.md)

## Required flow (Phase 12)

```
Meta POST /messages response
  → messages[0].id (wamid)
  → UPDATE campaign_recipients SET wamid, status=sent
  → (later) meta.message.status.* webhooks
  → delivery.Tracker.FindRecipientByWAMIDForTeam
```

**Current gap:** Send worker does not exist (`worker_compat.go` stub). Phase 11.6 delivery path is ready but receives no `wamid` until send is implemented.

## Write path (send worker — Phase 12.1+)

| Step | Requirement |
|------|-------------|
| 1 | Transition `queued → sending` with `WHERE status = queued` |
| 2 | Call Meta template API |
| 3 | On 200: `UPDATE ... SET wamid = $1, status = sent, sent_at = now() WHERE id = $2 AND status = sending` |
| 4 | On failure: classify → retry or `failed` |

**Atomicity:** `wamid` and `sent` status must be written in same transaction as counter increment for `sent_count` (or rely on webhook-only counters — **not recommended**; increment on send).

### Counter on send

Phase 11.6 `ReportAggregator` increments on status transition. Send worker should either:

- Call `delivery.ReportAggregator.ApplyTransition(..., sending, sent, ...)` after DB update, **or**
- Publish internal event consumed by aggregator

**Recommendation:** Direct aggregator call from send worker package to avoid new NATS subject.

## Read path (webhook — implemented)

| Requirement | Status |
|-------------|--------|
| O(1) lookup | ✓ `idx_campaign_recipients_wamid` |
| Tenant isolation | ✓ `FindRecipientByWAMIDForTeam(team_id, wamid)` |
| Monotonic status | ✓ DB trigger + `delivery.CanTransition` |
| Duplicate webhooks | ✓ Consumer idempotency + `WHERE status = from` |

## Index review

```sql
CREATE UNIQUE INDEX idx_campaign_recipients_wamid
  ON campaign_recipients (wamid, created_at)
  WHERE wamid IS NOT NULL;
```

| Property | Assessment |
|----------|------------|
| Uniqueness | Prevents duplicate wamid assignment |
| Partition key | `(wamid, created_at)` supports partitioned table |
| Lookup query | `WHERE wamid = $1 AND team_id = $2` — add composite index if explain shows seq scan at scale |

**Optional Phase 12.1 migration:**

```sql
CREATE INDEX idx_campaign_recipients_team_wamid
  ON campaign_recipients (team_id, wamid)
  WHERE wamid IS NOT NULL;
```

Defer until load testing; unique index may suffice.

## Dual-path correlation

| Table | wamid source | Consumer |
|-------|--------------|----------|
| `messages` | Inbound + outbound inbox | `message.ApplyDeliveryStatus` |
| `campaign_recipients` | Broadcast send worker | `delivery.Tracker` |

Same `wamid` should not appear in both for broadcast messages if architecture keeps inbox and campaign sends separate. If overlap occurs, both handlers no-op on miss — no conflict if only one row exists.

## Reconciliation support

| Mechanism | Role |
|-----------|------|
| `wamid` on recipient | Links send to webhook |
| `RepairCampaignCounters` | Fixes counter drift without wamid |
| Missing webhook | Recipient stays `sent`; nightly reconcile OK |

**Stuck `sending` without wamid:** Stale batch recovery resets to `queued` for retry.

## Failure modes

| Scenario | Handling |
|----------|----------|
| Meta 200 but DB write fails | Retry send; Meta may return same wamid or error — log for manual reconcile |
| Duplicate send attempt | Check `wamid IS NOT NULL` before API |
| Webhook before DB commit | Rare; retry webhook idempotency handles |

## Decision

| Requirement | Verdict |
|-------------|---------|
| Fast lookup | **Approved** — existing index |
| Tenant isolation | **Approved** — team-scoped query |
| Reconciliation | **Approved** — counters + wamid audit |
| Send path | **Phase 12.1 must write wamid atomically with `sent` status** |

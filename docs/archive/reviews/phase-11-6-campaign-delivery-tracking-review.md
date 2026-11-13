# Phase 11.6 — Campaign Delivery Tracking Review

**Verdict: APPROVE WITH FINDINGS**

## Recipient Lifecycle

- `delivery.Tracker` maps canonical status events to `campaign.RecipientStatus`
- Monotonic rules mirror DB trigger `enforce_recipient_status_monotonic()`
- Duplicate webhooks and invalid regressions (e.g. `read → delivered`) are ignored
- Tenant isolation via `FindRecipientByWAMIDForTeam(team_id, wamid)`

## Counter Accuracy

- `ReportAggregator` increments `campaigns.*_count` only on successful status transition
- `CounterDeltaForTransition` ensures one increment per funnel stage
- Consumer idempotency (`provider:wamid:status`) prevents duplicate processing
- `WHERE status = $from` on update provides row-level idempotency

## Aggregation

- Daily stats updated via `report.IncrementDailyStat` upsert on each transition
- Dashboard and funnel APIs read existing `campaigns` counters and `campaign_daily_stats` — now fed by real delivery events

## Completion Detection

- `CompletionMonitor` marks campaign `completed` when `CountNonTerminalRecipients` returns 0
- Terminal states: `read`, `failed`, `cancelled`, `skipped`
- Sets `completed_at` for campaigns in `sending`, `queued`, or `scheduled`

## Reconciliation

- Daily `Reconciler` repairs counter drift from recipient aggregates
- `RepairCampaignCounters` updates only when values differ

## Findings

1. **Broadcast worker stub** — `campaign/worker_compat.go` is a no-op; recipients receive `wamid` only when Phase 12 send workers run. Delivery tracking is ready but requires send path to populate `wamid`.
2. **CTR** — Click tracking unchanged; CTR in daily stats depends on link click events (existing report path).
3. **Failed error parsing** — Relies on Meta `raw` payload structure; extend when webhook normalizer adds top-level error fields.

## Recommendation

Approve for merge. Enable end-to-end validation once Phase 12 sets `wamid` on send. Monitor `campaign_status_updates_total` and `reconciliation_repairs_total` in staging with real Meta webhooks.

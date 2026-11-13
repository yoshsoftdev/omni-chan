# Campaign Counter Reconciliation

Daily job repairs drift between `campaign_recipients` status aggregates and `campaigns` counter columns.

## Purpose

Webhook retries, partial failures, or manual DB edits can desync:

- `campaigns.sent_count`
- `campaigns.delivered_count`
- `campaigns.read_count`
- `campaigns.failed_count`

Reconciliation rebuilds counters from recipient rows.

## Algorithm

```
FOR each campaign updated in lookback window (default 30 days):
  agg = GROUP BY status on campaign_recipients
  IF campaign counters != agg:
    SET campaigns counters = agg
    INCREMENT reconciliation_repairs_total
```

## Aggregate rules

| Counter | Recipient statuses counted |
|---------|---------------------------|
| `sent_count` | `sent`, `delivered`, `read` |
| `delivered_count` | `delivered`, `read` |
| `read_count` | `read` |
| `failed_count` | `failed` |

## Schedule

`delivery.Reconciler.StartDaily` runs:

1. Once at startup
2. Every 24 hours until process shutdown

## Implementation

- `campaign.Repository.RepairCampaignCounters` — single SQL CTE + conditional update
- `campaign.Repository.ListCampaignIDsForReconciliation` — active/recent campaigns

## Observability

- Metric: `reconciliation_repairs_total`
- Log: `campaign_id`, repaired counter values (no message content)

# Template Sync Strategy

## Meta API Constraints

Meta Graph API `message_templates` does **not** support reliable incremental sync by modified timestamp. Omni-Chan uses:

```text
Full paginated fetch
    +
Local change detection (drift)
```

## Per-Channel Freshness

| Signal | Source | Use |
|--------|--------|-----|
| `templates.last_sync_at` | Per template row | Fallback freshness check |
| `SyncRun.completed_at` | Sync run audit | Primary freshness for health/readiness |

## Incremental Behavior

On each sync:

1. Fetch **all pages** from Meta (cursor-based `after` pagination)
2. For each Meta template:
   - **Create** if `meta_template_id` not found locally
   - **Update** if name, language, status, components, or preview drift
   - **Skip count** if unchanged (still upserted for idempotency)
3. **Soft-delete** local templates whose `meta_template_id` is absent from Meta response

## Idempotency

Unique key: `(channel_id, meta_template_id)` — repeated syncs produce identical database state.

## Freshness Rules

| Check | Threshold |
|-------|-----------|
| Channel health `template_sync` | Last successful sync < 24h |
| Workspace `templates_recently_synced` | Last successful sync ≤ 24h |

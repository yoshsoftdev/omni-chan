# Template Sync Run Model

Each template synchronization execution is tracked as a **Sync Run**.

## Entity: `SyncRun`

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `team_id` | UUID | Workspace scope |
| `channel_id` | UUID | Sync target channel |
| `started_at` | TIMESTAMPTZ | Execution start |
| `completed_at` | TIMESTAMPTZ | Execution end (nullable while running) |
| `status` | enum | `running`, `success`, `failed` |
| `synced_count` | int | Total templates processed |
| `created_count` | int | New local templates |
| `updated_count` | int | Drift-updated templates |
| `removed_count` | int | Soft-deleted templates |
| `error_count` | int | Per-template upsert failures |
| `meta_cursor` | string | Last Graph API pagination cursor |
| `error_message` | text | Failure reason when `status = failed` |

## Lifecycle

```text
POST /templates/sync
    ↓
Create SyncRun (running)
    ↓
Paginated Meta fetch + drift upsert
    ↓
Soft-delete removed templates
    ↓
Update SyncRun (success | failed)
```

## Storage (Phase 11.1)

In-memory repository (`InMemorySyncRunRepository`) until `template_sync_runs` migration lands.

See `docs/reviews/template-sync-schema-gap.md`.

## API History Contract

```json
{
  "id": "uuid",
  "channel_id": "uuid",
  "channel_name": "Main WABA",
  "status": "success",
  "started_at": "2026-06-22T10:00:00Z",
  "completed_at": "2026-06-22T10:00:02Z",
  "created": 4,
  "updated": 8,
  "removed": 1
}
```

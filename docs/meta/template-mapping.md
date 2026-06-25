# Template Mapping

Meta Graph API templates are mapped to Omni-Chan domain templates via `template.MapFromMeta`.

## Flow

```text
Meta ListTemplates (Graph API)
    ↓
template.MapFromMeta
    ↓
template.Repository.UpsertFromSync
```

## Status Normalization

| Meta | Domain (`templates.status`) |
|------|----------------------------|
| `APPROVED` | `APPROVED` |
| `REJECTED` | `REJECTED` |
| `PENDING` / other | `PENDING` |

## Language Normalization

| Meta input | Stored |
|------------|--------|
| `en` | `en_EN` |
| `en_US` | `en_US` |
| `id` | `id_ID` |

## Component Support

Mapper preserves raw `components` JSON from Meta. Supported component types in preview/variable detection:

- `HEADER` (text, image, video, document)
- `BODY`
- `FOOTER`
- `BUTTONS` — `QUICK_REPLY`, `URL`, `COPY_CODE`

Variable count derived from `{{n}}` patterns in serialized components JSON.

## Idempotency

- Upsert key: `(channel_id, meta_template_id)` unique index
- Repeated sync updates existing rows, does not duplicate
- Templates removed from Meta are soft-deleted via `SoftDeleteExceptMetaIDs`

## API

```http
POST /api/v1/templates/sync?channel_id={uuid}
```

Returns:

```json
{
  "synced": 12,
  "removed": 1
}
```

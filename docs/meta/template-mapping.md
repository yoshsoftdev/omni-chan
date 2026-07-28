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

Handled exclusively by `template.LanguageResolver` (Phase 5).

Meta treats `en` and `en_US` as **distinct** `language.code` values. Coercing `en` → `en_US`
on send causes Meta `#132001` (template does not exist in the translation).

| Meta / input | Stored | Meta send (`language.code`) |
|--------------|--------|------------------------------|
| `en` | `en` | `en` |
| `en_US` | `en_US` | `en_US` |
| `en_EN` (legacy artifact) | `en_US` | `en_US` |
| `id` | `id` | `id` |
| `id_ID` | `id` | `id` |
| other `xx_YY` | case-normalized | same |
| other bare `xx` | lowercase `xx` (no invented region) | same |

Never invent invalid regions such as `en` → `en_EN`. Never coerce `en` ↔ `en_US` on send.
UI language filters may treat `en` / `en_US` as the English family for search only.

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

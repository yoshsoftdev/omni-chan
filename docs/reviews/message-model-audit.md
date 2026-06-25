# Message Model Audit

**Date:** 2026-06-22  
**Scope:** `messages` table + `raw_payload` JSONB

## Supported Today

| Capability | Storage |
|------------|---------|
| Text body | `content` |
| Media types | `content_type` + preview in `content` |
| Raw Meta payload | `raw_payload` JSONB |
| WAMID idempotency | `wamid` unique index |
| Direction | `direction` inbound/outbound |

## Gaps (Future Phases)

| Capability | Status | Recommendation |
|------------|--------|----------------|
| Reply context (`context.message_id`) | Not normalized | Parse from `raw_payload`; add `reply_to_wamid` column |
| Quoted messages | Not normalized | Store in `raw_payload`; optional `quoted_wamid` |
| Interactive replies (buttons/lists) | Not normalized | Add `interactive_type` + `interactive_payload` JSONB |
| Button replies | Not normalized | Extract from `raw_payload.interactive` |
| List replies | Not normalized | Extract from `raw_payload.interactive` |
| Media captions | Partial | Caption in `content`; full media in `raw_payload` |
| Reactions | Not supported | New `message_reactions` table or event stream |

## Verdict

Phase 11.5 stores full Meta payloads in `raw_payload` for forward compatibility. Structured fields for reply/interactive/reactions are deferred.

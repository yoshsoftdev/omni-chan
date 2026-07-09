# Inbox Reply & Delete Sync Review

## Scope

Implemented inbox reply and delete synchronization foundations:
- Reply context out/in with jump-to-source UX.
- Hover message actions menu (`Reply`, `Delete`) in inbox bubble.
- Soft-delete lifecycle in backend + webhook deleted status ingestion + app delete endpoint.

## Behavior

- Normal send flow stays unchanged when no reply target is selected.
- Clicking `Reply` on a message with WAMID sets reply target.
- Composer shows a small reply preview above the input.
- Reply target can be cancelled from composer.
- When reply target exists, backend sends WhatsApp Cloud API payload with `context.message_id`.
- Inbound webhook/application-layer reply linkage stores `reply_to_wamid` when provided by Meta.
- Reply resolution persists `reply_to_message_id` when source message exists in the same team.
- Message bubble quote can jump and highlight the original message.
- Message hover menu provides WhatsApp-like actions (`Reply`, `Delete`).
- Deleted messages render placeholder text (`This message was deleted`) instead of hard removal.
- Backend supports deleting from app with provider revoke attempt (best effort) and local soft-delete fallback.
- Webhook status `deleted` is normalized and consumed into message soft-delete state.

## Modified Files

- `frontend/src/features/inbox/types.ts`
- `frontend/src/features/inbox/mappers/inbox.mapper.ts`
- `frontend/src/features/inbox/components/message-bubble.tsx`
- `frontend/src/features/inbox/components/message-composer.tsx`
- `frontend/src/features/inbox/components/conversation-detail.tsx`
- `frontend/src/features/inbox/components/inbox-page.tsx`
- `frontend/src/features/inbox/services/message.service.ts`
- `frontend/src/types/message.ts`
- `services/backend-service/internal/api/v1/handler/message.go`
- `services/backend-service/internal/message/outbound.go`
- `services/backend-service/pkg/meta/client.go`
- `services/backend-service/pkg/meta/media.go`
- `services/webhook-service/internal/webhook/meta/normalizer/normalizer.go`
- `services/webhook-service/internal/webhook/meta/normalizer/normalizer_test.go`
- `shared/events/meta_payloads.go`
- `services/backend-service/internal/eventconsumer/meta_handlers.go`
- `services/backend-service/internal/message/inbound.go`
- `services/backend-service/internal/message/entity.go`
- `services/backend-service/internal/message/repository.go`
- `services/backend-service/migrations/000027_message_reply_context.up.sql`
- `services/backend-service/migrations/000027_message_reply_context.down.sql`
- `services/backend-service/internal/message/service.go`
- `services/backend-service/internal/api/v1/routes.go`
- `services/backend-service/internal/eventconsumer/registry.go`
- `shared/events/canonical.go`
- `services/backend-service/migrations/000028_message_lifecycle.up.sql`
- `services/backend-service/migrations/000028_message_lifecycle.down.sql`

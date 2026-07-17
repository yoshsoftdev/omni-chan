# Ponytail Review — Channel Management Backend Foundation

## Scope

Backend-only preparation for multi-channel capabilities:

- Lightweight `ChannelManagementService` listing supported types and connection status.
- New read-only APIs under `/api/v1/channels/*`.
- WhatsApp connection status derived from existing `channels` table.
- Instagram reported as not connected (no OAuth, no Meta calls, no mock connected state).

No frontend changes. No Instagram OAuth/SDK. No webhook consumers. No Embedded Signup / WhatsApp connection/token/health/sync changes. No schema migration.

## Modified Files (8)

1. `services/backend-service/internal/channelmgmt/service.go` (new)
2. `services/backend-service/internal/channelmgmt/service_test.go` (new)
3. `services/backend-service/internal/api/v1/handler/channel_management.go` (new)
4. `services/backend-service/internal/api/v1/routes.go`
5. `services/backend-service/internal/api/v1/deps.go`
6. `services/backend-service/cmd/api/main.go`
7. `docs/openapi/v1.yaml`
8. `docs/reviews/ponytail-channel-management-backend.md` (this review)

## API Additions

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/api/v1/channels/types` | Returns WhatsApp + Instagram as supported types |
| GET | `/api/v1/channels/status` | `{ channels: [{ type, connected, display_name? }, ...] }` |
| GET | `/api/v1/channels/instagram/status` | `{ connected: false }` |

Routes are registered **before** `/channels/:id` so static paths are not captured as UUIDs.

Authorization mirrors channel list: `device.view` + team membership.

## WhatsApp Status Source

- Reuses existing `channel.Repository.List`.
- `connected: true` when any non-deleted channel has `status = connected`.
- `display_name` uses the first connected channel name when available.
- Does not call Meta and does not change credentials/health/sync.

## Instagram Status Source

- Capability only.
- Always `connected: false` until a future Instagram implementation exists.
- No Instagram rows invented in the database.

## Regression Check

- [x] Existing `/channels`, Embedded Signup, manual connect, health, sync, payment routes untouched in behavior.
- [x] `internal/channel/service.go`, `internal/meta/embedded_signup_service.go`, and channel connection handlers not modified.
- [x] No migration; no WhatsApp data migration.
- [x] No inbox/contacts/conversations/broadcast/worker/storage changes.
- [x] Unit tests cover types list, WhatsApp connected/disconnected, Instagram always false.

## Remaining Findings

1. **Multiple WhatsApp lines** — `/channels/status` currently exposes a single WhatsApp aggregate entry (first connected display name), not one row per channel ID. Existing `GET /channels` remains the inventory API.
2. **Future channel types** — constants exist for messenger/telegram/email in the service package, but `/channels/types` only returns WhatsApp + Instagram (per this task’s response contract).
3. **Provider column** — existing `channels` rows are WhatsApp-oriented; Instagram connection storage is deferred to a future task.

## Confirmation

The existing WhatsApp Embedded Signup flow was **not** modified. This task only added a read-only channel capability layer beside the current channel APIs.

# Review — Instagram OAuth Onboarding + Inbox List Restore

## Scope

1. Restore inbox conversation list UI (remove platform text row from cards).
2. Implement Instagram onboarding via Facebook Login for Business (redirect OAuth).
3. Persist Instagram channel + encrypted credentials; report real connection status.
4. Do **not** implement Instagram messaging/webhooks in this phase.
5. Do **not** modify WhatsApp Embedded Signup.

## Modified / Added Files

### Inbox UI
- `frontend/src/features/inbox/components/conversation-card.tsx` — restored original card layout

### Database
- `services/backend-service/migrations/000033_channel_type_instagram.up.sql`
- `services/backend-service/migrations/000033_channel_type_instagram.down.sql`
- `services/backend-service/internal/channel/entity.go`
- `services/backend-service/internal/channel/repository.go`

### Instagram OAuth backend
- `services/backend-service/pkg/meta/instagram_oauth.go`
- `services/backend-service/internal/meta/client.go` (adapter methods only)
- `services/backend-service/internal/instagram/oauth_service.go`
- `services/backend-service/internal/instagram/oauth_service_test.go`
- `services/backend-service/internal/api/v1/handler/instagram_oauth.go`
- `services/backend-service/internal/channelmgmt/service.go` (real IG status)
- `services/backend-service/internal/channelmgmt/service_test.go`
- `services/backend-service/internal/api/v1/routes.go`
- `services/backend-service/internal/api/v1/deps.go`
- `services/backend-service/cmd/api/main.go`
- `services/backend-service/internal/config/config.go`
- `services/backend-service/.env.example`

### Frontend onboarding
- `frontend/src/features/channels/services/instagram-oauth.service.ts`
- `frontend/src/features/devices/components/devices-page.tsx`
- `frontend/src/app/(dashboard)/devices/instagram/callback/page.tsx`

### Docs
- `docs/openapi/v1.yaml`
- `docs/reviews/instagram-oauth-onboarding.md` (this file)

## API Additions

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/channels/instagram/oauth/start` | Return Facebook OAuth `auth_url` + signed `state` |
| POST | `/api/v1/channels/instagram/oauth/complete` | Exchange code, resolve IG account, persist channel |
| DELETE | `/api/v1/channels/instagram/connection` | Soft-delete connected Instagram channel |
| GET | `/api/v1/channels/instagram/status` | Now reads DB (`connected` + optional `display_name`) |

## Regression Check

- [x] WhatsApp Embedded Signup service/handlers/routes/wizard unchanged.
- [x] Conversation list restored (avatar badge only; no text platform row).
- [x] Existing WhatsApp channels default `channel_type=whatsapp` via migration default.
- [x] No Instagram webhook/messaging consumers added.
- [x] Unit tests for Instagram OAuth complete + channelmgmt status.

## Ops Prerequisites

Set `META_INSTAGRAM_REDIRECT_URI` to the frontend callback (e.g. `http://localhost:3000/devices/instagram/callback`) and add the same URI in Meta App Dashboard → Facebook Login → Valid OAuth Redirect URIs. Enable Instagram Messaging related permissions (App Review for production).

## Confirmation

Existing WhatsApp Embedded Signup flow was **not** modified.

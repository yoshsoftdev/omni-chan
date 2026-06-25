# Embedded Signup Flow (Backend Foundation)

**Note:** Meta JS SDK integration is Phase 11.2. This phase provides backend session and completion APIs only.

## Architecture

```text
Frontend (future SDK)
    ↓
POST /channels/:id/embedded-signup/start
    ↓
EmbeddedSignupService (in-memory session)
    ↓
POST /channels/:id/embedded-signup/:session_id/complete
    ↓
Meta OAuth code exchange + channel update + encrypted credentials
```

## Service Methods

```go
StartEmbeddedSignup(ctx, teamID, channelID) (*StartEmbeddedSignupResult, error)
GetEmbeddedSignupStatus(ctx, teamID, sessionID) (*EmbeddedSignupStatusResult, error)
CompleteEmbeddedSignup(ctx, teamID, sessionID, input CompleteEmbeddedSignupInput) error
```

## Stored on Completion

| Field | Destination |
|-------|-------------|
| `waba_id` | `channels.waba_id` |
| `phone_number_id` | `channels.phone_number_id` |
| `access_token` | `channel_credentials.access_token_enc` (AES-256-GCM) |
| `business_id` | Embedded signup session (in-memory until schema migration) |

## API Routes

| Method | Path |
|--------|------|
| POST | `/api/v1/channels/:id/embedded-signup/start` |
| GET | `/api/v1/channels/:id/embedded-signup/:session_id` |
| POST | `/api/v1/channels/:id/embedded-signup/:session_id/complete` |

## Configuration

Requires `META_APP_ID`, `META_APP_SECRET`, `META_EMBEDDED_CONFIG_ID`, `META_TOKEN_ENCRYPTION_KEY`.

Public SDK config available at `GET /api/v1/meta/embedded-config` (existing handler).

## Limitations (Phase 11)

- Sessions are in-memory (lost on restart)
- No SMB sync / coexistence triggers
- No frontend SDK wiring

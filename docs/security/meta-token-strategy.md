# Meta Token Strategy

## Storage Rules

| Rule | Implementation |
|------|----------------|
| Never in frontend | Tokens only in `channel_credentials.access_token_enc` |
| Never in API responses | Channel DTOs exclude credentials |
| Never in logs | No token logging in Meta services |
| Encrypted at rest | AES-256-GCM via `pkg/cryptoutil.TokenCipher` |

## Configuration

```env
META_TOKEN_ENCRYPTION_KEY=<base64 32-byte key>
```

Development fallback is set in `config.Load()` when `APP_ENV=development` and key is unset. **Production must set an explicit key.**

## Token Types

| Type | Usage |
|------|-------|
| System user token | Long-lived WABA management (manual connect / embedded signup result) |
| Long-lived user token | OAuth exchange from embedded signup code |
| Short-lived token | Not stored; exchanged immediately on signup completion |

## Rotation Strategy

1. **Proactive:** Monitor `token_expires_at`; alert 7 days before expiry
2. **Reactive:** Health check `token: false` triggers re-auth workflow in UI
3. **Rotation:** Update via embedded signup complete or future `PATCH /channels/:id/credentials` (Phase 11.2)
4. **Key rotation:** Decrypt with old key, re-encrypt with new `META_TOKEN_ENCRYPTION_KEY` via maintenance job (future)

## Access Pattern

```text
Handler → Service → TokenCipher.Decrypt → MetaClient (in-memory only) → Graph API
```

Decrypted tokens exist only for the duration of a single request goroutine.

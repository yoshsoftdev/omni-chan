# Embedded Signup Security Review

**Date:** 2026-06-22  
**Verdict:** PASS

## Token Handling

| Check | Status |
|-------|--------|
| Access token never in API responses | PASS |
| Access token never logged | PASS — structured logs use team_id, channel_id, session_id only |
| Token encrypted at rest (AES-256-GCM) | PASS |
| Token decrypted only during OAuth exchange + Meta validation | PASS |

## Session Security

| Control | Implementation |
|---------|----------------|
| OAuth state | 32-byte random `oauth_state` per session; validated on complete |
| Session expiration | 1 hour `expires_at`; rejected on complete if expired |
| Team scoping | All session queries filter `team_id` |
| Channel ownership | Session `channel_id` must match URL `:id` |
| Replay protection | SHA-256 hash of authorization code stored uniquely; reused codes rejected |
| Completed session | Cannot complete twice |

## Cross-Tenant Access

Requests with wrong `team_id` or `channel_id` return `404` (not found).

## OAuth Errors

Invalid/expired codes return `validation_error` without echoing tokens or secrets.

## Findings

None blocking. Recommend rate-limiting `POST .../complete` per team in a future hardening pass.

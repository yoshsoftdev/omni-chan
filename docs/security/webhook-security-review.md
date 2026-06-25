# Webhook Security Review

**Date:** 2026-06-22  
**Service:** `webhook-service` (port 8081)  
**Verdict:** PASS

## Verification (GET)

| Control | Implementation |
|---------|----------------|
| `hub.mode=subscribe` | Required |
| `hub.verify_token` | Compared to `META_VERIFY_TOKEN` |
| Challenge echo | Returns `hub.challenge` on success |

## Signature Validation (POST)

| Control | Implementation |
|---------|----------------|
| Header | `X-Hub-Signature-256` |
| Algorithm | HMAC-SHA256 of raw body with `META_APP_SECRET` |
| Invalid signature | `401 Unauthorized`, no NATS publish |

## Secret Handling

| Rule | Status |
|------|--------|
| App secret not logged | PASS |
| Access tokens never in webhook path | PASS (ingestion only) |
| Message body not logged at INFO | PASS — logs use event_id, type, phone_number_id |

## Network

- Webhook-service is public-facing; backend-service is internal-only
- `INTERNAL_SERVICE_TOKEN` header on NATS messages for consumer validation

## Findings

- Signature validation skipped when `META_APP_SECRET` empty (development only) — production must set secret

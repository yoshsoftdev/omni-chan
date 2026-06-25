# Phase 11.2 — Embedded Signup SDK Integration Review

**Date:** 2026-06-22  
**Verdict:** APPROVE WITH FINDINGS

## Summary

Replaced in-memory embedded signup with durable PostgreSQL sessions, OAuth code exchange, channel provisioning (business_id, waba, phone, encrypted token, onboarding completed), frontend Meta SDK integration, and automatic workspace readiness refresh.

## Session Durability

**PASS**

- Migration `000017_embedded_signup` — `embedded_signup_sessions` table
- Sessions survive process restart
- Resumable via `GET .../embedded-signup/:session_id`

## OAuth Exchange

**PASS**

- Real Meta `ExchangeEmbeddedCode` (no mock)
- Invalid / expired / reused code → `validation_error`
- Tests: success, invalid, expired, replay

## Security

**PASS**

- No token leakage in API or logs
- `oauth_state` validation, session expiry, code hash replay protection
- See `docs/security/embedded-signup-security-review.md`

## Health Integration

**PASS**

- Completion sets connected channel + credentials + onboarding completed
- `GET /channels/:id/health` returns healthy when Meta APIs respond (requires template sync for full healthy per Phase 11.1 rules)

## Workspace Readiness

**PASS**

- `RefreshReadinessInternal` called on successful signup
- Frontend also calls `refreshWorkspaceReadiness()` after complete

## Schema Review

**PASS**

- `channels.business_id` added (Option A — documented)
- `embedded_signup_sessions` with indexes on oauth_state, oauth_code_hash

## Frontend

**PASS**

- `EmbeddedSignupWizard` uses Meta JS SDK + backend APIs
- Mock simulation removed from primary wizard flow

## Findings

1. `template_sync_runs` still in-memory (Phase 11.1) — bundle Postgres migration in future phase
2. Rate limiting on complete endpoint not implemented (recommended hardening)

## Build & Test

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

## Recommendation

Proceed to **Phase 11.3 — Webhook Event Processing**.

# Meta Domain Audit

**Date:** 2026-06-22  
**Scope:** Channel, template, team, credentials schema and service usage

## Domain Mapping

| Omni-Chan Domain | Meta Entity | Storage | Status |
|------------------|-------------|---------|--------|
| `channels` | WhatsApp Business Account | `channels.waba_id` | PASS |
| `channels` | Phone Number (display) | `channels.phone_number` | PASS |
| `channels` | Phone Number ID | `channels.phone_number_id` | PASS |
| `channel_credentials` | Access Token | `channel_credentials.access_token_enc` | PASS (encrypted) |
| `templates` | Message Template | `templates.meta_template_id` | PASS |
| `teams` | Business Manager | **Not persisted** | GAP — see schema gap analysis |

## Field Consistency

| Field | Table / Location | Used By | Consistent |
|-------|------------------|---------|------------|
| `channel_id` | `channels.id` | All channel-scoped APIs | PASS |
| `waba_id` | `channels.waba_id`, `templates.waba_id` | Sync, health, validation | PASS |
| `phone_number_id` | `channels.phone_number_id` | Health, send (future) | PASS |
| `business_id` | Embedded signup session (in-memory) only | Embedded signup | GAP |
| `access_token` | `channel_credentials.access_token_enc` | Meta client calls | PASS |
| `token_expires_at` | `channel_credentials.token_expires_at` | Validation | PASS |

## Service Validation

| Service | Team scoping | Meta fields checked |
|---------|--------------|---------------------|
| `channel.ValidateForMeta` | Via `FindByID(teamID)` | status, waba, phone, token |
| `template.SyncService` | Via channel lookup | Full Meta readiness |
| `meta.HealthService` | Via `FindByID(teamID)` | token, waba, phone, templates API |
| `meta.EmbeddedSignupService` | Session `teamID` | waba, phone, token on complete |

## Legacy `devices` Module

`internal/device` still exists with plaintext `access_token` on device rows (legacy). Active API v1 uses `channels` + `channel_credentials`. Device worker path remains for campaign worker compatibility — out of Phase 11 migration scope.

## Verdict

**PASS with findings** — Core channel/template mapping is consistent. `business_id` not persisted at team or channel level (documented gap, no migration in Phase 11).

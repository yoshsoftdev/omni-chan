# Meta Schema Gap Analysis

**Date:** 2026-06-22  
**Rule:** No migrations in Phase 11

## Sufficient Today

| Requirement | Schema | Notes |
|-------------|--------|-------|
| WABA ID | `channels.waba_id VARCHAR(64)` | Indexed |
| Phone Number ID | `channels.phone_number_id VARCHAR(64)` | Unique per team when active |
| Access token | `channel_credentials.access_token_enc BYTEA` | Encrypted at rest |
| Token expiry | `channel_credentials.token_expires_at TIMESTAMPTZ` | Optional |
| Template identity | `templates.meta_template_id` + unique `(channel_id, meta_template_id)` | Idempotent upsert |
| Onboarding state | `channel_onboarding.status` | `completed` required for campaigns |

## Gaps (Future Migration)

| Field | Needed For | Current State | Recommendation |
|-------|------------|---------------|----------------|
| `business_id` | Business Manager ↔ Team mapping | In-memory embedded signup session only | Add `channels.business_id` or `teams.meta_business_id` in Phase 11.2 |
| `embedded_signup_sessions` | Durable signup flow across restarts | In-memory map in `EmbeddedSignupService` | Redis or DB table in Phase 11.2 |
| `meta_webhook_subscriptions` | Webhook event routing audit | Not stored | Phase 11.3 |
| `template_sync_runs` | Sync audit trail | Not stored | Optional observability table |

## No Action Required for Phase 11

Schema supports template sync, health checks, encrypted tokens, and channel readiness without migration.

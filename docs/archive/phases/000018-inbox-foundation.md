# Migration 000018 — Event Consumer Support (Phase 11.4)

Applied in Phase 11.4. Documented here for Phase 11.5 inbox foundation reference.

## Changes

| Object | Change |
|--------|--------|
| `contact_source` enum | Add `meta_inbound` |
| `messages` | Add `raw_payload JSONB` |
| `channels` | Add `health_metadata JSONB` |
| Index | `idx_channels_phone_number_id_lookup` |

## Migration 000019 — Inbox Foundation (Phase 11.5)

| Object | Change |
|--------|--------|
| `conversation_status` enum | Add `pending`, `resolved` |
| `conversations` | SLA columns + indexes |
| `team_inbox_settings` | Assignment mode |
| `team_member_presence` | Online/offline tracking |

Run via standard `migrate.Up` on backend startup.

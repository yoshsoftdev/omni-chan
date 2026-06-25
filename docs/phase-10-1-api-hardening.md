# Phase 10.1 — API Hardening

## Status

| Check | Result |
|-------|--------|
| `go build ./...` | Pass |
| `go test ./...` | Pass |
| OpenAPI spec | `docs/openapi/v1.yaml` |

## Route normalization

| Change | Before | After |
|--------|--------|-------|
| Messages nested | `GET/POST /messages/:conversationId` | `GET/POST /conversations/:id/messages` |
| Workspace detail removed | `GET /workspace/detail` | Removed — use `GET /workspace` |
| Team workspace removed | `GET /team/workspace` | Removed |
| Team base path | `/team/*` | `/teams/*` (users, members, roles only) |

### Current team routes

```
GET    /api/v1/teams/members
POST   /api/v1/teams/members
GET    /api/v1/teams/users
POST   /api/v1/teams/users/invite
PATCH  /api/v1/teams/users/:id/role
POST   /api/v1/teams/users/:id/deactivate
DELETE /api/v1/teams/users/:id
GET    /api/v1/teams/roles
```

## Pagination contract

Canonical query parameters only:

```
?page=1&page_size=20
```

| Rule | Value |
|------|-------|
| `page` | `>= 1` |
| `page_size` | `1..100` (default `20`) |
| `limit` alias | Rejected with `validation_error` |

## Error model

Failures:

```json
{
  "success": false,
  "error": {
    "code": "not_found",
    "message": "Campaign not found"
  }
}
```

| Code | HTTP |
|------|------|
| `validation_error` | 400 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `not_found` | 404 |
| `conflict` | 409 |
| `quota_exceeded` | 402 |
| `feature_not_available` | 501 |
| `internal_error` | 500 |

Implementation: `internal/api/response/errors.go`

## Feature availability (501)

| Endpoint | Message |
|----------|---------|
| `POST /templates/sync` | Template synchronization is not available yet |
| `POST /campaigns/:id/send` | Campaign sending is not available yet |
| `POST /campaigns/:id/cancel` | Campaign cancellation is not available yet |

## New endpoints

### Workspace readiness

```
GET /api/v1/workspace/readiness
```

Evaluates five checks (mirrors frontend `workspace-readiness.service.ts`):

- `channel_connected`
- `template_synced`
- `contacts_available`
- `team_members_added`
- `campaign_created`

Persists snapshot to `workspace_readiness_snapshots`.

### Dashboard overview

```
GET /api/v1/dashboard
```

Returns contacts, channels, campaigns, unread conversations, readiness score.

### Onboarding

```
GET  /api/v1/onboarding
POST /api/v1/onboarding/complete-step
```

Wizard progress stored in `workspace_readiness_snapshots.checks` JSON (onboarding sub-object).

## Files added / changed

| Path | Purpose |
|------|---------|
| `internal/api/response/errors.go` | Standard error envelope |
| `internal/api/v1/handler/base.go` | Pagination validation, error mapping |
| `internal/api/v1/routes.go` | Normalized routes |
| `internal/workspace/readiness*.go` | Readiness computation |
| `internal/workspace/onboarding.go` | Onboarding wizard state |
| `internal/workspace/snapshot_repository.go` | `workspace_readiness_snapshots` |
| `internal/dashboard/service.go` | Dashboard KPI aggregates |
| `internal/api/v1/handler/dashboard.go` | Dashboard handler |
| `docs/openapi/v1.yaml` | OpenAPI 3.0 spec |

## Out of scope (unchanged)

- Meta / WhatsApp Cloud API
- NATS / workers
- Repository interface redesign
- Existing service business rules

## Review checklist

- [ ] Frontend updated to `/teams/*` and nested message routes
- [ ] Frontend error parser handles `error.code`
- [ ] `limit` query param removed from frontend API client
- [ ] Dashboard + readiness wired to replace mocks

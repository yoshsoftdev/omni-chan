# DTO Consistency Audit

**Compared:** OpenAPI schemas ↔ v1 handler DTOs ↔ service layer return types ↔ frontend TypeScript types  
**Date:** 2026-06-22

## Summary

| Area | Result |
|------|--------|
| Auth | PASS (minor naming) |
| Workspace / Readiness | **FAIL** |
| Team / Users | **FAIL** |
| Pagination envelope | PASS (where used) |
| Conversations / Messages | **FAIL** (method + missing routes) |
| Campaigns | PASS (handler ↔ service) |
| Tags | PASS (handler ↔ service) |

---

## Auth

| Field | OpenAPI | Handler | Service | Frontend | Result |
|-------|---------|---------|---------|----------|--------|
| Login response tokens | `access_token`, `refresh_token` | Same | Same | `AuthSessionResponse` uses snake_case tokens | PASS |
| Error envelope | `{ error: { code, message } }` | `response/errors.go` | N/A | `ApiResponse.message` only (no `error` field) | **FAIL** (frontend) |

---

## Workspace Readiness

**Backend** (`workspace.ReadinessResult`):

```json
{
  "score": 80,
  "status": "in_progress",
  "checks": [{ "key": "channel_connected", "completed": true }]
}
```

**Frontend** (`WorkspaceReadiness` in `workspace-readiness.ts`):

```json
{
  "score": 80,
  "completedCount": 4,
  "totalCount": 5,
  "status": "in_progress",
  "checks": [{ "id": "channel_connected", "label": "...", "description": "...", "completed": true }],
  "quickActions": [{ "id": "connect_whatsapp", "label": "...", "href": "..." }]
}
```

| Issue | Severity |
|-------|----------|
| `checks[].key` (API) vs `checks[].id` (frontend) | HIGH |
| Missing `completedCount`, `totalCount` at top level | MEDIUM |
| Missing `label`, `description` per check | MEDIUM |
| Missing `quickActions` array | MEDIUM |
| Frontend uses client-side mock aggregation (`workspace-readiness.service.ts`), not `GET /workspace/readiness` | HIGH |

---

## Workspace / Team

| Endpoint | Handler returns | Frontend expects | Result |
|----------|-----------------|------------------|--------|
| `GET /workspace` | `team.Team` entity | `Team` via `getTeam()` calling `GET /teams` | **FAIL** — wrong path |
| `PATCH /workspace` | Updated team | `updateTeam()` calls `PATCH /teams` | **FAIL** — wrong path |
| `GET /workspace/quotas` | Quota object | `getWorkspace()` calls `GET /team/workspace` | **FAIL** — path and shape differ |
| `GET /teams/users` | `{ users: TeamMember[] }` (no pagination) | Paginated or filtered list with `?status=` | **FAIL** — OpenAPI documents pagination params handler ignores |
| `PATCH /teams/users/:id/role` | `TeamMember` | `PUT /team/users/:id/role` | **FAIL** — path prefix + HTTP method |
| `POST /teams/users/:id/deactivate` | `TeamMember` | `PATCH /team/users/:id/deactivate` | **FAIL** — path prefix + HTTP method |
| `GET /teams/roles` | `{ roles: [] }` | `GET /team/roles` | **FAIL** — path prefix |
| N/A | — | `GET /team/roles/:id`, `GET /team/permissions` | **FAIL** — routes not registered |

---

## Pagination Envelope

Handlers using `dto.NewPageResponse` return:

```json
{
  "success": true,
  "message": "ok",
  "data": {
    "items": [...],
    "pagination": { "page", "page_size", "total", "total_pages" }
  }
}
```

OpenAPI `PageResponse` schema matches. **PASS** where `pageParams()` is used.

Exceptions (no pagination wrapper):

- `GET /teams/members` → `{ members: [...] }`
- `GET /teams/users` → `{ users: [...] }`
- `GET /teams/roles` → `{ roles: [...] }`

---

## Conversations

| Field / Behavior | Backend | Frontend (`conversation.service.ts`) | Result |
|------------------|---------|----------------------------------------|--------|
| List path | `GET /conversations?page=&page_size=` | `GET /conversations` (no pagination params) | MEDIUM |
| Assign | `POST /conversations/:id/assign` | `PATCH /conversations/:id/assign` | **FAIL** |
| Mark read | Not implemented | `POST /conversations/:id/read` | **FAIL** — phantom route |
| Unread count | Not implemented | `GET /conversations/unread-count` | **FAIL** — phantom route |
| List response | `PageResponse` with conversation items | Expects `{ conversations: Conversation[] }` | **FAIL** |

---

## Messages

| Field | Backend | Frontend | Result |
|-------|---------|----------|--------|
| Path | `GET/POST /conversations/:id/messages` | Same | **PASS** |
| List response | `PageResponse` | Expects raw array or wrapped messages (verify at integration) | MEDIUM — envelope mismatch likely |

---

## Dashboard

| Field | Backend `dashboard.Overview` | OpenAPI `DashboardOverview` | Frontend | Result |
|-------|---------------------------|----------------------------|----------|--------|
| `contacts` | int64 | int64 | Not wired to API | PASS (spec) |
| `unread_conversations` | int64 | int64 | Not wired | PASS (spec) |
| `readiness_score` | int | int | Mock only | PASS (spec) |

Frontend dashboard page uses `WorkspaceReadinessDashboard` with mock readiness service, not `GET /dashboard`.

---

## Onboarding

| Field | Backend `OnboardingSnapshot` | Frontend | Result |
|-------|------------------------------|----------|--------|
| `current_step`, `completed_steps`, `score` | Present | Rich wizard state from `onboarding-data.service.ts` (mock) | **FAIL** — frontend not consuming API |

---

## Campaigns / Segments / Tags / Contacts / Channels

Handler request structs map cleanly to service inputs. Entity JSON tags align with database columns. No material field-name mismatches between handler DTOs and service contracts.

OpenAPI entity schemas for these resources are largely absent (documented as generic `object`), so OpenAPI ↔ handler comparison is **incomplete** rather than failing on specific fields.

---

## Unused / Dead DTO Paths

- `TeamHandler.Get` / `Update` — handler DTOs exist but routes not registered.
- `team.Service.GetWorkspace` — service method exists; no v1 route (workspace handler covers team get/update).

---

## Verdict

**FAIL** — Service ↔ handler DTO alignment is generally sound. Failures concentrate at the **frontend ↔ API** boundary (paths, HTTP methods, readiness shape, conversation list envelope) and **OpenAPI ↔ handler** boundary (missing schemas and incorrect pagination documentation for team users).

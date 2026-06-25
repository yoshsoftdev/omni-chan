# Frontend Compatibility Report

**Validated:** Frontend services and types against v1 API (`/api/v1`)  
**Date:** 2026-06-22

---

## Summary

| Area | Status |
|------|--------|
| Auth (login/register/refresh) | Compatible |
| Messages (`/conversations/:id/messages`) | **Compatible** (path) |
| Team APIs | **Incompatible** (paths, methods) |
| Workspace / readiness | **Not integrated** (mock) |
| Dashboard | **Not integrated** (mock) |
| Onboarding | **Not integrated** (mock) |
| Conversations | **Partially incompatible** |
| Error handling | **Incompatible** |

**Overall: FAIL**

---

## Auth

| Frontend | Backend | Result |
|----------|---------|--------|
| `POST /auth/login` | Registered | PASS |
| `POST /auth/register` | Registered | PASS |
| `POST /auth/refresh` | Registered | PASS |
| `GET /auth/me` | Registered | PASS |
| Token storage snake_case ↔ camelCase mapping | Handler returns snake_case in data | PASS |

---

## Team (`features/team/services/team.service.ts`)

| Frontend call | Backend route | Result |
|---------------|---------------|--------|
| `GET /teams` | **Not registered** (use `GET /workspace`) | **FAIL** |
| `PATCH /teams` | **Not registered** (use `PATCH /workspace`) | **FAIL** |
| `GET /teams/members` | `GET /teams/members` | PASS |
| `POST /teams/members` | `POST /teams/members` | PASS |
| `PATCH /teams/members/:id` | **Not registered** | **FAIL** |
| `DELETE /teams/members/:id` | **Not registered** (use `DELETE /teams/users/:id`) | **FAIL** |
| `GET /team/users` | `GET /teams/users` | **FAIL** — `/team` vs `/teams` |
| `POST /team/users/invite` | `POST /teams/users/invite` | **FAIL** |
| `PUT /team/users/:id/role` | `PATCH /teams/users/:id/role` | **FAIL** — method + path |
| `PATCH /team/users/:id/deactivate` | `POST /teams/users/:id/deactivate` | **FAIL** — method + path |
| `DELETE /team/users/:id` | `DELETE /teams/users/:id` | **FAIL** |
| `GET /team/roles` | `GET /teams/roles` | **FAIL** |
| `GET /team/roles/:id` | **Not registered** | **FAIL** |
| `GET /team/permissions` | **Not registered** | **FAIL** |
| `GET /team/workspace` | **Not registered** (use `GET /workspace`, `GET /workspace/quotas`) | **FAIL** |

---

## Workspace Readiness

| Frontend | Backend | Result |
|----------|---------|--------|
| `workspace-readiness.service.ts` | Aggregates from mock channel/template/contact/team/campaign services | **Not calling API** |
| `WorkspaceReadiness` type | `ReadinessResult` with different field names | **FAIL** if wired |
| Dashboard widget | `WorkspaceReadinessDashboard` uses mock service | **FAIL** |

Expected API fields when integrated:

- Rename consumer: `checks[].key` → map to `id`
- Add server fields or client mapping for `completedCount`, `totalCount`, `quickActions`

---

## Dashboard

| Frontend | Backend | Result |
|----------|---------|--------|
| `app/(dashboard)/dashboard/page.tsx` | No service calls `GET /dashboard` | **FAIL** — not wired |
| Expected overview stats | `dashboard.Overview` JSON matches `DashboardOverview` OpenAPI schema | PASS (shape only) |

---

## Onboarding

| Frontend | Backend | Result |
|----------|---------|--------|
| `onboarding-data.service.ts` | Mock step state | **Not calling API** |
| `GET /onboarding` | Registered | N/A until integrated |
| `POST /onboarding/complete-step` | Registered | N/A until integrated |

---

## Conversations (`features/inbox/services/conversation.service.ts`)

| Frontend | Backend | Result |
|----------|---------|--------|
| `GET /conversations` | Registered (paginated `PageResponse`) | **FAIL** — expects `{ conversations: [] }` |
| `GET /conversations/:id` | Registered | PASS (path) |
| `PATCH /conversations/:id/assign` | `POST /conversations/:id/assign` | **FAIL** |
| `PATCH /conversations/:id/status` | `PATCH /conversations/:id/status` | PASS |
| `POST /conversations/:id/read` | **Not registered** | **FAIL** |
| `GET /conversations/unread-count` | **Not registered** | **FAIL** |

---

## Messages (`features/inbox/services/message.service.ts`)

| Frontend | Backend | Result |
|----------|---------|--------|
| `GET /conversations/:id/messages` | Registered | **PASS** |
| `POST /conversations/:id/messages` | Registered | **PASS** |

---

## API Client (`services/api/client.ts`)

| Behavior | v1 API | Result |
|----------|--------|--------|
| Success parsing `body.success`, `body.data` | Matches | PASS |
| Error parsing `body.message` | Errors in `body.error.message` | **FAIL** |
| Base URL `/api/v1` | Matches router | PASS |

---

## UI Routes vs API Routes

Frontend navigation uses `/team/*` page routes (Next.js). This is fine for UI routing but **must not** be confused with API paths — team API calls incorrectly use `/team/` prefix.

---

## Verdict

**FAIL** — Messages path alignment (Phase 10.1) is the main success. Team, conversation, workspace, dashboard, and onboarding integrations are not aligned with v1 API. Error client must be updated before meaningful E2E testing.

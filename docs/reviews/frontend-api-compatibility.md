# Frontend ↔ Backend API Compatibility Matrix — Phase 12.6A

**Date:** 2026-06-22  
**API base:** `/api/v1`

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Wired and used by active frontend |
| ⚠️ | Partial / workaround |
| ❌ | Not exposed by backend yet |
| 🗑️ | Legacy frontend path (removed from active UI) |

---

## Core Features

| Frontend service | HTTP | Backend route | Status | Notes |
|------------------|------|---------------|--------|-------|
| `dashboard.service` | GET | `/dashboard` | ✅ | Metrics + unread count |
| `workspace-readiness.service` | GET | `/workspace/readiness` | ✅ | |
| | POST | `/workspace/readiness/refresh` | ✅ | |
| `teamService.getTeam` | GET | `/workspace` | ✅ | |
| `teamService.listUsers` | GET | `/teams/users` | ✅ | |
| `teamService.listMembers` | GET | `/teams/members` | ✅ | |
| `teamService.listRoles` | GET | `/teams/roles` | ✅ | |
| `teamService.getPermissionMatrix` | — | — | ❌ | Returns `[]`; matrix hidden |
| `channelQueryService` | GET | `/channels` | ✅ | Paginated `items` |
| | POST | `/channels` | ✅ | Manual connect |
| | POST | `/channels/manual/verify` | ✅ | Pre-save Meta validation |
| | GET | `/channels/:id` | ✅ | |
| | DELETE | `/channels/:id` | ✅ | |
| `embeddedSignupService` | POST | `/channels/:id/embedded-signup/start` | ✅ | |
| `templateQueryService` | GET | `/templates` | ✅ | |
| `broadcast.service` sync | POST | `/templates/sync?channel_id=` | ✅ | |
| `campaignQueryService` | GET | `/campaigns` | ✅ | `campaigns` array |
| | POST | `/campaigns` | ✅ | Uses `channel_id` |
| | POST | `/campaigns/:id/send` | ✅ | |
| `contactService` | GET | `/contacts` | ✅ | Paginated `items` |
| `tagService` | GET | `/tags` | ✅ | Paginated `items` |
| `segmentService` | GET | `/segments` | ✅ | Paginated `items` |
| `conversationService` | GET | `/conversations` | ✅ | |
| | POST | `/conversations/:id/read` | ✅ | |
| `messageService` | GET | `/conversations/:id/messages` | ✅ | Paginated `items` |

---

## Removed Legacy Paths

| Frontend (removed) | Backend | Status |
|--------------------|---------|--------|
| `deviceService.list()` → `GET /devices` | Not registered | 🗑️ |
| `deviceService.create()` → `POST /devices` | Not registered | 🗑️ |
| `deviceService.*onboarding*` | No channel onboarding routes | 🗑️ → ⚠️ polls `GET /channels/:id` |

---

## Response Shape Notes

| Endpoint | Frontend expects | Backend provides | Match |
|----------|------------------|------------------|-------|
| `GET /contacts` | `data.items` or `data.contacts` | Paginated `items` | ✅ |
| `GET /tags` | `data.items` | Paginated `items` | ✅ |
| `GET /segments` | `data.items` | Paginated `items` | ✅ |
| `GET /templates` | snake_case JSON | Entity JSON tags added | ✅ |
| `GET /channels` | `items[]` with `name` | Maps to `Channel.label` | ✅ |
| `GET /campaigns` | `campaigns[]` | `campaigns` wrapper | ✅ |
| `GET /teams/users` | `users[]` | `users` wrapper | ✅ |

---

## Known Gaps (post-12.6A)

| Feature | Gap | Frontend behavior |
|---------|-----|-------------------|
| Coex onboarding sync progress | No `GET /channels/:id/onboarding` | `channelOnboardingService` derives status from channel `status` |
| Workspace activity feed | No API | Empty state on `/team/activity` |
| Segment analytics | No API | Empty state on segment detail |
| Permission matrix persistence | No API | Read-only empty matrix |
| Campaign schedules UI | No schedule API | Page shows toast only on edit |
| Segment update | No PATCH wired in detail page | Local-only edit toast |

---

## Error Handling

| Condition | Frontend message |
|-----------|------------------|
| HTML error page | `Unexpected server response` or route-specific message |
| `success: false` JSON | `message` / `error.message` from API |
| 401 | Refresh token flow → clear auth |

---

## Recommended Backend Follow-ups (Phase 13+)

1. `GET /channels/:id/onboarding` — expose `channel_onboarding` table
2. `GET /workspace/activity` — audit log for team page
3. `GET /teams/permissions` — permission matrix
4. `GET /segments/:id/analytics` — segment charts
5. Remove any remaining `DeviceDTO` naming in OpenAPI (use `Channel`)

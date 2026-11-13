# Frontend API Cutover Audit — Phase 12.6A

**Date:** 2026-06-22  
**Scope:** `frontend/src/` (active routes; `_legacy/` and `mock/` folders noted separately)

## Summary

All **active dashboard routes** now load data from API v1 via services/repositories. Mock repositories remain in the tree for reference but are **not wired** by factory defaults.

| Area | Status | API |
|------|--------|-----|
| Dashboard | ✅ API | `GET /dashboard` |
| Workspace readiness | ✅ API | `GET /workspace/readiness`, `POST /workspace/readiness/refresh` |
| Channels (Devices page) | ✅ API | `GET/POST/PATCH/DELETE /channels`, `POST /channels/manual/verify` |
| Templates | ✅ API | `GET /templates`, `POST /templates/sync` |
| Campaigns | ✅ API | `GET/POST/PATCH /campaigns`, `POST /campaigns/:id/send` |
| Contacts | ✅ API | `GET/POST/PATCH/DELETE /contacts` |
| Tags | ✅ API | `GET/POST/PATCH/DELETE /tags` |
| Segments | ✅ API | `GET /segments` |
| Inbox | ✅ API | `GET /conversations`, `GET /conversations/:id/messages` |
| Team users | ✅ API | `GET /teams/users`, `GET /teams/members` |
| Team workspace | ✅ API | `GET /workspace` |
| Team activity | ✅ Empty state | No backend endpoint yet |
| Permission matrix | ✅ Empty state | `GET` returns `[]` until backend exists |

---

## Active Route Audit (by feature)

### Channels (`/devices`)

| File | Previous | Replacement |
|------|----------|-------------|
| `features/devices/components/devices-page.tsx` | `deviceService`, mock list | `channelQueryService.getChannels()` |
| `features/channels/components/manual-connect-form.tsx` | — | `POST /channels/manual/verify`, `POST /channels` |
| `features/broadcast/components/templates-page.tsx` | `deviceService.list()` | `channelQueryService.getConnectedChannels()` |
| `app/(dashboard)/devices/onboarding/[id]/page.tsx` | `deviceService.list()` | `channelQueryService.getChannels()` |
| `features/devices/components/onboarding-wizard.tsx` | `deviceService` onboarding | `channelOnboardingService` → `GET /channels/:id` |

### Broadcast

| File | Previous | Replacement |
|------|----------|-------------|
| `repositories/template.repository.ts` | `template.repository.mock` | `template.repository.api` → `GET /templates` |
| `repositories/campaign.repository.ts` | `campaign.repository.mock` | `campaign.repository.api` → `GET /campaigns` |
| `components/campaigns/campaign-wizard.tsx` | `mockSampleContactValues` | User mapping only; templates/channels from query services |
| `components/create-template/*` | `deviceService` | `channelQueryService.getConnectedChannels()` |
| `services/broadcast.service.ts` | body `device_id` sync | `POST /templates/sync?channel_id=` |

### Contacts

| File | Previous | Replacement |
|------|----------|-------------|
| `components/contacts-list-page.tsx` | `MOCK_CONTACTS`, `MOCK_TAGS` | `contactService.list()`, `tagService.list()` |
| `components/tags-list-page.tsx` | mock tags | `tagService.list()` |
| `components/tag-detail-page.tsx` | `MOCK_*` | `tagService.getById()`, `contactService.list({ tag_id })` |
| `components/segment-detail-page.tsx` | `MOCK_SEGMENTS` | `segmentService.getById()` |
| `components/segments/segment-analytics.tsx` | mock charts | Empty state (no analytics API) |
| `components/import/import-wizard.tsx` | `SAMPLE_CSV_ROWS` | Parsed upload preview + `contactService.importCSV/Excel` |

### Inbox

| File | Previous | Replacement |
|------|----------|-------------|
| `components/inbox-page.tsx` | `MOCK_CONVERSATIONS`, `MOCK_MESSAGES` | `conversationService`, `messageService` |
| `components/conversation-detail.tsx` | `MOCK_AGENTS`, `MOCK_QUICK_REPLIES` | `teamService.listMembers()`, empty quick replies |
| `components/send-message-form.tsx` | `deviceService` | `channelQueryService.getConnectedChannels()` |

### Team

| File | Previous | Replacement |
|------|----------|-------------|
| `components/team-users-page.tsx` | `MOCK_TEAM_MEMBERS` | `teamService.listUsers()` |
| `components/team-workspace-page.tsx` | `MOCK_WORKSPACE`, activity | `teamService.getTeam()`, `listMembers()` |
| `components/team-activity-page.tsx` | `MOCK_WORKSPACE_ACTIVITY` | Empty state |

### Workspace / Dashboard

| File | Previous | Replacement |
|------|----------|-------------|
| `services/dashboard.service.ts` | — | `GET /dashboard` |
| `services/workspace-readiness.service.ts` | local score calc | `GET /workspace/readiness` |
| `services/onboarding-data.service.ts` | `buildRecentImports()` fake history | `recentImports: []` + real channel/template/segment queries |

---

## Remaining Mock / Legacy Files (not wired to active UI)

These files **exist** but are not imported by active pages:

| Path | Notes |
|------|-------|
| `features/broadcast/mock/*` | Mock data; factories point to API repos |
| `features/broadcast/_legacy/**` | Old broadcast UI; still references `deviceService` |
| `features/contacts/mock-data.ts` | Unused by active pages |
| `features/inbox/mock-data.ts` | Unused by active inbox |
| `features/team/mock-data.ts` | `INITIAL_ROLES` / permission UI constants only on team-users page |
| `features/devices/services/device.service.ts` | Legacy `/devices` client; **not used by devices-page** |
| `features/devices/components/device-card.tsx` | Legacy component |
| `features/channels/services/embedded-signup.mock.ts` | Orphan file |

---

## API Client

| File | Change |
|------|--------|
| `services/api/client.ts` | Non-JSON responses throw `Unexpected server response` instead of JSON parse errors |

---

## Validation Commands

```bash
cd frontend
pnpm lint          # pass (warnings only)
pnpm exec tsc --noEmit  # pass
pnpm build         # pass
```

## Grep Exit Criteria

| Check | Result |
|-------|--------|
| Active pages import `deviceService` | ❌ None on `/devices`, templates, campaigns, inbox |
| Active pages import `MOCK_*` data | ❌ None |
| Factory defaults use API repositories | ✅ template + campaign |
| Empty states when API returns 0 rows | ✅ Implemented per feature |

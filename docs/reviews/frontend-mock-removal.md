# Frontend Mock Removal Report — Phase 12.6A

**Date:** 2026-06-22

## Objective

Remove mock/fake/demo data from **active** frontend routes so UI reflects backend API v1 state.

## Removed / Replaced

### Data sources removed from active UI

| Mock source | Was used by | Replacement |
|-------------|-------------|-------------|
| `MOCK_CONTACTS`, `MOCK_TAGS`, `CONTACT_STATS` | contacts-list-page | `contactService`, computed stats |
| `MOCK_SEGMENTS` | segment-detail-page | `segmentService.getById()` |
| `MOCK_TAGS` | tag-detail-page | `tagService.getById()` |
| `SEGMENT_*_STATS` | segment-analytics | Empty state |
| `SAMPLE_CSV_ROWS` | import-wizard | Client CSV parse from uploaded file |
| `MOCK_CONVERSATIONS`, `MOCK_MESSAGES` | inbox-page | conversation + message services |
| `MOCK_AGENTS`, `MOCK_QUICK_REPLIES` | conversation-detail | `teamService.listMembers()`, `[]` |
| `MOCK_TEAM_MEMBERS` | team-users-page | `teamService.listUsers()` |
| `MOCK_WORKSPACE`, `MOCK_WORKSPACE_ACTIVITY` | team-workspace, activity | API workspace + empty activity |
| `mockSampleContactValues` | campaign-wizard | Mapping placeholder text |
| `defaultCreateTemplateForm` (demo body) | create-template-page | Empty form defaults |
| `buildRecentImports()` | onboarding-data.service | `recentImports: []` |
| `template.repository.mock` factory | template list | `template.repository.api` |
| `campaign.repository.mock` factory | campaign list/reports | `campaign.repository.api` |

### Device → Channel migration

| Removed | Replaced with |
|---------|---------------|
| `deviceService.list()` on templates page | `channelQueryService.getConnectedChannels()` |
| `deviceService` on create-template | `channelQueryService` |
| `deviceService` on send-message-form | `channelQueryService` |
| `deviceService` on onboarding page | `channelQueryService` + `channelOnboardingService` |
| `POST /templates/sync` body `device_id` | query `channel_id` |

### Constants extracted (no mock dependency)

| New file | Purpose |
|----------|---------|
| `features/contacts/constants.ts` | `CONTACT_SOURCES` |
| `features/contacts/constants/import.ts` | `IMPORT_FIELD_OPTIONS` |
| `features/broadcast/constants/create-template.ts` | Template form option lists |

### Chart types moved out of mock

| File | Types now in |
|------|--------------|
| `hourly-bar-chart.tsx` | `features/broadcast/types/campaign.ts` |
| `performance-line-chart.tsx` | `features/broadcast/types/campaign.ts` |
| `delivery-donut-chart.tsx` | `features/broadcast/types/campaign.ts` |

## Intentionally retained (not active UI data)

| Item | Reason |
|------|--------|
| `features/broadcast/mock/**` | Dead code; API repos are default |
| `features/broadcast/_legacy/**` | Quarantined; not routed |
| `features/team/mock-data.ts` | Role/permission **UI scaffolding** on team-users (matrix hidden when API empty) |
| `features/devices/services/device.service.ts` | Legacy file; delete in follow-up cleanup PR |
| `permission-matrix.tsx` mock toggle note | UI-only until `GET /teams/permissions` exists |

## Empty State Policy

Implemented messages:

- No channels found
- No templates found
- No campaigns found
- No contacts found
- No tags found
- No segments found
- No conversations found
- No workspace activity yet
- No permission matrix data available yet

## Build Verification

```
pnpm lint         ✅ (warnings only)
pnpm exec tsc --noEmit  ✅
pnpm build        ✅
```

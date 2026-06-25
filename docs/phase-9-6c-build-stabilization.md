# Phase 9.6c — Build Stabilization Report

## Build Status

| Command | Result |
|---------|--------|
| `go build ./...` | ✅ Pass |
| `go test ./...` | ✅ Pass |

## Compile Errors Fixed

| Package | Issue | Fix |
|---------|-------|-----|
| `internal/auth` | `user.ErrEmailTaken` undefined | Added alias in `user/errors.go` |
| `internal/contact` | Handler/service signature drift | Updated handler for `actorRole`, `ListFilter`, `PageParams`, pointer fields |
| `internal/template` | Legacy `ListFilter`, missing `Sync`/`Create` | Mapped filter fields; stubbed Meta-only endpoints |
| `internal/conversation` | `ListInput`, missing `actorRole`, `UnreadCount`/`MarkRead` | Updated to v1 service API; stubbed unmigrated endpoints |
| `internal/message` | `List`/`Send` vs `ListByConversation`/`Create` | Updated handler; added `HandleMetaInbound` stub |
| `internal/campaign` | Legacy worker/scheduler/handler drift | Handler aligned to v1 `Create`; worker/scheduler stubs; legacy code behind `//go:build legacy` |
| `internal/onboarding` | `FindOrCreateByPhone`, legacy `Message` fields | Switched to repositories + local helpers |
| `cmd/api/main.go` | Constructor mismatches across services | Full DI rewire (quota, rbac, channel, segment, uow) |

## Constructor Changes (`cmd/api/main.go`)

```go
quotaSvc := quota.NewService(teamRepo, channelRepo, contactRepo)
teamService := team.NewService(teamRepo, userRepo, quotaSvc, rbacService)
contactService := contact.NewService(contactRepo, teamRepo, quotaSvc, rbacService)
conversationService := conversation.NewService(conversationRepo, contactRepo, teamRepo, rbacService)
messageService := message.NewService(messageRepo, conversationRepo, teamRepo, rbacService)
templateService := template.NewService(templateRepo, channelRepo, teamRepo, rbacService)
campaignTx := uow.NewCampaignRunner(db)
campaignService := campaign.NewService(campaignRepo, campaignTx, teamRepo, channelRepo, templateRepo, segmentRepo, rbacService)
onboardingService := onboarding.NewService(deviceRepo, contactRepo, conversationRepo, messageRepo, teamRepo, metaClient)
```

## Test Fixes

| File | Change |
|------|--------|
| `internal/team/service_test.go` | Completed `mockUserRepo` with refresh-token methods |
| `internal/auth/service_test.go` | Same `mockUserRepo` completion |
| `internal/device/service_test.go` | Completed `mockTeamRepo`; added `config.MetaConfig` |
| `internal/tag/service_test.go` | Completed `mockTeamRepo` interface |
| `internal/campaign/create_tx_test.go` | Added `CountMembers` to segment mock |

## Dependency Graph Updates

```
cmd/api
  ├── rbac.Service
  ├── quota.Service ── team.Repository, channel.Repository, contact.Repository
  ├── team.Service ── quota.Service (MemberLimiter), rbac.Service
  ├── contact.Service ── quota.Service, rbac.Service
  ├── conversation.Service ── contact.Repository, rbac.Service
  ├── message.Service ── conversation.Repository, rbac.Service
  ├── template.Service ── channel.Repository, rbac.Service
  ├── campaign.Service ── uow.CampaignRunner, channel/segment/template repos, rbac.Service
  └── onboarding.Service ── contact/conversation/message repositories (not services)
```

## Import Cycles

No cycles introduced. `campaign` defines `UnitOfWorkRunner`; `pkg/uow` implements it via `CampaignRunner`.

## UoW Validation

- `CampaignService.Create` uses `UnitOfWorkRunner` ✅
- No compensating `SoftDelete` on create failure ✅
- `pkg/uow` and `internal/campaign` tests pass ✅

## Legacy Compatibility Notes

- **Campaign worker/scheduler**: Legacy implementations tagged `//go:build legacy`. Default build uses no-op stubs (`worker_compat.go`, `scheduler_compat.go`).
- **Endpoints returning 501**: Contact import, template sync/create, campaign send/overview/cancel/report, conversation mark-read — pending full service migration.
- **Device package**: Retained for legacy HTTP routes; v1 services use `channel` repository where applicable.

## Files Changed

### New
- `internal/campaign/worker_compat.go`
- `internal/campaign/scheduler_compat.go`
- `internal/message/legacy_inbound.go`
- `internal/onboarding/helpers.go`
- `docs/phase-9-6c-build-stabilization.md`

### Modified
- `cmd/api/main.go`
- `internal/user/errors.go`
- `internal/auth/service_test.go`
- `internal/contact/handler.go`
- `internal/template/handler.go`
- `internal/conversation/handler.go`
- `internal/message/handler.go`
- `internal/campaign/handler.go`
- `internal/campaign/worker.go` (+ legacy build tag)
- `internal/campaign/scheduler.go` (+ legacy build tag)
- `internal/campaign/render.go` (+ legacy build tag)
- `internal/campaign/variables.go` (+ legacy build tag)
- `internal/campaign/*_test.go` (legacy tags + mock fixes)
- `internal/onboarding/service.go`
- `internal/team/service_test.go`
- `internal/device/service_test.go`
- `internal/tag/service_test.go`

### Removed
- `internal/template/filter.go` (duplicate `ListFilter` — already removed in 9.6b)

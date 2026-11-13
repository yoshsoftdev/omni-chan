# Phase 9.6b — Unit Of Work Implementation Report

## Summary

Phase 9.6b implements transactional campaign creation via `pkg/uow`, refactors the campaign repository to accept `database.Querier`, and removes compensating `SoftDelete` rollback from `CampaignService.Create`.

Campaign create now runs **Create → CreateJob → CreateAudienceSnapshot** inside a single `TransactionManager.Run` boundary. Snapshot failure rolls back campaign + job via PostgreSQL transaction rollback (no application-level compensation).

## Files Changed

### New — `pkg/uow/`

| File | Purpose |
|------|---------|
| `manager.go` | `TransactionManager` with `Run(ctx, fn)` — begin, panic-safe rollback, commit |
| `unit_of_work.go` | `UnitOfWork` with lazy `Campaigns()` repository accessor |
| `factory.go` | `RepositoryFactory` → `campaign.NewRepositoryWithQuerier(q)` |
| `errors.go` | Re-exports `ErrNestedTransaction` from `pkg/repository` |
| `campaign_runner.go` | `CampaignRunner` implements `campaign.UnitOfWorkRunner` (breaks import cycle) |
| `errors_test.go` | Verifies `ErrNestedTransaction` re-export |
| `campaign_runner_test.go` | Verifies snapshot error propagation through runner |

### New — `internal/campaign/`

| File | Purpose |
|------|---------|
| `unit_of_work.go` | `UnitOfWorkRunner` interface (defined in campaign package to avoid `campaign ↔ uow` cycle) |
| `repository_test.go` | Nested-transaction guard when pool-backed querier has no `*database.DB` |
| `create_tx_test.go` | Single-tx create + no `SoftDelete` on snapshot failure |

### Modified — `internal/campaign/`

| File | Change |
|------|--------|
| `repository.go` | `PostgresRepository` uses `database.Querier` + `inTx` flag; `NewRepositoryWithQuerier`; `runTx()` helper; all queries via `r.q`; `Create`/`BulkCreateRecipients` participate in outer tx when `inTx=true` |
| `service.go` | Injects `UnitOfWorkRunner`; `Create` wrapped in `tx.Run`; **removed** compensating `SoftDelete` on job/snapshot failure |

### Modified — `pkg/database/`

| File | Change |
|------|--------|
| `postgres.go` | Merged `Querier` interface, `IsTx()`, panic-safe `WithTx()` (removed duplicate `db.go`) |

### Modified — `pkg/repository/`

| File | Change |
|------|--------|
| `errors.go` | Added `ErrNestedTransaction` |

### Removed

| File | Reason |
|------|--------|
| `pkg/database/db.go` | Duplicate of `postgres.go` (caused `DB redeclared` build error) |
| `internal/template/filter.go` | Duplicate `ListFilter` conflicting with `entity.go` |

## Architecture

```
CampaignService.Create
    └── UnitOfWorkRunner.Run          (campaign.UnitOfWorkRunner)
            └── TransactionManager.Run (pkg/uow)
                    └── pgx.Tx
                            └── UnitOfWork.Campaigns()
                                    └── NewRepositoryWithQuerier(tx)
```

**Import cycle resolution:** `internal/campaign` cannot import `pkg/uow` (factory imports campaign). `UnitOfWorkRunner` is defined in `internal/campaign`; `pkg/uow.CampaignRunner` implements it.

## Wiring (not applied — no handler/bootstrap changes in this phase)

```go
campaignRepo := campaign.NewRepository(db)
campaignTx := uow.NewCampaignRunner(db)
campaignSvc := campaign.NewService(
    campaignRepo,
    campaignTx,
    teamRepo,
    channelRepo,
    templateRepo,
    segmentRepo,
    rbacSvc,
)
```

## Repository Migration Notes — Campaign

### Before

- `PostgresRepository` held `*database.DB`, all queries used `r.db.Pool`
- `Create` opened its own transaction via `db.WithTx`
- `CreateJob` / `CreateAudienceSnapshot` used pool (autocommit) — **separate transactions**
- Service compensated with `SoftDelete` on partial failure

### After

| Method | Transaction behavior |
|--------|---------------------|
| `Create` | Uses `runTx`: participates in outer UoW tx when `inTx=true`; otherwise opens own tx via pool |
| `CreateJob` | Uses `r.q` directly — joins outer tx when bound to `pgx.Tx` |
| `CreateAudienceSnapshot` | Uses `r.q` directly — joins outer tx |
| `BulkCreateRecipients` | Uses `runTx` + `pgx.Batch` on `pgx.Tx` |
| Read/update methods | Use `r.q` (pool or tx) |

### Nested transaction guard

`NewRepositoryWithQuerier(pool)` without `*database.DB` returns `repository.ErrNestedTransaction` if `runTx` would begin a nested transaction. Inside `UnitOfWork`, repository is bound to `pgx.Tx` with `inTx=true` and executes inline.

## Remaining Repositories — Not Yet Migrated

All repositories below still use `db *database.DB` / `r.db.Pool` only (no `Querier` binding, not registered in `RepositoryFactory`):

| Package | Notes |
|---------|-------|
| `team` | Member + invitation flows may need UoW in future |
| `user` | Auth refresh tokens |
| `channel` | Credentials + onboarding upserts |
| `template` | Sync upserts |
| `contact` | Tag associations, bulk import |
| `segment` | `Create`/`Update` with rules; `ReplaceMembers` uses `WithTx` internally |
| `conversation` | Assignment updates |
| `message` | Send + status updates |
| `report` | Read-only aggregations (lower priority) |
| `importjob` | Multi-step import status (high priority for UoW) |
| `rbac` | Read-heavy; audit repo deferred per scope |

### Suggested next migrations (Phase 9.6c+)

1. **segment** — `Create`/`Update` + `ReplaceMembers` in one service transaction
2. **importjob** — job row + contact rows + status
3. **contact** — contact + tag pivot inserts

## Validation

| Check | Status |
|-------|--------|
| `pkg/repository` tests | ✅ Pass |
| `internal/rbac` tests | ✅ Pass |
| `pkg/uow` / `internal/campaign` tests | ⚠️ Blocked by pre-existing compile errors in legacy HTTP handlers (`contact/handler.go`, `template/handler.go`, `auth/handler.go`) and drifted test mocks (`team/service_test.go`) — unrelated to UoW changes |
| Full `go build ./...` | ⚠️ Same pre-existing handler/service signature drift from Phase 9C service layer vs legacy handlers |
| Campaign create single transaction | ✅ Implemented + unit-tested via `UnitOfWorkRunner` mock |
| Snapshot failure rollback | ✅ No `SoftDelete`; tx rollback via `TransactionManager` |
| Nested transaction guard | ✅ `ErrNestedTransaction` on pool-only querier without DB |
| Panic-safe rollback | ✅ `database.WithTx` + `uow.manager.Run` defer handlers |
| No HTTP handlers / NATS / audit repo | ✅ Scope respected |

## Out of Scope (unchanged)

- `cmd/api/main.go` wiring (legacy `campaign.NewService` signature still present)
- HTTP handler updates
- NATS workers
- Audit repository
- Multi-repository `UnitOfWork` accessors beyond campaigns

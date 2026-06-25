# Phase 9B — Repository Layer

**Status:** Ready for review  
**Scope:** PostgreSQL repositories only (no handlers, REST, NATS workers)

---

## 1. Objective

Implement the data-access layer for Omni-Chan backend v1, aligned with:

- Approved migrations (`services/backend-service/migrations/000001`–`000016`)
- API v1 contract (`docs/api-v1-contract.md`)
- Broadcast execution architecture (`docs/broadcast-execution-architecture.md`)

```
Handler → Service → Repository → PostgreSQL
```

All SQL is isolated in repositories. Services must not contain SQL.

---

## 2. Repository Inventory

| Spec name | Package | Interface | Implementation |
|-----------|---------|-----------|----------------|
| `TeamRepository` | `internal/team` | `Repository` | `PostgresRepository` |
| `UserRepository` | `internal/user` | `Repository` | `PostgresRepository` |
| `ChannelRepository` | `internal/channel` | `Repository` | `PostgresRepository` |
| `TemplateRepository` | `internal/template` | `Repository` | `PostgresRepository` |
| `ContactRepository` | `internal/contact` | `Repository` | `PostgresRepository` |
| `SegmentRepository` | `internal/segment` | `Repository` | `PostgresRepository` |
| `CampaignRepository` | `internal/campaign` | `Repository` | `PostgresRepository` |
| `ReportRepository` | `internal/report` | `Repository` | `PostgresRepository` |
| `ConversationRepository` | `internal/conversation` | `Repository` | `PostgresRepository` |
| `MessageRepository` | `internal/message` | `Repository` | `PostgresRepository` |
| `ImportRepository` | `internal/importjob` | `Repository` | `PostgresRepository` |

**Shared infrastructure:**

| Path | Purpose |
|------|---------|
| `pkg/database/db.go` | pgx pool, `WithTx`, `Querier` |
| `pkg/repository/pagination.go` | `PageParams`, `Page[T]`, `Meta` |
| `pkg/repository/errors.go` | `ErrNotFound`, `ErrConflict`, `MapError` |

**RBAC** (`internal/rbac`) retains its existing repository for permissions; not part of the Phase 9B deliverable list but used by team middleware.

---

## 3. Transaction Strategy

### 3.1 Default: single-statement autocommit

Most read and simple write operations use `db.Pool` directly (one round-trip, autocommit).

### 3.2 `database.WithTx`

Multi-step writes use `db.WithTx(ctx, fn)`:

| Repository | Operation | Why transactional |
|------------|-----------|-------------------|
| `campaign` | `Create` (campaign + audiences) | Atomic campaign + audience rows |
| `campaign` | `BulkCreateRecipients` | Batch insert consistency |
| `importjob` | `BulkCreateRows` | Batch row inserts |

Pattern:

```go
return r.db.WithTx(ctx, func(tx pgx.Tx) error {
    // use tx.Exec / tx.Query / pgx.Batch via tx.SendBatch
    return nil
})
```

### 3.3 Future: cross-repository transactions

Services that span repositories (e.g. campaign send + job creation) should accept an optional `pgx.Tx` or use a `UnitOfWork` wrapper in a later phase. Phase 9B keeps transactions inside repository boundaries.

### 3.4 Partitioned tables

`messages`, `campaign_recipients`, and `contact_import_rows` are partitioned by `created_at`. Inserts must always set `created_at` explicitly so rows land in the correct partition.

---

## 4. Error Strategy

### 4.1 Canonical errors (`pkg/repository/errors.go`)

| Error | When |
|-------|------|
| `ErrNotFound` | `pgx.ErrNoRows`, zero `RowsAffected` on expected update |
| `ErrConflict` | PostgreSQL `23505` (unique violation) |
| `ErrInvalidInput` | `23503` FK, `23514` check constraint |
| `ErrLimitExceeded` | Reserved for plan/quota enforcement in services |

### 4.2 Mapping

All repositories call `repository.MapError(err)` on pgx errors from `Exec`/`Query`.

Scan helpers map `pgx.ErrNoRows` → `repository.ErrNotFound`.

### 4.3 Team package legacy errors

`internal/team` retains package-local `ErrNotFound`, `ErrMemberNotFound`, etc. for backward compatibility with existing services. New packages use `pkg/repository` errors exclusively.

### 4.4 Service layer responsibility

Services translate repository errors to HTTP status:

- `ErrNotFound` → 404
- `ErrConflict` → 409
- `ErrInvalidInput` → 400
- Other → 500 (logged)

---

## 5. Pagination Strategy

### 5.1 Input: `repository.PageParams`

Matches API v1 query params:

- `page` (default 1)
- `page_size` (default 20, max 100)

### 5.2 Output: `repository.Page[T]`

```go
type Page[T any] struct {
    Items      []T
    Pagination Meta // page, page_size, total, total_pages
}
```

### 5.3 SQL pattern

1. `COUNT(*)` with same `WHERE` clause
2. `SELECT … ORDER BY … LIMIT $n OFFSET $m`

List endpoints use stable sort keys (`created_at DESC`, `last_message_at DESC NULLS LAST`, etc.).

### 5.4 Repositories with pagination

`user` (N/A lists), `channel`, `template`, `contact`, `segment`, `campaign`, `report` (funnels, link clicks), `conversation`, `message`, `importjob`.

---

## 6. Domain Models

Each package exposes typed entities in `entity.go`. Repositories return domain structs, not DTOs.

Key broadcast entities in `internal/campaign`:

- `Campaign`, `Audience`
- `BroadcastJob`, `BroadcastBatch`, `Recipient`, `AudienceSnapshot`

Aligned with `broadcast_jobs`, `broadcast_batches`, `campaign_recipients`, `campaign_audience_snapshots`.

---

## 7. Unit Test Strategy

### 7.1 Phase 9B (implemented)

| Package | Tests | Type |
|---------|-------|------|
| `pkg/repository` | `pagination_test.go`, `errors_test.go` | Pure unit |

### 7.2 Phase 9C (recommended)

| Layer | Approach |
|-------|----------|
| Repository integration | `testcontainers-go` PostgreSQL + `migrate-up` before suite |
| Scan / SQL builders | Table-driven tests with real DB |
| Cross-repo flows | Service tests with real DB or mocked interfaces |

### 7.3 Mocking for service tests

Generate mocks from repository interfaces (`//go:generate mockgen`) for handler/service unit tests without a database.

### 7.4 Coverage target

80% on repository packages once integration tests land (per backend rules).

---

## 8. File Layout

```
services/backend-service/
├── pkg/
│   ├── database/db.go
│   └── repository/
│       ├── pagination.go
│       ├── errors.go
│       └── *_test.go
└── internal/
    ├── team/       entity.go  repository.go
    ├── user/       entity.go  repository.go
    ├── channel/    entity.go  repository.go
    ├── template/   entity.go  repository.go
    ├── contact/    entity.go  repository.go
    ├── segment/    entity.go  repository.go
    ├── campaign/   entity.go  repository.go
    ├── report/     entity.go  repository.go
    ├── conversation/ entity.go repository.go
    ├── message/    entity.go  repository.go
    └── importjob/  entity.go  repository.go
```

---

## 9. Out of Scope (Phase 9B)

- HTTP handlers / Fiber routes
- REST API controllers
- NATS JetStream workers
- Redis rate limiting
- Service layer wiring
- DTO mapping

---

## 10. Review Checklist

- [ ] All 11 repository interfaces present
- [ ] SQL only in `repository.go` files
- [ ] `context.Context` on every method
- [ ] Domain models match migration columns
- [ ] Broadcast tables covered in `campaign` repository
- [ ] Partitioned tables set `created_at` on insert
- [ ] Pagination matches API v1 contract
- [ ] Transaction boundaries documented
- [ ] Error mapping consistent

---

## 11. Next Phase (9C suggestion)

1. Service layer for each feature module
2. Wire repositories in `cmd/api/main.go` DI
3. Integration tests with testcontainers
4. Handler layer per API v1 contract

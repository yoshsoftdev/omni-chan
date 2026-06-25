# Phase 10 — API v1 Handlers

## Status

| Check | Result |
|-------|--------|
| `go build ./...` | Pass |
| `go test ./...` | Pass |

HTTP-only API v1. No workers, NATS, Meta, or realtime added in this phase.

## Architecture

Handlers live under `internal/api/v1/handler/` and follow a thin pattern:

1. Parse and validate request shape
2. Resolve `auth.Actor` from context
3. Call the service layer (`actor.TeamID`, `actor.UserID`, `actor.Role`)
4. Map errors to the standard API envelope

RBAC enforcement remains in services. Handlers do not call `RequirePermission`.

### Middleware chain

| Route group | Middleware |
|-------------|------------|
| Public auth (`register`, `login`, `refresh`, `logout`) | None |
| Protected routes | `Auth(JWT)` → `ResolveActor(ActorResolver)` |

`ActorResolver` (`internal/auth/resolver.go`) loads team membership + RBAC permission snapshot into `auth.Actor`.

## Response format

```json
{ "success": true, "message": "...", "data": {} }
{ "success": false, "message": "..." }
```

Paginated lists return `PageResponse[T]`:

```json
{
  "items": [],
  "pagination": { "page": 1, "page_size": 20, "total": 0, "total_pages": 0 }
}
```

Query params: `page`, `page_size` (alias: `limit`).

## Routes

Base path: `/api/v1`

### Auth — `/auth`

| Method | Path | Handler |
|--------|------|---------|
| POST | `/register` | Register |
| POST | `/login` | Login |
| POST | `/refresh` | Refresh |
| POST | `/logout` | Logout |
| GET | `/me` | Me (protected) |

### Workspace — `/workspace`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | Get |
| GET | `/detail` | GetDetail |
| PATCH | `/` | Update |
| GET | `/quotas` | GetQuotas |

### Team — `/team`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | Get |
| PATCH | `/` | Update |
| GET | `/members` | ListMembers |
| POST | `/members` | AddMember |
| GET | `/users` | ListUsers |
| POST | `/users/invite` | InviteUser |
| PATCH | `/users/:id/role` | UpdateUserRole |
| POST | `/users/:id/deactivate` | DeactivateUser |
| DELETE | `/users/:id` | RemoveUser |
| GET | `/workspace` | GetWorkspace |

### Channels — `/channels`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| PATCH | `/:id` | Update |
| DELETE | `/:id` | Delete |

### Templates — `/templates`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| GET | `/:id` | Get |
| POST | `/sync` | Sync (501 — Meta not wired) |

### Contacts — `/contacts`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| PATCH | `/:id` | Update |
| DELETE | `/:id` | Delete |

### Segments — `/segments`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| PATCH | `/:id` | Update |
| DELETE | `/:id` | Delete |

### Tags — `/tags`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| PATCH | `/:id` | Update |
| DELETE | `/:id` | Delete |
| POST | `/:id/contacts/:contact_id` | Assign |
| DELETE | `/:id/contacts/:contact_id` | Unassign |

### Campaigns — `/campaigns`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| PATCH | `/:id` | Update |
| DELETE | `/:id` | Delete |
| POST | `/:id/send` | Send (501 — worker pending) |
| POST | `/:id/cancel` | Cancel (501 — pending) |

### Reports — `/reports`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/dashboard` | Dashboard (`from`, `to`, optional `channel_id`) |
| GET | `/daily` | DailyTrend |
| GET | `/campaigns` | ListCampaigns |
| GET | `/campaigns/:id` | GetCampaign |
| GET | `/campaigns/:id/clicks` | ListLinkClicks |

### Conversations — `/conversations`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/` | List |
| POST | `/` | Create |
| GET | `/:id` | Get |
| POST | `/:id/assign` | Assign |
| PATCH | `/:id/status` | UpdateStatus |

### Messages — `/messages`

| Method | Path | Handler |
|--------|------|---------|
| GET | `/:conversationId` | List |
| POST | `/:conversationId` | Send |

## Files added / changed

| Path | Purpose |
|------|---------|
| `internal/api/v1/handler/*.go` | v1 HTTP handlers |
| `internal/api/v1/dto/pagination.go` | `PageResponse[T]` |
| `internal/api/v1/deps.go` | Handler dependency struct |
| `internal/api/v1/routes.go` | Route registration |
| `internal/api/router.go` | Wires `v1.Deps` |
| `cmd/api/main.go` | Service + handler wiring |
| `internal/auth/actor*.go` | Actor model + resolver |
| `internal/api/middleware/actor.go` | `ResolveActor` middleware |
| `internal/tag/{entity,repository,service}.go` | Restored tag module |
| `internal/rbac/permissions.go` | Tag permission constants |

Legacy feature handlers (`internal/*/handler.go`, `routes.go`) tagged `//go:build legacy` where they conflicted with the new v1 layer.

## Deferred (out of scope)

- Campaign send/cancel worker integration
- Template Meta sync
- NATS / webhook consumers
- WebSocket realtime

## Review checklist

- [ ] Route naming matches frontend expectations
- [ ] Pagination query param conventions (`page_size` vs `limit`)
- [ ] Message path shape (`/messages/:conversationId` vs nested under conversations)
- [ ] 501 stubs acceptable for send/sync until worker/Meta phases

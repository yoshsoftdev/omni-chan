# Sprint 6 — Team Management & RBAC

## Overview

Workspace-based role and permission system. All authorization uses permission codes — not hardcoded role checks in handlers.

## Database

Migration `000012_rbac`:

- `roles` — system roles (owner, admin, supervisor, agent)
- `permissions` — granular permission codes
- `role_permissions` — role ↔ permission mapping
- `team_members.role_id` — FK to roles
- `team_members.status` — active / inactive
- `team_invitations` — mock invite flow

## API (`/api/v1/team/*`)

| Method | Endpoint | Permission |
|--------|----------|------------|
| GET | `/team/users` | `user.view` |
| POST | `/team/users/invite` | `user.invite` |
| PUT | `/team/users/:id/role` | `user.update` |
| PATCH | `/team/users/:id/deactivate` | `user.update` |
| DELETE | `/team/users/:id` | `user.delete` |
| GET | `/team/roles` | `user.view` |
| GET | `/team/roles/:id` | `user.view` |
| GET | `/team/permissions` | `team.manage` |
| GET | `/team/workspace` | `user.view` |

Legacy `/api/v1/teams/*` routes remain unchanged.

## Middleware

`rbac.RequirePermission(service, "permission.code")` returns HTTP 403:

```json
{ "success": false, "message": "permission denied" }
```

## Frontend Routes

- `/team/users` — user table, search, invite modal
- `/team/roles` — role list + permission matrix
- `/team/permissions` — editable matrix (mock)
- `/team/workspace` — workspace details

## Permission Constants

See `internal/rbac/permissions.go` for all codes used across modules.

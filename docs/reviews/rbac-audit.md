# RBAC Audit

**Rule:** Handlers must not enforce permissions. Services must enforce via `rbac.Require(...)` or equivalent.

**Date:** 2026-06-22

---

## Handler Layer

Searched `internal/api/v1/handler/` for `rbac.Require`, `RequirePermission`, and permission checks.

| Finding | Result |
|---------|--------|
| No handler calls `rbac.Require` for authorization | **PASS** |
| `TeamHandler.ListRoles` calls `h.rbac.ListRoles(c.Context())` | **PASS** — read-only catalog, not permission enforcement |
| All handlers delegate to services with `actor.Role` | **PASS** |

---

## Service Layer RBAC Pattern

Standard pattern across feature services:

```go
func (s *Service) authorize(role, perm string) error {
    if s.rbac != nil && s.rbac.Require(role, perm) {
        return nil
    }
    return ErrForbidden // or serviceerrors.ErrForbidden
}
```

| Service | `authorize()` helper | Permissions enforced | Result |
|---------|---------------------|---------------------|--------|
| contact | Yes | Create, View, Update, Delete | PASS |
| channel | Yes | DeviceManage, DeviceView | PASS |
| template | Yes | TemplateView, TemplateSync | PASS |
| segment | Yes | ContactCreate/View/Update/Delete | PASS |
| tag | Yes | TagCreate/View/Update/Delete | PASS |
| campaign | Yes | CampaignCreate/View | PASS |
| conversation | Yes | ConversationView, Assign, Reply | PASS |
| message | Yes | ConversationView, Reply | PASS |
| report | Yes | ReportView | PASS |
| workspace | Yes | UserView, WorkspaceManage | PASS |
| team | Yes | UserView/Invite/Update/Delete, WorkspaceManage | PASS |
| dashboard | Inline `rbac.Require` | UserView | PASS |

---

## Gaps and Inconsistencies

### MEDIUM — `team.Service.ListMembers`

Only checks team membership (`FindMemberByUser`), not `PermUserView`. Any member can list all members regardless of role permissions.

### MEDIUM — `team.Service.Get`

Membership-only check; no `authorize(actorRole, ...)`. Contrasts with `workspace.Service.Get` which requires `PermUserView`.

### LOW — `dashboard.Service`

Uses inline `!s.rbac.Require(...)` instead of shared `authorize()` helper. Functionally equivalent.

### LOW — Duplicated membership checks

Many services call both `authorize()` and `ensureMember()` / `FindMemberByUser`. Redundant but not a security defect.

---

## Handler-Level RBAC

**None found** in v1 handlers. **PASS**

---

## Legacy Code

`internal/rbac/middleware.go` may contain HTTP middleware for permission checks. Not wired into v1 route registration (`routes.go` uses only `Auth` + `ResolveActor`). **PASS** for active API.

---

## Permission Matrix vs API

Frontend calls `GET /team/permissions` for a permission matrix. **No v1 route registered.** RBAC data is only exposed via `GET /teams/roles` (role list without full matrix). Frontend feature is disconnected from API.

---

## Verdict

**PASS** — Architecture rule (RBAC in services, not handlers) is followed for all v1 endpoints. Minor gaps: `ListMembers` and `team.Get` lack explicit permission checks (membership-only). No handler-level RBAC violations.

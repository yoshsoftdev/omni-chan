# RBAC audit (TEAM-04, roadmap 9.1)

Date: 2026-11-16. Replaces [the June audit](../archive/reviews/rbac-audit.md), which predates automation, channels v2, quick replies and import.

## Rule

Permissions are enforced in **services** (`authorize(role, rbac.Perm…)` or `rbac.Require`), never in handlers. The UI hides what a role may not do, using the same codes from `GET /auth/me`. The UI is a convenience; the API is the gate.

## Method

1. A Go AST pass over `services/backend-service/internal` listed every exported method that takes an actor role but never calls a permission check. Each hit was read by hand.
2. Every handler in `internal/api/v1/handler` that never passes `actor.Role` to a service was read by hand. These are endpoints that cannot be permission-checked.
3. Only `internal/api/v1/routes.go` is mounted. The feature-local `routes.go` files (campaign, conversation, device, …) are not registered and are out of scope.

## Findings and fixes

| Area | Finding | Fix |
|---|---|---|
| Channels: Embedded Signup start/complete/status, re-authorize | Any member, including an agent, could connect a WhatsApp number or re-authorize a channel | `device.manage` checked in `meta.EmbeddedSignupService` |
| Channels: payment link | Any member could create a Meta payment link | `device.manage` in `meta.PaymentService.GetPaymentLink` |
| Channels: payment status, health | No check | `device.view` in `PaymentService.GetPaymentStatus`, `HealthService.CheckChannel` |
| Meta services without an authorizer | n/a | Fail closed: `ErrForbidden` (tested) |
| UI | Hard-coded role → permission map, out of sync with the database (agents really have `quickreply.view`, `conversation.close`, `automation.view`); custom roles got nothing | `/auth/me` returns the role's permission codes; `usePermission` uses them. The old map is only a fallback until that loads |
| UI | Management actions visible to every role | Hidden per permission in contacts, tags, segments, broadcast, templates, automation, quick replies, inbox (assign, close, reply, note) and the sidebar |

Methods the pass flagged that are fine: `SetEndpointStatus`, `SetForwardWebhookStatus`, `ListResponses`, `GetResponse` and `SendReply` delegate to checked methods. The inbox list is checked inside `conversation.ListCursor`. Presence goes through `conversation.Get`.

## Deliberate policies (not gaps)

- **Team member list (`GET /teams/members`) and workspace info** need only team membership. Agents do not hold `user.view` but need member names for assignment and display.
- **Assigning tags to a contact** needs only membership (`tag.Service.Assign`): it is a collaborative inbox action. Creating, editing and deleting tags stays permission-checked.
- **Public endpoints**: login, register, refresh, invitation accept, Meta webhook verify, OAuth callbacks and the signed media proxy are authenticated by their own token or signature, not by role.

## Known limits

- The RBAC cache is keyed by role **name**, so two teams' custom roles with the same name share permissions. Fix before custom roles are sold per team.
- Segment endpoints are checked with `contact.*` codes. The `segment.*` codes in the database are unused.

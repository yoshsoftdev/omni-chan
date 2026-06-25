Phase 1

[x] Backend Foundation
[x] Frontend Foundation
[x] Auth
[x] Team
[x] Device Management
[x] WebSocket Hub

Phase 2 — Team Inbox

[x] Contact
[x] Conversation
[x] Message
[x] Assignment
[x] Realtime WebSocket Inbox UI
[x] Tags
[x] CSV Import
[x] Excel Import

Phase 2 — Broadcast (Meta template campaigns)

[x] Template sync from Meta (no in-app template creation)
[x] Campaign create with variable mapping
[x] Async queue worker (Redis) + Meta API send
[x] Campaign scheduling
[x] Campaign reporting (sent, delivered, read, failed)
[x] Broadcast UI (/broadcast)

Deferred

[ ] Automation
[ ] Webhook delivery/read status sync for campaigns

Phase 3

[ ] Keyword Reply
[ ] Welcome Message
[ ] Analytics
[ ] Reports

Phase 4 — Onboarding & Subscription

[ ] Deferred — rolled back, billing not in scope

Phase 5 — Team Management & RBAC (awaiting review)

[x] Database migration (roles, permissions, role_permissions)
[x] RBAC middleware and permission cache
[x] Team management APIs (/api/v1/team/*)
[x] Frontend team pages (users, roles, permissions, workspace)
[x] Reusable components (UserTable, RoleBadge, PermissionMatrix, InviteUserModal, RoleSelector)
[ ] Apply permission middleware to all existing modules (inbox, contacts, broadcast)

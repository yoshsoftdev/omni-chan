# Multi-Tenant Safety Audit

**Rule:** All data access must be scoped by `actor.TeamID` from the resolved JWT actor.

**Date:** 2026-06-22

---

## Actor Resolution

Flow: `middleware.Auth(JWT)` → `middleware.ResolveActor` → `auth.Resolver.Resolve` loads team membership → `auth.Actor{ TeamID, UserID, Role }`.

Handlers retrieve actor via `requireActor(c)` and pass `actor.TeamID` to services. **PASS** — no v1 handler uses client-supplied `team_id` from body or query for authorization.

---

## Handler Audit (v1)

All handlers in `internal/api/v1/handler/` pass `actor.TeamID` as the first team-scoping argument to service methods. Spot-checked modules:

| Module | TeamID passed | Result |
|--------|---------------|--------|
| workspace | Yes | PASS |
| team | Yes | PASS |
| channel | Yes | PASS |
| contact | Yes | PASS |
| template | Yes | PASS |
| segment | Yes | PASS |
| tag | Yes | PASS |
| campaign | Yes | PASS |
| conversation | Yes | PASS |
| message | Yes | PASS |
| report | Yes | PASS |
| dashboard | Yes | PASS |

No handler performs direct repository access. **PASS**

---

## Service Layer Team Scoping

Repository methods consistently accept `teamID uuid.UUID` and include `team_id = $1` (or equivalent join) in SQL queries for:

- contacts, channels, templates, segments, tags, campaigns, conversations, messages

`ensureMember(ctx, teamID, userID)` or `FindMemberByUser` is called before data access in most services. **PASS**

---

## CRITICAL Finding

### `report.Service.ListLinkClicks`

**Location:** `internal/report/service.go:114-125`

```go
func (s *Service) ListLinkClicks(ctx context.Context, teamID, userID uuid.UUID, actorRole string, campaignID uuid.UUID, page repository.PageParams) (...) {
    // RBAC + membership checked
    result, err := s.reports.ListLinkClicks(ctx, campaignID, page)  // NO teamID passed
}
```

**Repository:** `internal/report/repository.go:175-186`

```sql
SELECT ... FROM campaign_link_clicks WHERE campaign_id = $1
```

**Risk:** If an attacker obtains a valid `campaign_id` belonging to another team, they can list that campaign's link clicks. Membership is verified for the **actor's** team, but **campaign ownership is not verified**.

**Severity:** **CRITICAL**

**Contrast:** `GetCampaignReport` / `GetCampaignFunnel` pass `teamID` and join/filter on `campaigns.team_id`.

---

## HIGH Finding

### `team.Service.Get` — No RBAC, membership-only gate

Any team member can read team metadata. Low cross-team risk (still team-scoped by ID) but worth noting. Not CRITICAL.

---

## Direct ID Lookups

| Operation | Team filter on lookup | Result |
|-----------|----------------------|--------|
| `contact.Get(ctx, teamID, id)` | Yes | PASS |
| `campaign.Get(ctx, teamID, id)` | Yes | PASS |
| `conversation.Get(ctx, teamID, id)` | Yes | PASS |
| `channel.Get(ctx, teamID, id)` | Yes | PASS |
| `tag.Assign(ctx, teamID, tagID, contactID)` | Both IDs team-scoped in service | PASS |
| `ListLinkClicks(ctx, campaignID)` | **No team filter** | **FAIL** |

---

## JWT Team Selection

Actor resolver binds to a single team from JWT claims / membership lookup. Users cannot switch teams via request parameters without re-authenticating with a different team context. **PASS**

---

## Verdict

**FAIL** — One **CRITICAL** cross-team data exposure risk in `GET /reports/campaigns/:id/clicks`. All other v1 handlers and services correctly propagate `actor.TeamID`.

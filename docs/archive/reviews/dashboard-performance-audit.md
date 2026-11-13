# Dashboard Performance Audit

**Reviewed:** `dashboard` service, `workspace` readiness computation, `report` aggregations  
**Scale targets:** 50k contacts, 300k contacts, millions of campaign recipients  
**Date:** 2026-06-22

---

## `GET /dashboard` Request Cost

**Service:** `internal/dashboard/service.go` → `GetOverview`

Per request:

| Step | Query / Operation | Index expectation |
|------|-------------------|-------------------|
| 1 | `COUNT(*) FROM contacts WHERE team_id = $1 AND deleted_at IS NULL` | `team_id` index |
| 2 | `COUNT(*) FROM channels WHERE team_id = $1 AND deleted_at IS NULL` | `team_id` index |
| 3 | `COUNT(*) FROM campaigns WHERE team_id = $1 AND deleted_at IS NULL` | `team_id` index |
| 4 | `COUNT(*) FROM conversations WHERE team_id = $1 AND unread_count > 0` | `team_id` index |
| 5 | `workspace.GetReadiness` → see below | Multiple queries + **write** |

**Total:** ~9–12 database round-trips per dashboard load, including one upsert.

---

## Readiness Computation (`computeReadiness`)

**File:** `internal/workspace/readiness.go`

| Check | Operation | Cost at scale |
|-------|-----------|---------------|
| Channel connected | `List(..., PageSize: 1)` with status filter | Low — LIMIT 1 |
| Template synced | `List(..., PageSize: 1)` | Low |
| Contacts available | `contacts.CountByTeam(teamID)` — **full COUNT(\*)** | **Scales with row count** |
| Team members | `teams.ListMembers(teamID)` — loads **all members** | Low (typically < 100) |
| Campaign created | `List(..., PageSize: 1)` | Low |
| Snapshot | `Upsert` write | Constant |

---

## Report Aggregations

**`GET /reports/dashboard`:** `AggregateDashboard` uses `COUNT(DISTINCT c.id)` with date range and optional channel filter. Acceptable with indexes on `team_id` + `created_at`.

**`GET /reports/campaigns`:** Paginated list with COUNT subquery per page request. Standard pattern.

**`GET /reports/campaigns/:id/clicks`:** `COUNT(*)` on `campaign_link_clicks WHERE campaign_id = $1`. At millions of recipients/clicks, this table grows large; pagination mitigates list payload but COUNT for total still scans index on `campaign_id`.

---

## Scale Analysis

### 50k contacts

| Endpoint | Assessment |
|----------|------------|
| `GET /dashboard` | **Acceptable** — 4 indexed COUNTs + readiness. Sub-second on modest hardware if `contacts.team_id` indexed. |
| `GET /workspace/readiness` | **Acceptable** — single COUNT on 50k rows |
| Report dashboards | **Acceptable** with date-bounded queries |

### 300k contacts

| Endpoint | Assessment |
|----------|------------|
| `GET /dashboard` | **MEDIUM risk** — COUNT on 300k rows on every page view; compounded by readiness recomputation |
| `GET /workspace/readiness` | **MEDIUM risk** — same COUNT per request |
| Repeated polling | Frontend mock services don't poll API yet; when wired, risk increases |

### Millions of recipients / link clicks

| Endpoint | Assessment |
|----------|------------|
| Campaign report funnel | Joins on campaign_recipients — needs composite indexes |
| Link click COUNT/LIST | **HIGH risk** without partitioning or pre-aggregated stats |
| Dashboard contact COUNT | Unaffected by recipient count (contacts table only) |

---

## Anti-Patterns Identified

1. **Readiness on every dashboard GET** — expensive COUNT + 4 list probes + snapshot write.
2. **No caching layer** — Redis not used for dashboard/readiness aggregates.
3. **Full recompute vs cached snapshot** — snapshot table exists but GET always recomputes before upserting.
4. **`ListMembers` loads all rows** for readiness check instead of `COUNT(*) > 1`.

---

## Index Dependencies (Assumed)

Performance relies on:

- `contacts(team_id)` partial index where `deleted_at IS NULL`
- `campaigns(team_id)`
- `conversations(team_id, unread_count)` or filter on `unread_count > 0`
- `campaign_link_clicks(campaign_id)`

Migration review not in scope; indexes should be verified before production scale.

---

## Verdict

| Scale | Dashboard | Readiness | Reports |
|-------|-----------|-----------|---------|
| 50k contacts | PASS | PASS | PASS |
| 300k contacts | MEDIUM | MEDIUM | PASS |
| Millions of recipients | PASS (contacts) | PASS (contacts) | **HIGH** (click aggregates) |

**Overall: MEDIUM risk** — acceptable for early rollout at tens of thousands of contacts; requires caching, read-from-snapshot, and pre-aggregation before enterprise scale.

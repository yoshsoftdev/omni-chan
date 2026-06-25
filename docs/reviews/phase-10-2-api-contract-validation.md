# Reviewed by Codex

## Summary

**APPROVE WITH FINDINGS**

Phase 10.2 contract validation is complete. The v1 API architecture (thin handlers → services → repositories, actor model, service-layer RBAC, standardized error envelope) is sound and ready for iterative hardening. However, **cross-team data exposure**, **GET side effects**, **frontend/API misalignment**, and **OpenAPI incompleteness** must be resolved before Meta integration.

No code was modified in this phase per scope constraints.

---

## CRITICAL

1. **`GET /reports/campaigns/:id/clicks` cross-team access** — `report.Service.ListLinkClicks` queries by `campaign_id` only without verifying the campaign belongs to `actor.TeamID`. An authenticated user who knows another team's campaign UUID could read link click data. See `docs/reviews/multi-tenant-audit.md`.

---

## HIGH

1. **GET side effects on readiness, onboarding, and dashboard** — `GET /workspace/readiness`, `GET /onboarding`, and `GET /dashboard` all trigger `workspace_readiness_snapshots` upsert via `computeReadiness()`. Violates safe GET semantics and amplifies load. See `docs/reviews/get-side-effect-audit.md`.

2. **Frontend team API path and method mismatches** — `team.service.ts` calls `/team/*` and `PUT` for role updates; backend exposes `/teams/*` and `PATCH`. `GET/PATCH /teams` not registered (workspace routes used instead). See `docs/reviews/frontend-compatibility.md`.

3. **Frontend error client incompatible with v1 errors** — `api/client.ts` reads `body.message`; API returns `body.error.message`. All failed requests surface as generic errors. See `docs/reviews/error-contract-audit.md`.

4. **Readiness response shape mismatch** — API returns `checks[].key`; frontend expects `checks[].id`, plus `label`, `description`, `completedCount`, `totalCount`, `quickActions`. Frontend uses mock aggregation, not the API. See `docs/reviews/dto-consistency-audit.md`.

5. **15 registered routes undocumented in OpenAPI** — Including campaign CRUD, conversation detail/assign/status, all report sub-routes except dashboard, and tag contact assignment. See `docs/reviews/openapi-route-audit.md`.

---

## MEDIUM

1. **OpenAPI pagination defect** — `GET /channels` uses `PaginationMeta` schema as a query parameter (invalid OpenAPI). Most list endpoints lack documented `page`/`page_size` params.

2. **`GET /teams/users` pagination mismatch** — OpenAPI documents pagination; handler returns full array without `pageParams()`.

3. **Conversation frontend incompatibilities** — `PATCH` vs `POST` for assign; phantom routes `/conversations/unread-count` and `/conversations/:id/read`; list response envelope mismatch.

4. **Dashboard performance at 300k+ contacts** — `COUNT(*)` on contacts per dashboard/readiness request without caching. Acceptable at 50k; degrades at scale. See `docs/reviews/dashboard-performance-audit.md`.

5. **Dashboard, onboarding, readiness not wired in frontend** — UI uses mock services; API endpoints exist but are unused.

6. **`team.Service.ListMembers` / `Get`** — Membership-only gate without explicit `PermUserView` RBAC check.

---

## LOW

1. **Dead handler code** — `TeamHandler.Get` and `TeamHandler.Update` implemented but not registered in `routes.go`.

2. **OpenAPI request/response bodies** — Most operations documented as `description: OK` without schemas.

3. **Frontend phantom RBAC routes** — `GET /team/roles/:id`, `GET /team/permissions` have no backend equivalent.

4. **Documentation drift** — Older project rules reference legacy error format `{ success, message }` without `error` object.

---

## Architecture Validation

**PASS**

- Handler → Service → Repository layering enforced
- Actor model with JWT + team membership resolution
- RBAC enforced in services, not handlers
- Standard success/error response envelopes in v1 handlers
- API versioned at `/api/v1`
- No business logic in handlers (thin DTO mapping)

---

## API Contract Validation

**FAIL**

- OpenAPI covers ~72% of registered routes with incomplete schemas
- DTO alignment strong handler↔service; weak API↔frontend
- Pagination correct on resource lists; inconsistent on team users
- Error contract correct on backend; frontend not updated

---

## RBAC Validation

**PASS**

- No handler-level permission enforcement
- Services use `rbac.Require` via `authorize()` helpers
- Minor gaps on `ListMembers` / `team.Get` (membership-only)

---

## Multi-Tenant Validation

**FAIL**

- One CRITICAL finding: link clicks endpoint lacks campaign team ownership check
- All other v1 handlers correctly pass `actor.TeamID`

---

## Frontend Compatibility

**FAIL**

- Messages nested route aligned
- Team, conversations, workspace, dashboard, onboarding, and error handling not aligned
- Most feature UIs still on mock data

---

## Audit Artifacts

| Task | Document |
|------|----------|
| 1. OpenAPI vs routes | `docs/reviews/openapi-route-audit.md` |
| 2. DTO consistency | `docs/reviews/dto-consistency-audit.md` |
| 3. Error contract | `docs/reviews/error-contract-audit.md` |
| 4. Pagination | `docs/reviews/pagination-audit.md` |
| 5. Multi-tenant safety | `docs/reviews/multi-tenant-audit.md` |
| 6. RBAC | `docs/reviews/rbac-audit.md` |
| 7. GET side effects | `docs/reviews/get-side-effect-audit.md` |
| 8. Dashboard performance | `docs/reviews/dashboard-performance-audit.md` |
| 9. Frontend compatibility | `docs/reviews/frontend-compatibility.md` |
| 10. OpenAPI quality | `docs/reviews/openapi-quality-report.md` |

---

## Recommendation

**Proceed to Phase 10.3 — Contract Remediation** (code changes allowed) with this priority order:

1. **Security:** Add `teamID` filter to `ListLinkClicks` (verify campaign ownership before listing clicks).
2. **HTTP semantics:** Make readiness GET read-only; move snapshot upsert to explicit refresh or background job; decouple dashboard from full readiness recompute.
3. **Frontend alignment:** Update `team.service.ts` paths/methods, `api/client.ts` error parsing, conversation service methods, wire dashboard/readiness/onboarding to real APIs.
4. **OpenAPI completion:** Document all 53 routes with pagination params, request/response schemas, and fix invalid parameter references.
5. **Integration testing:** E2E tests for `/teams/*`, `/workspace/readiness`, `/dashboard`, `/conversations/:id/messages` against running API.

Meta integration (templates sync, campaign send workers, webhooks) should **not** begin until CRITICAL and HIGH findings are resolved.

---

*Phase 10.2 complete. Awaiting review.*

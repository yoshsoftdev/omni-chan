# OpenAPI vs Route Audit

**Audited:** `docs/openapi/v1.yaml` vs `services/backend-service/internal/api/v1/routes.go`  
**Date:** 2026-06-22  
**Scope:** Registered v1 routes only (`/api/v1` prefix implied)

## Summary

| Result | Count |
|--------|-------|
| PASS   | 38    |
| FAIL   | 15    |
| **Total registered routes** | **53** |

15 registered routes are missing or materially incomplete in OpenAPI. Several OpenAPI entries have invalid parameter references or document pagination that handlers do not implement.

---

## Route Matrix

| Method | Path | In OpenAPI | Method Match | Schema Match | Result | Notes |
|--------|------|------------|--------------|--------------|--------|-------|
| GET | `/health` | Yes | Yes | Partial | **PASS** | Response body not fully specified |
| POST | `/auth/register` | Yes | Yes | Partial | **PASS** | Request/response bodies minimal |
| POST | `/auth/login` | Yes | Yes | Partial | **PASS** | |
| POST | `/auth/refresh` | Yes | Yes | Partial | **PASS** | |
| POST | `/auth/logout` | Yes | Yes | Partial | **PASS** | |
| GET | `/auth/me` | Yes | Yes | Partial | **PASS** | |
| GET | `/workspace` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/workspace` | Yes | Yes | Partial | **PASS** | |
| GET | `/workspace/quotas` | Yes | Yes | Partial | **PASS** | |
| GET | `/workspace/readiness` | Yes | Yes | Partial | **PASS** | `ReadinessResponse` schema present |
| GET | `/teams/members` | Yes | Yes | Partial | **PASS** | |
| POST | `/teams/members` | Yes | Yes | Partial | **PASS** | |
| GET | `/teams/users` | Yes | Yes | **FAIL** | **FAIL** | OpenAPI documents `page`/`page_size`; handler returns full `{ users }` array without pagination |
| POST | `/teams/users/invite` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/teams/users/{id}/role` | Yes | Yes | Partial | **PASS** | |
| POST | `/teams/users/{id}/deactivate` | Yes | Yes | Partial | **PASS** | |
| DELETE | `/teams/users/{id}` | Yes | Yes | Partial | **PASS** | |
| GET | `/teams/roles` | Yes | Yes | Partial | **PASS** | |
| GET | `/channels` | Yes | Yes | **FAIL** | **FAIL** | Invalid `$ref: '#/components/schemas/PaginationMeta'` used as query parameter (should be individual `page`/`page_size` params) |
| POST | `/channels` | Yes | Yes | Partial | **PASS** | No request body schema |
| GET | `/channels/{id}` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/channels/{id}` | Yes | Yes | Partial | **PASS** | No request body schema |
| DELETE | `/channels/{id}` | Yes | Yes | Partial | **PASS** | |
| GET | `/templates` | Yes | Yes | Partial | **FAIL** | Missing `page`/`page_size` query params |
| GET | `/templates/{id}` | Yes | Yes | Partial | **PASS** | |
| POST | `/templates/sync` | Yes | Yes | Partial | **PASS** | |
| GET | `/contacts` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/contacts` | Yes | Yes | Partial | **PASS** | |
| GET | `/contacts/{id}` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/contacts/{id}` | Yes | Yes | Partial | **PASS** | |
| DELETE | `/contacts/{id}` | Yes | Yes | Partial | **PASS** | |
| GET | `/segments` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/segments` | Yes | Yes | Partial | **PASS** | |
| GET | `/segments/{id}` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/segments/{id}` | Yes | Yes | Partial | **PASS** | |
| DELETE | `/segments/{id}` | Yes | Yes | Partial | **PASS** | |
| GET | `/tags` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/tags` | Yes | Yes | Partial | **PASS** | |
| GET | `/tags/{id}` | Yes | Yes | Partial | **PASS** | |
| PATCH | `/tags/{id}` | Yes | Yes | Partial | **PASS** | |
| DELETE | `/tags/{id}` | Yes | Yes | Partial | **PASS** | |
| POST | `/tags/{id}/contacts/{contact_id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| DELETE | `/tags/{id}/contacts/{contact_id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/campaigns` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/campaigns` | Yes | Yes | Partial | **PASS** | |
| GET | `/campaigns/{id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| PATCH | `/campaigns/{id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| DELETE | `/campaigns/{id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| POST | `/campaigns/{id}/send` | Yes | Yes | Partial | **PASS** | |
| POST | `/campaigns/{id}/cancel` | Yes | Yes | Partial | **PASS** | |
| GET | `/reports/dashboard` | Yes | Yes | Partial | **PASS** | |
| GET | `/reports/daily` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/reports/campaigns` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/reports/campaigns/{id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/reports/campaigns/{id}/clicks` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/conversations` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/conversations` | Yes | Yes | Partial | **PASS** | |
| GET | `/conversations/{id}` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| POST | `/conversations/{id}/assign` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| PATCH | `/conversations/{id}/status` | **No** | — | — | **FAIL** | Route registered; not in OpenAPI |
| GET | `/conversations/{id}/messages` | Yes | Yes | Partial | **FAIL** | Missing pagination query params |
| POST | `/conversations/{id}/messages` | Yes | Yes | Partial | **PASS** | |
| GET | `/onboarding` | Yes | Yes | Partial | **PASS** | |
| POST | `/onboarding/complete-step` | Yes | Yes | Partial | **PASS** | |
| GET | `/dashboard` | Yes | Yes | Partial | **PASS** | `DashboardOverview` schema present |

---

## Handler Code Not Registered (Out of Scope for Route Table)

`TeamHandler.Get` and `TeamHandler.Update` exist in `handler/team.go` but are **not** registered in `routes.go`. Frontend `team.service.ts` calls `GET/PATCH /teams`, which will 404 against the current router. Workspace team data is served at `GET/PATCH /workspace` instead.

---

## OpenAPI Quality Issues Affecting This Audit

1. **Invalid parameter reference** on `GET /channels` — `PaginationMeta` is a response schema, not a parameter object.
2. **List endpoints** — Most list routes lack documented `page` and `page_size` query parameters despite handlers enforcing them.
3. **Request bodies** — Create/update operations generally lack request body schemas.
4. **Response bodies** — Most `200` responses are `description: OK` only, without `$ref` to entity or `PageResponse` schemas.
5. **Error responses** — `Unauthorized`/`Forbidden` components exist but are not referenced on most operations.

---

## Verdict

**FAIL** — OpenAPI is incomplete relative to the live route table. Core auth, workspace, and dashboard paths are documented; CRUD sub-routes for campaigns and conversations, all report sub-routes, and tag contact assignment routes are missing.

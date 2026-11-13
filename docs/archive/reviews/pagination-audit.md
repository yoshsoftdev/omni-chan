# Pagination Audit

**Standard:** `?page=` and `?page_size=` only; `page >= 1`; `page_size <= 100`; reject `?limit=`

**Implementation:** `handler/base.go` → `pageParams()`

**Date:** 2026-06-22

---

## `pageParams()` Validation

| Rule | Implemented | Result |
|------|-------------|--------|
| Query param `page` (default from `repository.DefaultPage`) | Yes | PASS |
| Query param `page_size` (default from `repository.DefaultPageSize`) | Yes | PASS |
| Reject `?limit=` | Yes — returns `validation_error`: "use page_size instead" | PASS |
| `page >= 1` | Yes | PASS |
| `page_size` between 1 and 100 (`repository.MaxPageSize`) | Yes | PASS |

---

## List Endpoints Using Pagination

| Endpoint | Uses `pageParams()` | Response Shape | Result |
|----------|---------------------|----------------|--------|
| `GET /channels` | Yes | `PageResponse` | PASS |
| `GET /templates` | Yes | `PageResponse` | PASS |
| `GET /contacts` | Yes | `PageResponse` | PASS |
| `GET /segments` | Yes | `PageResponse` | PASS |
| `GET /tags` | Yes | `PageResponse` | PASS |
| `GET /campaigns` | Yes | `PageResponse` | PASS |
| `GET /conversations` | Yes | `PageResponse` | PASS |
| `GET /conversations/:id/messages` | Yes | `PageResponse` | PASS |
| `GET /reports/campaigns` | Yes | `PageResponse` | PASS |
| `GET /reports/campaigns/:id/clicks` | Yes | `PageResponse` | PASS |

---

## List Endpoints NOT Using Pagination

| Endpoint | Behavior | Documented in OpenAPI | Result |
|----------|----------|----------------------|--------|
| `GET /teams/members` | Returns full `{ members: [] }` | No pagination params | PASS (intentional small set) |
| `GET /teams/users` | Returns full `{ users: [] }` | **Documents** `page`/`page_size` | **FAIL** — OpenAPI/handler mismatch |
| `GET /teams/roles` | Returns full `{ roles: [] }` | No pagination | PASS |
| `GET /reports/daily` | Date-range query, not page-based | N/A | PASS |
| `GET /reports/dashboard` | Aggregated stats | N/A | PASS |
| `GET /reports/campaigns/:id` | Single resource funnel | N/A | PASS |

---

## Frontend Pagination Usage

Most frontend services do not yet pass `page`/`page_size` query parameters. Backend defaults apply (typically page 1, default page size). This is compatible at runtime but may cause silent truncation as data grows.

| Service | Passes pagination | Result |
|---------|-------------------|--------|
| `conversation.service.ts` | No | MEDIUM |
| `team.service.ts` (`listUsers`) | Optional `?status=` only | MEDIUM |
| Broadcast/contacts services | Mixed / mock | Out of active API scope |

---

## `limit` Parameter

No v1 handler accepts `limit` as an alias. Attempting `?limit=50` returns:

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "limit: use page_size instead"
  }
}
```

**PASS**

---

## OpenAPI Pagination Documentation

`PaginationMeta` and `PageResponse` schemas exist in `docs/openapi/v1.yaml`. Most list operations do not reference them in parameters or response bodies. See `openapi-route-audit.md`.

---

## Verdict

**PASS** (runtime behavior on paginated endpoints) / **FAIL** (OpenAPI documents pagination on `GET /teams/users` where handler does not paginate)

Pagination enforcement is consistent on all resource list endpoints that use `pageParams()`. Team user listing is the sole handler/OpenAPI inconsistency.

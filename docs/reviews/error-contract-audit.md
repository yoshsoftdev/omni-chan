# Error Contract Audit

**Expected v1 failure shape:**

```json
{
  "success": false,
  "error": {
    "code": "validation_error | unauthorized | forbidden | not_found | conflict | quota_exceeded | feature_not_available | internal_error",
    "message": "human-readable message"
  }
}
```

**Implementation:** `services/backend-service/internal/api/response/errors.go`

**Date:** 2026-06-22

---

## v1 Handler Layer

All active v1 handlers in `internal/api/v1/handler/` use typed helpers from `response/errors.go`:

| Helper | HTTP Status | Code | Used By |
|--------|-------------|------|---------|
| `BadRequest` | 400 | `validation_error` | Body parse failures, invalid UUIDs, `pageParams` validation |
| `Unauthorized` | 401 | `unauthorized` | Missing/invalid actor, auth failures |
| `Forbidden` | 403 | `forbidden` | `mapError` → `serviceerrors.IsForbidden` |
| `NotFound` | 404 | `not_found` | `mapError` → `serviceerrors.IsNotFound` |
| `Conflict` | 409 | `conflict` | `mapError` → `serviceerrors.ErrConflict` |
| `QuotaExceeded` | 402 | `quota_exceeded` | `mapError` → `serviceerrors.ErrQuotaExceeded` |
| `FeatureNotAvailable` | 501 | `feature_not_available` | Template sync stub |
| `InternalError` | 500 | `internal_error` | Unmapped errors in `mapError` |

`mapError` in `handler/base.go` correctly maps service-layer errors to semantic codes. Unmapped errors fall through to `internal_error` (does not leak raw error strings for unknown types). **PASS**

`requireActor` failures consistently return `Unauthorized`. **PASS**

---

## Middleware

`middleware.Auth` and `middleware.ResolveActor` use `response.Unauthorized` with the new envelope. **PASS**

---

## Legacy Response Path

`response.Fail(c, status, message)` in `response/response.go` delegates to `fail()` with `CodeInternalError`. This produces the correct **envelope shape** but always uses `internal_error` code regardless of HTTP status.

Legacy handlers tagged `//go:build legacy` or outside the v1 router are **not mounted** on the active API. They do not affect the live v1 contract.

---

## Error Code Coverage in v1 Handlers

| Code | Exercised in v1 | Notes |
|------|-----------------|-------|
| `validation_error` | Yes | `pageParams`, body parser, UUID parse |
| `unauthorized` | Yes | All protected routes |
| `forbidden` | Yes | RBAC denials via `mapError` |
| `not_found` | Yes | Entity lookups |
| `conflict` | Yes | Duplicate slug, etc. |
| `quota_exceeded` | Yes | Team invite / member limits |
| `feature_not_available` | Yes | `POST /templates/sync` |
| `internal_error` | Yes | Default fallback |

All eight documented codes are implemented and reachable. **PASS**

---

## Legacy Format Remnants

| Location | Issue | Severity |
|----------|-------|----------|
| `frontend/src/services/api/client.ts` | Reads `body.message` on failure; v1 API returns `body.error.message` | **HIGH** |
| `frontend/src/types/api.ts` | `ApiResponse` has top-level `message` only; no `error` field | MEDIUM |
| Project rules (`.cursor/rules`) | Older docs mention `{ success, message }` for errors | LOW (docs drift) |

No v1 handler returns the legacy top-level-only `message` failure format.

---

## Auth Handler

`handler/auth.go` uses dedicated error mapping for login/register failures and falls back to `mapError` for service errors. All paths use `response/errors.go` helpers. **PASS**

---

## Verdict

**PASS** (backend v1 handlers) / **FAIL** (frontend client)

Active v1 API handlers uniformly emit the standardized error envelope with semantic codes. The primary contract gap is **frontend error parsing**, which will surface generic `"Request failed"` for all API errors until updated to read `error.message` and optionally `error.code`.

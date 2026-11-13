# OpenAPI Quality Report

**Spec:** `docs/openapi/v1.yaml`  
**Compared against:** `internal/api/v1/routes.go` and handler implementations  
**Date:** 2026-06-22

---

## Checklist

| Requirement | Present | Quality | Result |
|-------------|---------|---------|--------|
| JWT bearer auth (`bearerAuth`) | Yes | Applied on protected routes | **PASS** |
| Error schemas (`ErrorResponse`) | Yes | 8 codes enumerated; not referenced on most operations | **PARTIAL** |
| Pagination schemas (`PaginationMeta`, `PageResponse`) | Yes | Invalid use as query param on `/channels`; missing on most lists | **FAIL** |
| Enums | Partial | Readiness `status` enum; missing campaign/conversation status enums | **PARTIAL** |
| Request bodies | Minimal | Most POST/PATCH lack `requestBody` schemas | **FAIL** |
| Response bodies | Minimal | Most responses are `description: OK` only | **FAIL** |
| Undocumented routes | 15 routes | See `openapi-route-audit.md` | **FAIL** |
| Server URL `/api/v1` | Yes | Matches router mount | **PASS** |
| Tags | Yes | Match feature areas | **PASS** |

---

## Security Scheme

```yaml
securitySchemes:
  bearerAuth:
    type: http
    scheme: bearer
    bearerFormat: JWT
```

Protected operations include `security: [{ bearerAuth: [] }]`. Auth register/login/refresh correctly omit security. **PASS**

---

## Error Schema Quality

`ErrorResponse` correctly documents:

- `validation_error`
- `unauthorized`
- `forbidden`
- `not_found`
- `conflict`
- `quota_exceeded`
- `feature_not_available`
- `internal_error`

Reusable responses `Unauthorized` and `Forbidden` exist under `components.responses` but are **not linked** from individual path operations. Consumers cannot codegen error handling per endpoint.

---

## Pagination Schema Quality

`PaginationMeta` fields align with `dto.PaginationMeta` in Go:

- `page`, `page_size`, `total`, `total_pages`

**Defect:** `GET /channels` references `$ref: '#/components/schemas/PaginationMeta'` under `parameters`. OpenAPI requires `Parameter` objects, not schema objects. This is **invalid OpenAPI 3.0**.

**Missing:** Standard reusable parameters:

```yaml
parameters:
  PageParam:
    in: query
    name: page
    schema: { type: integer, minimum: 1, default: 1 }
  PageSizeParam:
    in: query
    name: page_size
    schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
```

---

## Documented vs Undocumented Routes

**Documented:** 38 of 53 registered routes (~72%)

**Undocumented (15):**

1. `GET /campaigns/{id}`
2. `PATCH /campaigns/{id}`
3. `DELETE /campaigns/{id}`
4. `GET /conversations/{id}`
5. `POST /conversations/{id}/assign`
6. `PATCH /conversations/{id}/status`
7. `GET /reports/daily`
8. `GET /reports/campaigns`
9. `GET /reports/campaigns/{id}`
10. `GET /reports/campaigns/{id}/clicks`
11. `POST /tags/{id}/contacts/{contact_id}`
12. `DELETE /tags/{id}/contacts/{contact_id}`

Plus pagination parameter gaps on documented list routes (counts as quality failures, not missing routes).

---

## Entity Schemas

Only these domain schemas are defined:

- `ReadinessResponse`
- `DashboardOverview`
- `OnboardingSnapshot`
- `SuccessResponse` / `ErrorResponse` / `PageResponse`

Missing entity schemas for: Team, Contact, Channel, Template, Segment, Tag, Campaign, Conversation, Message, Report types.

---

## Request Body Coverage

| Operation | Request body documented |
|-----------|------------------------|
| `POST /auth/register` | Partial (inline properties) |
| `POST /auth/login` | Partial |
| `PATCH /workspace` | No |
| `POST /channels` | No |
| `POST /contacts` | No |
| `POST /campaigns` | No |
| `POST /conversations/:id/messages` | No |
| All other mutations | No |

---

## Spec Validity

The invalid `PaginationMeta` parameter reference may cause OpenAPI validators (Swagger UI, codegen tools) to fail or skip parameter validation for `GET /channels`.

Recommend running `npx @redocly/cli lint docs/openapi/v1.yaml` in a future hardening phase.

---

## Verdict

**FAIL** — Foundation is present (auth, error model, a few response schemas) but the spec is not production-ready for codegen or external consumers. ~28% of routes undocumented; request/response bodies largely absent; pagination incorrectly specified.

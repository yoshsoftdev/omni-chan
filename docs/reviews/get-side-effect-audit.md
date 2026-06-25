# GET Side-Effect Audit

**Rule:** `GET` requests must not INSERT, UPDATE, or DELETE persistent state.

**Date:** 2026-06-22

---

## Methodology

Reviewed all `GET` routes in `internal/api/v1/routes.go` and traced handler → service → repository call chains for write operations.

---

## Clean GET Endpoints (No Writes)

| Endpoint | Read-only | Result |
|----------|-----------|--------|
| `GET /health` | Yes | PASS |
| `GET /auth/me` | Yes | PASS |
| `GET /workspace` | Yes | PASS |
| `GET /workspace/quotas` | Yes | PASS |
| `GET /teams/members` | Yes | PASS |
| `GET /teams/users` | Yes | PASS |
| `GET /teams/roles` | Yes | PASS |
| `GET /channels`, `GET /channels/:id` | Yes | PASS |
| `GET /templates`, `GET /templates/:id` | Yes | PASS |
| `GET /contacts`, `GET /contacts/:id` | Yes | PASS |
| `GET /segments`, `GET /segments/:id` | Yes | PASS |
| `GET /tags`, `GET /tags/:id` | Yes | PASS |
| `GET /campaigns`, `GET /campaigns/:id` | Yes | PASS |
| `GET /reports/dashboard` | Yes (aggregations only) | PASS |
| `GET /reports/daily` | Yes | PASS |
| `GET /reports/campaigns` | Yes | PASS |
| `GET /reports/campaigns/:id` | Yes | PASS |
| `GET /reports/campaigns/:id/clicks` | Yes | PASS |
| `GET /conversations`, `GET /conversations/:id` | Yes | PASS |
| `GET /conversations/:id/messages` | Yes | PASS |

---

## HIGH Severity — GET Writes State

### 1. `GET /workspace/readiness`

**Chain:** `WorkspaceHandler.GetReadiness` → `workspace.Service.GetReadiness` → `computeReadiness` → `snapshots.Upsert`

**File:** `internal/workspace/readiness.go:79-93`

Every readiness request recomputes checks and **upserts** `workspace_readiness_snapshots` (INSERT or UPDATE).

**Severity:** **HIGH** — violates HTTP idempotency expectations for GET.

---

### 2. `GET /onboarding`

**Chain:** `WorkspaceHandler.GetOnboarding` → `workspace.Service.GetOnboarding` → `loadOnboardingState` → `computeReadiness` → **same `Upsert`**

**File:** `internal/workspace/onboarding.go` (calls `computeReadiness`)

**Severity:** **HIGH**

---

### 3. `GET /dashboard`

**Chain:** `DashboardHandler.Get` → `dashboard.Service.GetOverview` → `workspace.Service.GetReadiness` → **`Upsert`**

Additionally performs 4× `COUNT(*)` queries plus full readiness evaluation (5+ queries).

**File:** `internal/dashboard/service.go:106-112`

**Severity:** **HIGH** — GET dashboard triggers snapshot write as side effect of readiness score inclusion.

---

## Special Attention Endpoints (Per Task Spec)

| Endpoint | Side effects | Severity |
|----------|--------------|----------|
| `GET /workspace/readiness` | Upsert snapshot | HIGH |
| `GET /dashboard` | Upsert via readiness + counts | HIGH |
| `GET /onboarding` | Upsert via readiness | HIGH |

---

## Recommendations (Documentation Only — Out of Scope for 10.2)

1. Split readiness into `GET` (read cached snapshot) and `POST /workspace/readiness/refresh` (compute + persist).
2. Dashboard should read cached readiness score from snapshot without recomputing on every request.
3. Onboarding GET should read persisted onboarding state without triggering readiness recomputation.

---

## Verdict

**FAIL** — Three GET endpoints perform database writes. All three are flagged **HIGH** per audit criteria.

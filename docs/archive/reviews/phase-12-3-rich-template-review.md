# Phase 12.3 — Rich Template Rendering & Send Validation Review

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Rich template payload builder, pre-send validation, worker integration — no orchestration, webhook, or frontend changes

---

## Build Status

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./...` | PASS |

---

## Phase 12.2 Finding Validation

| 12.2 Finding | Status |
|--------------|--------|
| BODY-only `BuildTemplateComponents` | **Resolved** — replaced by `pkg/meta/template_builder.go` |
| HEADER/FOOTER/BUTTONS/media not expanded | **Resolved** for send-time payload generation |
| Rich templates not production-ready | **Resolved** with validation gate before Meta API |

---

## Architecture Review

### Render pipeline

```text
Template.Components (synced JSON)
        ↓
campaign.VariableMapping + contact + sample_values
        ↓
campaign.ValidateTemplateSend (send_validation.go)
        ↓
meta.BuildTemplatePayload (template_builder.go)
        ↓
worker.Sender → meta.SendTemplateMessage
```

Worker contains no template business logic — only orchestrates validation, build, and Meta call.

### Component support

| Component | Send behavior |
|-----------|---------------|
| HEADER TEXT | `{{n}}` resolved from `variable_mapping.header.text` |
| HEADER IMAGE/VIDEO/DOCUMENT | `variable_mapping.header.media_link` absolute URL |
| BODY | `{{n}}` resolved from `variable_mapping.body` (ordered) |
| FOOTER | Static in approved template — no send component (correct) |
| BUTTON URL | Dynamic suffix from `variable_mapping.buttons[n].url` |
| BUTTON COPY_CODE | `variable_mapping.buttons[n].copy_code` |
| BUTTON QUICK_REPLY | Static — no send parameters |

### Validation layer (`internal/campaign/send_validation.go`)

| Check | Failure code | Retry |
|-------|--------------|-------|
| Missing variable mapping/value | `missing_variable` | No |
| Invalid media URL (non-absolute http/https) | `invalid_media_url` | No |
| Missing/invalid button parameters | `invalid_button_parameter` | No |
| Builder dry-run failure | `missing_variable` | No |

Validation runs **before** Meta API — confirmed by `TestValidationFailureNoMetaCall`.

### Failure classification extension

`ClassifySendError` maps `SendValidationError` to `FailurePermanent` with codes:
- `missing_variable`
- `invalid_media_url`
- `invalid_button_parameter`

### Metrics

| Metric | Location |
|--------|----------|
| `template_validation_failed_total` | `worker.Metrics` |
| `template_render_success_total` | `worker.Metrics` |
| `template_render_failed_total` | `worker.Metrics` |

### Logging

Sender logs: `campaign_id`, `recipient_id`, `template_id`, `channel_id`, `phone_number_id`.  
Does not log phone numbers, message content, access tokens, or OTP/copy codes.

### API v1 contract

`variable_mapping` extended additively:

```json
{
  "header": { "text": { "1": "name" }, "media_link": "https://..." },
  "body": { "1": "name", "2": "order_id" },
  "buttons": { "0": { "url": { "1": "slug" } }, "1": { "copy_code": "SAVE20" } }
}
```

Existing `body`-only campaigns remain compatible.

---

## Deliverables Checklist

| Component | Status |
|-----------|--------|
| `pkg/meta/template_builder.go` | Done |
| `pkg/meta/template_types.go` | Done |
| `internal/campaign/send_validation.go` | Done |
| Extended `VariableMapping` | Done |
| Worker sender integration | Done |
| Failure classifier extension | Done |
| Template render metrics | Done |
| Tests (all required scenarios) | Done |

---

## Tests

| Test | Coverage |
|------|----------|
| `TestBuildTextHeader` | TEXT header variables |
| `TestBuildImageHeaderMediaURL` | IMAGE header media link |
| `TestBuildURLButtonSubstitution` | URL button dynamic param |
| `TestBuildMissingVariableFails` | Builder rejects missing vars |
| `TestBuildCopyCodeButton` | COPY_CODE coupon_code payload |
| `TestBuildComplexTemplate` | IMAGE + BODY + FOOTER + URL + COPY_CODE |
| `TestValidateMissingVariableNoMetaCall` | Validation before send |
| `TestValidateInvalidMediaURL` | Relative URL rejected |
| `TestValidateURLButtonRequiresMapping` | Button param validation |
| `TestValidateCopyCodeRequired` | COPY_CODE required |
| `TestValidationFailureNoMetaCall` | Worker end-to-end, no Meta on validation fail |
| `TestBuildTemplateComponentsBodyOrder` | Parameter ordering preserved |

---

## Findings

### MEDIUM

1. **Media header uses `link` URL, not Meta upload handle** — Builder supports absolute `https://` links per Cloud API send docs. Templates created with Meta resumable-upload handles only (no public URL) require a follow-up to resolve handles → sendable media IDs.

2. **`sample_values` used as fallback resolver** — Custom fields not on `contact` (e.g. `order_number`) resolve via `sample_values` index/key. Campaign creation UI still primarily maps body fields; rich header/button mapping is API-ready but not yet exposed in frontend.

3. **COPY_CODE via `OTP` + `otp_type` not distinguished in builder** — Builder handles `COPY_CODE` button type from synced templates. Legacy `OTP`/`COPY_CODE` auth templates use existing `BuildKodeOTPComponents` path outside broadcast worker.

### LOW

4. **Footer not emitted in send payload** — Correct per Meta spec (static in template definition). Documented in builder.

5. **QUICK_REPLY buttons skipped at send** — Correct (static). No parameters required.

6. **Metrics in-memory only** — Counters exposed via `Metrics.Snapshot()`; no Prometheus scrape endpoint in this phase.

---

## Recommendation

**APPROVE WITH FINDINGS**

Phase 12.3 closes the Phase 12.2 gap: approved template component types are rendered into valid Meta send payloads with pre-send validation and permanent failure classification. Rich marketing templates with public media URLs, dynamic URL buttons, and copy-code buttons are production-ready on the backend. Address media-handle resolution (MEDIUM #1) before templates that rely exclusively on Meta-uploaded header media.

---

## Out of Scope (confirmed unchanged)

Frontend, webhooks, delivery tracking, orchestration, rate limiter, campaign completion, analytics

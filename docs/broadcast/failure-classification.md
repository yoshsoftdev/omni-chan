# Failure Classification

Failed delivery webhooks store structured failure metadata on `campaign_recipients`.

## Columns

| Column | Purpose |
|--------|---------|
| `failure_code` | Normalized category for reporting |
| `last_error` | Human-readable reason from Meta |

## Categories

| Code | Triggers |
|------|----------|
| `invalid_number` | Invalid recipient, error 131026 |
| `blocked` | User blocked business, error 131047 |
| `rate_limited` | Throughput / 130429 / rate limit text |
| `template_error` | Template parameter or approval issues |
| `unknown` | Default when no pattern matches |

## Source

`delivery.ClassifyFailure` inspects Meta error text from:

1. Canonical event failure text (when present)
2. `payload.raw` → `statuses[].errors[]`

## Rules

- Classification runs only on transition to `failed`
- Message body and access tokens are never logged
- Counters increment `failed_count` once per recipient on first `failed` transition

## Example

```
Meta: "Message failed: invalid recipient number (131026)"
  → failure_code = invalid_number
  → last_error = Message failed: invalid recipient number (131026)
```

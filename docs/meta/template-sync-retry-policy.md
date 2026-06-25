# Template Sync Retry Policy

## Retryable Failures

| Failure | Action |
|---------|--------|
| Timeout | Retry with backoff |
| HTTP 5xx | Retry with backoff |
| Network error | Retry with backoff |
| HTTP 429 | Retry with backoff |

## Non-Retryable Failures

| Failure | Action |
|---------|--------|
| Invalid / expired token (401, 403) | Fail immediately |
| Missing WABA (404) | Fail immediately |

## Backoff Schedule

```text
Attempt 1 → wait 30s → Attempt 2 → wait 2m → Attempt 3 → wait 10m → FAILED
```

Implementation: `pkg/meta.Retry` with `RetryDelays = []time.Duration{30s, 2m, 10m}`.

## Security

- Retry logs never include `access_token`
- API error bodies are truncated in logs
- Invalid token surfaces as validation error to client (no token echo)

## Pagination + Retry

Each **page fetch** is retried independently. A failed page does not commit partial upserts — the sync run is marked `failed`.

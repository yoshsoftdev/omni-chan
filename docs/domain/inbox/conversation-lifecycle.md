# Conversation Lifecycle

## States

| Status | Meaning |
|--------|---------|
| `open` | Active conversation |
| `pending` | Awaiting agent action |
| `resolved` | Issue resolved |
| `closed` | Archived |

## Inbound Rules

| Current | Inbound message |
|---------|-----------------|
| `resolved` | → `open` |
| `closed` | → `open` |
| `open` | unchanged |
| `pending` | unchanged |

`resolution_at` is cleared when reopening from `resolved` or `closed`.

## Resolution

Setting status to `resolved` or `closed` sets `resolution_at` (first time only).

# SLA Tracking

## Stored Timestamps

| Field | Set when |
|-------|----------|
| `first_response_at` | First outbound agent message |
| `last_customer_message_at` | Inbound customer message |
| `last_agent_message_at` | Outbound agent message |
| `resolution_at` | Status → `resolved` or `closed` |

## Derived Metrics (Dashboard)

| Metric | Calculation |
|--------|-------------|
| First response time | `first_response_at - last_customer_message_at` |
| Resolution time | `resolution_at - created_at` |

Exposed on `GET /api/v1/dashboard` as `avg_first_response_seconds` and `avg_resolution_seconds`.

## Source of Truth

All timestamps live on `conversations` in PostgreSQL. WebSocket events are projections only.

# Services

| Service | Port (default) | Exposure | Responsibility |
|---------|----------------|----------|----------------|
| `webhook-service` | 8081 | Public | Webhook verify, signature check, normalize, NATS publish |
| `backend-service` | 8090 | Internal | REST API, auth, DB, NATS consume, business logic |

Shared code: `../shared/`

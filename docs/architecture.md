Frontend

Next.js
↓
REST API (internal)
↓
Backend Service

Webhook Service (public)

Meta / Telegram
↓
Webhook Service
↓
NATS JetStream
↓
Backend Service (consumer)
↓
Service Layer
↓
Repository Layer

Infrastructure

PostgreSQL
Redis
NATS JetStream
MinIO

Realtime

Browser
↓
WebSocket
↓
Hub
↓
Redis PubSub

See `docs/migration-webhook-service.md` for webhook split details.
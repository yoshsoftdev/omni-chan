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

Instagram outbound media:

Backend Service (private)
↓ mint short-lived media token
Meta receives `https://<public-webhook>/media/{token}`
↓ synchronous HTTP GET
Webhook Service (public; validates media token)
↓ HTTP internal + `X-Internal-Token`
Backend Service `/api/v1/internal/media/{token}` (private; validates again)
↓ HTTP internal
Storage Service (private)
↓ streamed response
Webhook Service
↓ streamed response
Meta

Media download does not use NATS because Meta requires an immediate HTTP
response. Only Webhook Service is internet-facing; Backend and Storage remain
on the internal network.

If Instagram send fails with `error_subcode` **2018047** (“Failed to upload
attachment” / “Gagal mengunggah lampiran”), Meta could not download the public
media URL (or the MIME/type did not match). Verify:

1. `WEBHOOK_PUBLIC_BASE_URL` is the public HTTPS host for webhook-service
2. That host proxies `GET /media/{token}` to a webhook build that includes the route
3. Webhook `MEDIA_TOKEN_SECRET` matches backend; `MEDIA_BACKEND_URL` reaches Main API
4. Outbound images are JPEG/PNG (not WebP); videos are MP4/MOV/WebM/OGG/AVI up to 25MB
5. `ValidateAttachment` for Instagram uses IG limits (not WhatsApp 16MB video cap)

Realtime

Browser
↓
WebSocket
↓
Hub
↓
Redis PubSub

See `docs/migration-webhook-service.md` for webhook split details.
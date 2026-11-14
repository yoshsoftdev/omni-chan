# Arsitektur — Ringkasan

Detail alur data dan risikonya ada di [data-flow.md](data-flow.md); versi dan pilihan teknologi di [tech-stack.md](tech-stack.md).
Port tiap service **berbeda per environment**: lihat tabel "Port per environment" di [README root](../../README.md).

## Service

| Service | Terbuka ke internet | Tugas |
|---|---|---|
| frontend (Next.js) | ya, lewat nginx | Dashboard; mem-proxy `/api/*` dan `/ws` ke backend |
| webhook-service | **ya, satu-satunya service publik** | Terima webhook Meta/Instagram, invoke automation publik, proxy media untuk Meta |
| backend-service `cmd/api` | tidak | REST API v1, WebSocket |
| backend-service `cmd/event-consumer` | tidak | Consumer event webhook `meta.>` dari NATS (replika berbagi durable consumer) |
| backend-service `cmd/scheduler` | tidak | Job terjadwal: window recontact, rekonsiliasi kampanye, purge `processed_events` (tiap job memegang advisory lock Postgres) |
| backend-service `cmd/broadcast-orchestrator` / `cmd/broadcast-worker` | tidak | Penjadwalan dan pengiriman kampanye |
| storage-service | tidak | Upload dan penyimpanan file |

Infrastruktur: PostgreSQL 16, Redis 7 (pub/sub WebSocket, rate limit, idempotensi), NATS JetStream 2.10 (event webhook & broadcast), disk lokal atau MinIO.

## Alur utama

```
Meta ──webhook──▶ webhook-service ──publish──▶ NATS ──▶ backend consumer ──▶ Postgres ──▶ Redis pub/sub ──▶ WebSocket ──▶ browser
browser ──/api/v1──▶ backend ──▶ Meta Graph API
orchestrator ──outbox──▶ NATS ──▶ broadcast-worker ──▶ Meta
```

Jika NATS mati, webhook-service membalas **503** ke Meta agar Meta mengirim ulang; payload rusak tetap dibalas 200 dan masuk DLQ.

## Media Instagram

```
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
```

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

Penjelasan pemisahan webhook-service: [archive/phases/migration-webhook-service.md](../archive/phases/migration-webhook-service.md).

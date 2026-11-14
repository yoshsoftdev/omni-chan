# Omni Chan v2

Omnichannel SaaS platform (monorepo parent + git submodules).

## Clone (repo dalam repo)

Parent repo berisi submodule: `backend`, `frontend`, `webhook-service`, `storage-service`.

```bash
# Opsi 1 — script
./scripts/git-clone.sh https://github.com/yoshsoftdev/omni-chan.git

# Opsi 2 — manual
git clone --recurse-submodules https://github.com/yoshsoftdev/omni-chan.git
cd omni-chan
```

Jika sudah clone tanpa submodule:

```bash
git submodule update --init --recursive
```

Setup parent repo pertama kali (dari folder yang sudah punya nested `.git`):

```bash
./scripts/git-setup-submodules.sh
```

## Backend

```bash
cd services/backend-service
cp .env.example .env
make migrate-up
make run
```

Default port: `8070`

## Webhook service

```bash
cd services/webhook-service
cp .env.example .env
make run
```

Default port: `8090`

## Frontend

```bash
cd frontend
cp .env.example .env
pnpm install
pnpm exec next dev -p 8095
```

## Storage service

```bash
cd services/storage-service
cp .env.example .env
make migrate-up
make run
```

Default port: `8082`

## Port per environment

Port berbeda di tiap cara menjalankan. Selalu cek tabel ini dan file `.env` yang dipakai.

| Service | Default di kode | docker-compose | systemd (`deploy/`) | Dev LAN (`.env` contoh) |
|---|---|---|---|---|
| frontend | 3000 (`next dev`) | – | 8095 | 8095 |
| backend-service | 8080 | 8090 (internal) | 8070 | 8066 |
| webhook-service | 8081 | 8081 (publik) | 8090 | 8060 |
| storage-service | 8082 | 8082 (internal) | – | 8065 |

> Port 8090 dipakai **backend** di docker-compose tapi **webhook** di systemd. Jangan menyalin `.env` antar cara deploy tanpa menyesuaikan port.

## Test

```bash
make test-race       # semua modul Go dengan race detector
make test-frontend   # vitest + typecheck
make test-e2e        # Playwright (backend di-mock); sekali saja: cd frontend && pnpm test:e2e:install
make test-all        # ketiganya
(cd frontend && pnpm test:coverage)   # vitest + gate coverage
# Integration test Postgres (database kosong, sudah dimigrasi):
(cd services/backend-service && DATABASE_URL=postgres://... APP_ENV=development go test -tags=integration ./...)
```

CI: `.github/workflows/ci.yml` di repo ini dan di tiap submodule (butuh secret `SUBMODULES_TOKEN` / `PARENT_REPO_TOKEN` bila repo private).

## Monitoring (opsional)

```bash
GRAFANA_ADMIN_PASSWORD=... docker compose -f docker-compose.yml -f docker-compose.observability.yml up -d
```

Grafana di `http://127.0.0.1:3001` (folder **Omni Chan**: Ingress webhook, Inbox realtime, Broadcast, Infra). Tiap proses juga punya `/metrics`, `/health`, `/ready` di port internal `METRICS_ADDR` (default `127.0.0.1:9101` api, `9102` worker, `9103` orchestrator, `9104` webhook, `9105` storage) yang tidak pernah dipublikasikan. Trace diekspor bila `OTEL_EXPORTER_OTLP_ENDPOINT` diisi.

## Structure

```
services/
  backend-service/   — REST API + WebSocket (submodule)
  webhook-service/   — Meta/Telegram ingress (submodule)
  storage-service/   — Media/File storage service (submodule)
shared/                — shared Go packages (parent repo)
frontend/              — Next.js (submodule)
docs/                  — mulai dari docs/README.md (PRD, fitur, arsitektur, roadmap)
deploy/                — systemd + nginx
```

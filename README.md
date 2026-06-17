# Omni Chan v2

Omnichannel SaaS platform (monorepo parent + git submodules).

## Clone (repo dalam repo)

Parent repo berisi submodule: `backend`, `frontend`, `webhook-service`.

```bash
# Opsi 1 — script
./scripts/git-clone.sh git@git.int.visibel.id:omni-chan/omni-chan-v2.git

# Opsi 2 — manual
git clone --recurse-submodules git@git.int.visibel.id:omni-chan/omni-chan-v2.git
cd omni-chan-v2
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

## Structure

```
services/
  backend-service/   — REST API + WebSocket (submodule)
  webhook-service/     — Meta/Telegram ingress (submodule)
shared/                — shared Go packages (parent repo)
frontend/              — Next.js (submodule)
docs/                  — PRD, tasks, architecture
deploy/                — systemd + nginx
```

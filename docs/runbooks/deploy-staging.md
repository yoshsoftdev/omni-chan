# Runbook: deploy ke VM staging (Docker)

Staging = satu VM Ubuntu 24.04 amd64 (4 vCPU / 8 GB / 80 GB) yang menjalankan **seluruh stack di Docker Compose**, termasuk Grafana/Prometheus/Loki/Tempo untuk debugging. Image di-build **di laptop atau di GitHub Actions**, lalu di-push ke GHCR. VM hanya `pull` + `up`, jadi tidak perlu Go/Node/git/akses submodule di VM.

```
laptop / GitHub Actions ── build + push ──► ghcr.io/yoshsoftdev/omni-chan-{backend,webhook,storage,frontend}:<sha>
                         ── rsync compose files ──► VM:/srv/omni-chan ── docker compose pull && up --wait
```

| Container | Di VM | Untuk |
|---|---|---|
| frontend | 127.0.0.1:8095 | UI (nginx) |
| backend-service | 127.0.0.1:8070 (API + /ws), :9101 metrics | nginx, debug |
| webhook-service | 127.0.0.1:8090 | Meta/Telegram (nginx) |
| backend-event-consumer, backend-scheduler, broadcast-orchestrator, broadcast-worker, storage-service | internal | – |
| postgres / redis | 127.0.0.1:5432 / 6379 | debug via SSH tunnel |
| nats-1..3 | 127.0.0.1:8222 (monitor) | debug |
| grafana | 127.0.0.1:3001 | logs, metrics, traces |

Semua port hanya di 127.0.0.1. Yang terbuka ke internet hanya nginx (80/443) dan SSH.

## 0. Staging lokal + E2E (tanpa VM)

Stack yang sama, dijalankan di mesin ini (compose project `omni-staging`, volume terpisah dari dev):

```bash
cp .env.staging.example .env          # isi secret; tambah POSTGRES_PUBLISH_PORT=15432 / REDIS_PUBLISH_PORT=16379 kalau 5432/6379 dipakai
./deploy/deploy-staging.sh local      # build image + up --wait, tanpa registry
./deploy/e2e-staging.sh               # E2E nyata (tanpa mock): frontend/e2e-stack
cd frontend && E2E_BASE_URL=http://127.0.0.1:8095 pnpm test:e2e   # suite mock terhadap build produksi
```

Mode `local` otomatis menambahkan `docker-compose.e2e.yml`: **fake Meta Graph** (`deploy/docker/fake-graph.mjs`, port 127.0.0.1:8099) dan `META_GRAPH_BASE_URL` di semua role backend. Dengan begitu channel bisa dibuat dan balasan "terkirim" tanpa kredensial Meta. Backend **menolak** `META_GRAPH_BASE_URL` saat `APP_ENV=production`, dan VM tidak pernah memakai overlay ini. `E2E=0` mematikannya.

Yang dicek `e2e-staging.sh` (tiap test membuat team sendiri):

| Test | Jalur yang dilewati |
|---|---|
| login akun nyata / password salah | UI → Next proxy → backend → Postgres |
| connect channel WhatsApp lewat UI | UI → backend → verifikasi ke Graph (fake) → channel `connected` |
| token Meta kedaluwarsa ditolak | backend → Graph 401 → 400 ke klien |
| inbound WhatsApp sampai inbox | webhook (signature Meta) → NATS → event-consumer → Postgres → API → UI |
| pesan masuk saat thread terbuka | + WebSocket lewat Next `/ws` |
| signature palsu ditolak | webhook 401, tidak ada data |
| Meta kirim ulang wamid sama | tidak ada duplikat |
| balasan agent dari inbox | UI → backend → Graph (isi yang diterima Meta dicek) → tersimpan outbound |

Konvensi fake Graph: WABA `w<pnid>` memiliki phone `<pnid>`, token `bad-token` = kedaluwarsa, `GET :8099/_sent?to=<wa_id>` = pesan yang diterima "Meta".

Report: `frontend/playwright-report-stack/index.html`. Debug pakai `./dc logs -f <service>` di root repo. Matikan dengan `./dc down` (`-v` untuk hapus data).

## 1. Siapkan VM (sekali)

```bash
# Docker Engine + compose plugin
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER        # logout/login sesudahnya
sudo apt-get install -y nginx certbot python3-certbot-nginx rsync
sudo mkdir -p /srv/omni-chan && sudo chown $USER: /srv/omni-chan

# Izin pull image private dari GHCR: PAT classic dengan scope read:packages
echo <PAT> | docker login ghcr.io -u <github-user> --password-stdin
```

Lalu isi env:

```bash
# dari laptop
scp .env.staging.example <user>@<ip-vm>:/srv/omni-chan/.env
# di VM: isi semua nilai, lalu
chmod 600 /srv/omni-chan/.env
```

`APP_ENV=staging` memberi log level info dengan pemeriksaan keamanan sama seperti production. `APP_ENV=development` memberi log debug, tetapi juga mengaktifkan default dev yang longgar. Pakai sementara saja saat mengejar bug.

nginx: salin `deploy/nginx.example.conf` (port-nya sudah cocok: 8095/8070/8090), ganti `server_name`, lalu `sudo certbot --nginx`. Domain webhook harus sama dengan `WEBHOOK_PUBLIC_BASE_URL`.

## 2a. Deploy dari laptop

Butuh Docker lokal dan `docker login ghcr.io` (PAT dengan scope `write:packages`).

```bash
STAGING_HOST=<user>@<ip-vm> NEXT_PUBLIC_WEBHOOK_URL=https://<domain-webhook>/api/v1 ./deploy/deploy-staging.sh
```

- `./deploy/deploy-staging.sh build`: hanya build + push.
- `./deploy/deploy-staging.sh deploy`: hanya deploy (`TAG=<sha>` untuk tag tertentu).
- Tag = short sha repo induk. Kalau ada perubahan yang belum di-commit, tag diberi akhiran `-dirty`, jadi kode yang belum di-commit tetap bisa dites di staging.
- `OBSERVABILITY=0` mematikan stack Grafana.

## 2b. Deploy dari GitHub Actions

Workflow `.github/workflows/staging.yml` jalan saat ada push ke branch `staging`, atau lewat **Actions → staging → Run workflow** (pilih branch mana saja; isi `tag` untuk deploy tag lama tanpa build).

Setup sekali di GitHub (Settings → Environments → `staging`):

| Jenis | Nama | Isi |
|---|---|---|
| secret | `SUBMODULES_TOKEN` | sudah ada untuk CI |
| secret | `STAGING_HOST` | `deploy@<ip-vm>` |
| secret | `STAGING_SSH_KEY` | private key (`ssh-keygen -t ed25519 -f stag`; `stag.pub` masuk ke `~/.ssh/authorized_keys` di VM) |
| secret | `STAGING_KNOWN_HOSTS` | output `ssh-keyscan -H <ip-vm>` |
| variable | `NEXT_PUBLIC_WEBHOOK_URL` | `https://<domain-webhook>/api/v1` |

Image di GHCR dibuat private dan dimiliki akun/org `yoshsoftdev`. Kalau workflow tidak bisa push, buka Packages → package → Settings → *Manage Actions access* dan beri repo `omni-chan` akses **Write**.

## 3. Debugging di VM

Setiap deploy menulis `/srv/omni-chan/dc`: `docker compose` yang sudah terkunci ke file dan tag yang sedang jalan.

```bash
cd /srv/omni-chan
./dc ps                                   # status + health
./dc logs -f --tail=200 backend-service   # log satu service
./dc logs -f backend-event-consumer broadcast-worker
./dc restart webhook-service
./dc exec postgres psql -U postgres omni_chan
./dc exec redis redis-cli
curl -s 127.0.0.1:8222/jsz?streams=1 | jq   # NATS JetStream
head -3 dc                                # tag yang sedang jalan
```

Dari laptop lewat SSH tunnel: Grafana (Loki untuk log, Tempo untuk trace), DB, dan metrics:

```bash
ssh -N -L 3001:127.0.0.1:3001 -L 5432:127.0.0.1:5432 -L 6379:127.0.0.1:6379 -L 8222:127.0.0.1:8222 <user>@<ip-vm>
```

Lalu buka http://localhost:3001, atau sambungkan DBeaver/psql ke `localhost:5432`.

Mencoba patch cepat tanpa commit: jalankan dari laptop, hasilnya tag `<sha>-dirty`.

## 4. Rollback

```bash
STAGING_HOST=<user>@<ip-vm> TAG=<sha-lama> ./deploy/deploy-staging.sh deploy
```

Bisa juga lewat GitHub Actions → Run workflow dengan `tag=<sha-lama>`. Migrasi DB **tidak** di-rollback otomatis.

## 5. Reset data staging

```bash
cd /srv/omni-chan && ./dc down -v   # HAPUS semua volume: DB, NATS, storage, Grafana
```

Lalu deploy ulang.

# Tech Stack — Omni Chan

| | |
|---|---|
| Status | Kondisi per 2026-11-13 + usulan perubahan |
| Terkait | [Alur data](data-flow.md) · [Overview](overview.md) · [Roadmap](../roadmap/improvement-plan.md) |

## 1. Peta service

| Service | Repo | Bahasa | Peran | Terbuka ke internet? |
|---|---|---|---|---|
| `frontend` | submodule | TypeScript | Dashboard web (Next.js) | Ya (via nginx) |
| `backend-service` → `cmd/api` | submodule | Go | REST API v1, WebSocket, event consumer, job terjadwal | Tidak |
| `backend-service` → `cmd/broadcast-orchestrator` | submodule | Go | Penjadwal kampanye, outbox relay, recovery | Tidak |
| `backend-service` → `cmd/broadcast-worker` | submodule | Go | Kirim batch ke Meta, rate limit, retry | Tidak |
| `backend-service` → `cmd/migrate`, `cmd/enrich-ig-profiles` | submodule | Go | Utilitas | – |
| `webhook-service` | submodule | Go | Ingress webhook Meta/IG/Telegram, invoke automation publik, proxy media | **Ya, satu-satunya** |
| `storage-service` | submodule | Go | Upload & penyimpanan file | Tidak |
| `shared/` | parent repo | Go | Event envelope, NATS pub/sub, normalizer, validasi, wamid, media token | – |

## 2. Stack saat ini

### Backend (Go 1.26.11, `go.work` monorepo)
| Kebutuhan | Pilihan | Versi |
|---|---|---|
| HTTP framework | Fiber v2 (fasthttp) — v2 sudah *deprecated* (digantikan v3) | 2.52.16 |
| WebSocket | `gofiber/contrib/websocket` | 1.3 |
| DB driver | pgx v5 (SQL ditulis tangan di `repository.go`) | 5.9 |
| Migrasi | golang-migrate (`storage_schema_migrations` terpisah untuk storage) | 4.19 |
| Cache / koordinasi | go-redis v9 | 9.17 |
| Event bus | nats.go + JetStream | 1.43 |
| Auth | golang-jwt v5 (access + refresh) | 5.3 |
| Excel import | excelize | 2.11.1 |
| Object storage client | minio-go (hanya di storage-service) | 7.0 |
| Logging | `log/slog` | stdlib |
| Metrik | Prometheus client_golang (counter in-process dibaca saat scrape) | 1.25 |

### Infrastruktur
| Komponen | Versi | Dipakai untuk |
|---|---|---|
| PostgreSQL | 16 | Data utama. Tabel `messages`, `campaign_recipients`, `contact_import_rows` di-partisi `RANGE(created_at)`, tapi **hanya ada partisi DEFAULT** |
| Redis | 7 | Pub/sub fanout WebSocket, idempotensi consumer, rate limiter worker, flag pembatalan broadcast, worker registry, collision detection |
| NATS JetStream | 2.10 | Stream `WEBHOOK_EVENTS` (`meta.>`; `events.>` tidak dipakai sejak 6.4) dan stream broadcast (`broadcast.*`, MaxAge 7 hari) |
| Object storage | local disk (default) / MinIO | File media |
| Deploy | docker-compose, systemd + nginx (`deploy/`) | – |

### Frontend
| Kebutuhan | Pilihan | Versi |
|---|---|---|
| Framework | Next.js (App Router) | 15.5.28 |
| UI runtime | React | 19 |
| Server state | TanStack Query | 5 |
| Virtual list | TanStack Virtual | 3 |
| Client state | Zustand | 5 |
| Styling | Tailwind CSS + CVA + `tailwind-merge` | 3.4 |
| Primitive UI | Radix (hanya `label`, `slot`) + 3 komponen di `components/ui` | – |
| Ikon | lucide-react | – |
| Test | Vitest (unit) + Playwright (e2e, backend di-mock) | 4 / 1.64 |
| Package manager | pnpm (lockfile tunggal) | 10 |

### Kualitas & operasional
| Area | Kondisi |
|---|---|
| Unit test Go | ±150 file `_test.go`; coverage shared 57 %, backend 24 %, webhook 41 %, storage 27 % (jadi batas minimum CI) |
| Unit test FE | 30 file, 118 test; coverage lines 49 % (batas minimum 48 %) |
| Integration / e2e test | Postgres (`-tags=integration`), JetStream embedded, ingress webhook end-to-end; Playwright 9 test untuk 5 alur utama |
| CI | GitHub Actions di 5 repo (vet, golangci-lint only-new, race test, coverage gate, integration Postgres, lint, tsc, build, e2e) |
| Coverage report | Gate di CI (Go `cover`, vitest `thresholds`) |
| Monitoring / alert | `/metrics` `/health` `/ready` internal per proses; OpenTelemetry → Tempo; log JSON → Loki (Alloy); Prometheus + 8 alert; Grafana 4 dashboard (opt-in `docker-compose.observability.yml`). Error tracking (Sentry) belum |

## 3. Penilaian

| Area | Penilaian | Alasan |
|---|---|---|
| Go + Postgres + modular monolith | **Pertahankan** | Cocok untuk skala saat ini; pemisahan binary (api/orchestrator/worker) sudah tepat |
| Fiber v2 | Pertahankan dulu | fasthttp tidak kompatibel dengan `net/http`, sehingga integrasi OpenTelemetry/middleware standar lebih repot. Evaluasi ulang saat fase monitoring |
| SQL tulis tangan | Perbaiki | Rawan drift skema ↔ struct; `sqlc` di atas pgx memberi type-safety tanpa ORM |
| NATS JetStream | Masalah ada di **konfigurasi**, bukan engine | Lihat [alur data §6](data-flow.md#6-risiko-per-alur). Tim memutuskan migrasi ke Kafka (fase 8 roadmap) |
| Redis | Pertahankan | Rate limiter & koordinasi memang cocok di Redis. Idempotensi sebaiknya tidak diam-diam fallback ke memory di production |
| Frontend stack | Pertahankan, lengkapi | Belum ada design system (hanya 3 komponen UI); DTO ditulis tangan padahal ada OpenAPI |

## 4. Usulan perubahan stack

| # | Perubahan | Dari → Ke | Fase roadmap |
|---|---|---|---|
| S1 ✅ | CI pipeline | tidak ada → GitHub Actions (`go test ./...`, `go vet`, `golangci-lint`, `vitest`, `next lint`, `tsc`) | 3 |
| S2 ✅ | Integration test | Postgres via service container CI + nats-server embedded (tanpa testcontainers) | 3 |
| S3 ✅ | E2E test | Playwright untuk 5 alur utama | 3 |
| S4 | Observability | metrik JSON custom → OpenTelemetry SDK + Prometheus + Grafana + Loki/Tempo, Sentry untuk FE/BE | 4 |
| S5 ✅ | Type-safe SQL | SQL tulis tangan → `sqlc` (bertahap per modul; `make sqlc`, cek drift di CI) | 6 |
| S6 ✅ | Kontrak API FE | `docs/openapi/v1.yaml` lengkap (174/174, dijaga CI) → tipe `schema.gen.ts` (`pnpm gen:api`); modul FE pindah ke tipe ini saat disentuh | 6 |
| S7 | Design system | 3 komponen → shadcn/ui (Radix) + token dari `ui-design.mdc` + Storybook | 5 |
| S8 ✅ | Package manager FE | npm + pnpm → **pnpm saja** | 0 |
| S9 | Event bus | NATS JetStream → **Apache Kafka** (atau Redpanda, kompatibel API Kafka) lewat abstraksi `EventBus` | 8 |
| S10 | Partisi Postgres | partisi DEFAULT saja → partisi bulanan dengan `pg_partman` + tabel dedup `wamid` | 6 |

### Catatan S9 (Kafka)
- **Keuntungan:** retensi & replay panjang, ekosistem (Kafka Connect, CDC/Debezium, stream processing), ordering per partition key (misal `team_id:conversation_id`).
- **Biaya:** operasional lebih berat dari NATS (broker + KRaft, tuning partisi, consumer group rebalancing); butuh monitoring consumer lag sebelum migrasi.
- **Syarat sebelum mulai:** abstraksi `EventBus` di `shared/`, idempotensi berbasis DB, monitoring (fase 4) sudah aktif. Keputusan dicatat sebagai ADR-0008.

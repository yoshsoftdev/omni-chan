# Rencana Perbaikan — Omni Chan

| | |
|---|---|
| Status | Draft, disusun 2026-11-13 |
| Cara pakai | Kerjakan **berurutan, satu langkah per PR**. Centang `[x]` saat Definition of Done terpenuhi, di sini **dan** di [TODO.md](../TODO.md). |
| Terkait | [PRD](../product/prd.md) · [Fitur](../product/features.md) · [Tech stack](../architecture/tech-stack.md) · [Alur data](../architecture/data-flow.md) |

## Pemetaan masalah → fase

| # | Masalah yang dilaporkan | Temuan di codebase | Fase |
|---|---|---|---|
| 1 | Dokumen berantakan | 118 dokumen; patch/log/zip di root; PRD hanya 36 baris; port antar dokumen saling bertentangan | **1** (struktur sudah dirapikan) |
| 2 | List fitur berantakan | `tasks.md` tidak sinkron dengan kode; banyak halaman mock/placeholder | **2** ([features.md](../product/features.md) sudah dibuat) |
| 3 | Ganti NATS ke Kafka | Masalah saat ini ada di konfigurasi JetStream (D2–D6), bukan engine | 0 (perbaikan sementara) → **8** |
| 4 | Belum ada design fix | `components/ui` hanya punya `button`, `input`, `label`; aturan warna ada di `.cursor/rules/ui-design.mdc` tapi belum jadi token/komponen | **5** |
| 5 | UI/UX belum bagus | Komponen inbox kohesinya paling rendah (graph: 335 node, cohesion 0,015); mock & legacy masih tercampur | **7** |
| 6 | Tidak ada unit testing | Test **ada** (Go: 141 file/±426 test; FE: 30 file), tapi **tidak ada CI**, tidak ada coverage, tidak ada integration/e2e | **3** |
| 7 | Tidak ada monitoring | Hanya counter JSON; tidak ada metrik terpusat, tracing, alert, error tracking | **4** |

## Urutan & alasannya

```
0 Stabilisasi  →  1 Dokumen  →  2 Fitur  →  3 Test+CI  →  4 Monitoring
      →  5 Design system  →  6 Integritas data  →  7 UI/UX  →  8 Kafka  →  9 Fitur R3
```

- **0 paling awal** karena ada risiko kebocoran secret dan pesan pelanggan terkirim ganda.
- **Test & monitoring sebelum perubahan besar.** Tanpa keduanya, refactor UI dan migrasi Kafka tidak bisa diverifikasi.
- **Design system sebelum UI/UX.** Kalau dibalik, halaman akan didesain ulang dua kali.
- **Kafka setelah monitoring dan integritas data.** Migrasi bus butuh metrik consumer lag dan dedup berbasis DB supaya aman di-replay.
- Bila ada 2 orang: fase 5 dan 7 (frontend) bisa jalan paralel dengan fase 6 dan 8 (backend) setelah fase 4 selesai.

Ukuran: **S** ≤ 2 hari, **M** ≤ 1 minggu, **L** > 1 minggu.

---

## Fase 0 — Stabilisasi darurat (S)

- [ ] **0.1 Rotasi automation API key** yang pernah ter-commit di `campign-bug.md` (commit `e9d1e1b`). Revoke key lama di UI automation. *(Tindakan manual oleh pemilik.)*
  - Opsional: hapus dari history dengan `git filter-repo` (koordinasikan karena mengubah history semua clone).
  - DoD: key lama ditolak (401); file sudah di-redact (✅ sudah dilakukan di [bugs/campaign-broadcast-bugs.md](../bugs/campaign-broadcast-bugs.md)).
- [x] **0.2 Window recontact anti-dobel** (D1): hasil verifikasi menunjukkan kode **sudah** mengklaim lewat ledger UNIQUE. Perilakunya kini dikunci dengan test konkurensi 16 replika (`internal/inbox/window_recontact_service_test.go`).
- [x] **0.3 Hardening JetStream** (D2–D6): `shared/nats/stream.go` (`EnsureStream` yang hanya mengisi nilai kosong: MaxAge 7 hari, Duplicates 2 menit, subject digabung, konfigurasi operator dipertahankan); `Nats-Msg-Id` di publisher; `shared/nats/redelivery.go` (NAK dengan backoff eksponensial, `MaxDeliver` → `Term`, token salah dibatasi, jeda saat fetch error). Consumer broadcast: hanya backoff (recovery DB tetap pemilik kegagalan final). Diuji dengan embedded nats-server 2.10 + uji mutasi.
  - Temuan tambahan yang ikut diperbaiki: **webhook-service bisa kehilangan pesan** saat NATS mati (kunci idempotensi sudah ditandai dan Meta tetap dibalas 200). Sekarang kunci dilepas dan Meta dibalas 503 agar mengirim ulang; Redis mati → tetap publish (fail-open).
  - Variabel env baru (opsional): `NATS_STREAM_MAX_AGE`, `NATS_STREAM_MAX_BYTES`, `NATS_CONSUMER_MAX_DELIVER`, `NATS_RETRY_BASE_DELAY`, `NATS_RETRY_MAX_DELAY`.
- [x] **0.4 Tutup `/api/v1/metrics`** di webhook-service (D11): wajib header `X-Internal-Token` (constant-time); tanpa token → 404; token belum dikonfigurasi → selalu tertutup.
- [x] **0.5 Kebersihan repo:** pnpm saja (`package-lock.json` dihapus, `pnpm-lock.yaml` disinkronkan), `.gitignore` untuk `dump.rdb` dan `graphify-out/`, tabel port per environment + target `make test-race|test-frontend|test-e2e|test-all` di README.
  - Temuan tambahan: build Docker webhook-service gagal (`go.sum` tanpa `golang-jwt`), sudah diperbaiki dan diverifikasi dengan `docker build`.

- [x] **0.6 Patch dependensi rentan** (govulncheck + `pnpm audit`: 0 temuan):
  - Go toolchain **1.25 → 1.26.11** (Go 1.25 sudah EOL; belasan celah `net/http`/`crypto/tls`/`crypto/x509` hanya ditambal di 1.26). Dockerfile → `golang:1.26-alpine`; `go.work` + semua `go.mod` → `toolchain go1.26.11`; backend `go 1.26.0` (syarat excelize 2.11.1). Server yang build via `deploy/` butuh Go ≥ 1.26 atau `GOTOOLCHAIN=auto`.
  - fiber 2.52.16, brotli 1.2.7, pgx 5.9.2, excelize 2.11.1 (celah parsing file Excel unggahan), x/net 0.62, x/text 0.43; `minio-go` yang tidak terpakai terbuang oleh `go mod tidy`.
  - Frontend: next 15.5.28, postcss 8.5.31, eslint-config-next 15.5.28; `pnpm.overrides` untuk dependensi yang di-pin next (postcss, sharp 0.35.5) dan tool dev (brace-expansion, js-yaml, postcss-selector-parser, minimatch). Output lint identik sebelum/sesudah.
- [x] **0.7 Bug single-flight upload media** (ditemukan saat verifikasi): dua worker bisa meng-upload media yang sama ke Meta (cek cache → klaim tidak atomik). Sekarang setelah klaim, baris ACTIVE yang masih segar dipakai ulang (`reuseConcurrentRefresh`), dikunci test deterministik + uji mutasi.

### Fase T — Tooling & baseline (selesai sebelum fase 0)
- [x] T1 Playwright 1.64 + harness e2e (`frontend/e2e`, backend di-mock, kontrak API dikunci), 4 smoke test auth
- [x] T3 Plugin agent **ponytail** 5.3.3 terpasang (scope user)
- [x] T2 Baseline hijau: 3 paket test yang gagal build, 2 test usang, 1 peringatan vet, dan **data race nyata** di `internal/media/metrics.go` (counter dipakai konkuren oleh api & worker) diperbaiki

## Fase 1 — Dokumen (S) · *sebagian besar selesai*

- [x] 1.1 Struktur folder `docs/` baru + index [`docs/README.md`](../README.md)
- [x] 1.2 Pindahkan patch, log, dan zip dari root ke `docs/archive/`
- [x] 1.3 PRD v2 ([product/prd.md](../product/prd.md))
- [x] 1.4 Tech stack & alur data
- [ ] 1.5 Review PRD oleh product owner: jawab "Pertanyaan terbuka" dan semua tanda **[KONFIRMASI]**
- [x] 1.6 Tulis ulang [`architecture/overview.md`](../architecture/overview.md) dengan port yang benar setelah 0.5
- [x] 1.7 Aturan: setiap PR fitur wajib meng-update `features.md` dan `TODO.md` — template PR di 5 repo.

## Fase 2 — Daftar fitur & triase backlog (S) · *daftar sudah dibuat*

- [x] 2.1 Inventaris fitur dengan status ✅/🟡/🧪/⬜ ([features.md](../product/features.md))
- [ ] 2.2 Validasi status bersama PO dan tetapkan prioritas (MoSCoW) per fitur
- [ ] 2.3 Putuskan nasib halaman 🧪: `billing`, `automation/workflows`, `integrations`, edit jadwal, test invoke. Pilihan: sembunyikan di balik feature flag atau hapus.
- [ ] 2.4 Ubah tiap butir di [bugs/campaign-broadcast-bugs.md](../bugs/campaign-broadcast-bugs.md) menjadi issue tracker dengan langkah reproduksi
- [x] 2.5 Hapus kode `_legacy/` dan `*.mock.ts` yang tidak lagi di-import (24 file). Sisa: `broadcast/mock/template-legacy-adapters.ts` dan `template-mock-data.ts` masih dipakai `template-consumer.ts` & wizard automation

## Fase 3 — Testing & CI (M)

- [x] 3.1 GitHub Actions per repo/submodule: `go vet`, `golangci-lint`, `go test -race ./...`, `pnpm lint`, `tsc --noEmit`, `vitest`
- [x] 3.2 Laporan coverage (Go `-coverprofile`, Vitest `--coverage`) + baseline. Aturan: PR tidak boleh menurunkan coverage.
- [x] 3.3 Unit test domain prioritas: normalizer webhook, resolver tenant, AssignmentEngine, template variable validation, failure classifier broadcast, RBAC permission
- [x] 3.4 Integration test (Postgres via env `DATABASE_URL` + service container CI, bukan testcontainers — tanpa dependensi baru): webhook → bus → consumer → DB (alur §2), outbox → worker (alur §4), idempotensi (event dikirim 2x)
- [x] 3.5 Test FE: komponen inbox (React Testing Library) dan mapper DTO
- [x] 3.6 E2E Playwright untuk 5 alur: login, hubungkan channel (mock Meta), balas pesan di inbox, buat & kirim kampanye, buat endpoint automation
- DoD fase: CI wajib hijau sebelum merge (branch protection aktif — **tindakan manual** di GitHub, lihat TODO).
- Bug yang ditemukan & diperbaiki selama fase 3:
  - Resolver tenant mengubah error DB sementara menjadi DLQ terminal → event hilang saat Postgres gangguan. Kini hanya "channel tidak ada" yang terminal.
  - Error throughput Meta 130429/131056 (HTTP 400) dicap gagal permanen → penerima kampanye gagal padahal hanya di-throttle. `IsRateLimited` kini membaca kode error Meta.
  - `ClaimRefresh` media: klaim pertama yang bersamaan gagal `unique violation` di Postgres (tidak terlihat di mock) → kirim header gambar gagal. Kini `ON CONFLICT` + kunci ulang.
  - Wizard automation menampilkan template **mock** bila API kosong/error → endpoint terikat template palsu. Fallback dihapus.

## Fase 4 — Monitoring & observability (M)

- [x] 4.1 Standar log: JSON `slog` dengan `trace_id`, `team_id`, `event_id` di semua service
- [x] 4.2 OpenTelemetry tracing: HTTP masuk, panggilan Meta, publish/consume event (propagasi `traceparent` di header event), query DB
- [x] 4.3 Metrik Prometheus: RED per endpoint, consumer lag & redelivery, ukuran DLQ, error rate Meta per kode, throughput broadcast, koneksi WS aktif
- [ ] 4.4 Stack: Prometheus + Grafana + Loki (log) + Tempo (trace) di docker-compose/deploy ✅; **Sentry untuk FE & BE — belum, menunggu keputusan**
- [x] 4.5 Dashboard Grafana: *Ingress webhook*, *Inbox realtime*, *Broadcast*, *Infra (PG/Redis/bus)*
- [x] 4.6 Alert: DLQ > 0, consumer lag > X, error rate Meta > Y%, worker tidak heartbeat, disk bus > 80%
- [x] 4.7 Healthcheck compose (`depends_on: condition: service_healthy`) + endpoint `/health` & `/ready` di semua service
- DoD fase: satu pesan masuk bisa ditelusuri end-to-end di Tempo dari webhook sampai WS. ✅ Diverifikasi pada stack nyata: trace `POST /api/v1/webhooks/meta` → `consume meta.message.received` → query Postgres, log consumer di Loki membawa `trace_id` yang sama; span `realtime publish` menutup hop WS.
- Temuan & perbaikan selama fase 4:
  - `docker-compose.yml` tidak meneruskan `META_TOKEN_ENCRYPTION_KEY` → backend **tidak bisa start** di compose. Kini variabel wajib.
  - Migrasi storage-service mereferensikan tabel `teams` milik backend → di database baru storage gagal start. Urutan startup kini storage menunggu backend sehat.
  - Header `traceparent` dari fasthttp ber-case kanonik (`Traceparent`) → trace masuk tidak tersambung. Carrier kini case-insensitive.

## Fase 5 — Design system / "design fix" (M)

- [ ] 5.1 Audit UI: screenshot semua route (±45 halaman), catat inkonsistensi spacing, tipografi, warna, state kosong, loading, dan error
- [ ] 5.2 Design token dari `ui-design.mdc` (warna, radius, spacing, tipografi) → CSS variables + `tailwind.config.ts`
- [ ] 5.3 Pustaka komponen berbasis shadcn/ui (Radix): Button, Input, Select, Dialog, Sheet, Tabs, Table, Badge, Toast, Dropdown, Tooltip, Skeleton, EmptyState, Pagination, DataTable
- [ ] 5.4 Pola layout: AppShell, PageHeader, FilterBar, DetailPanel, Wizard
- [ ] 5.5 Storybook + visual regression (Chromatic atau Playwright screenshot)
- [ ] 5.6 Dokumen `docs/design/design-system.md` (do/don't, contoh)
- DoD fase: komponen baru hanya boleh memakai pustaka ini (ditegakkan lewat lint rule/import restriction).

## Fase 6 — Integritas data & pemecahan proses (M–L)

- [ ] 6.1 Dedup wamid di DB (D7): partisi bulanan dengan `pg_partman` + tabel `message_wamids(team_id, wamid) PRIMARY KEY`, **atau** hapus partisi (keputusan via ADR)
- [ ] 6.2 Idempotensi consumer berbasis DB (tabel `processed_events` / unique constraint); Redis jadi cache saja; production fail-fast bila Redis tidak ada (D8)
- [ ] 6.3 Pecah `cmd/api` menjadi `cmd/api` (HTTP+WS), `cmd/event-consumer`, `cmd/scheduler` (leader lock untuk recontact & reconcile)
- [ ] 6.4 Matikan subscriber legacy `events.>` (D10) mengikuti [12-6-legacy-decommission](../archive/reviews/12-6-legacy-decommission.md)
- [ ] 6.5 WS resync: klien kirim cursor terakhir saat reconnect; server kirim delta (D9)
- [ ] 6.6 `sqlc` untuk modul baru dan modul yang disentuh; generate client TS dari OpenAPI (S5, S6)

## Fase 7 — UI/UX revamp (L)

Kerjakan per modul, satu PR per modul, memakai design system fase 5:
- [x] 7.1 **Inbox** (prioritas tertinggi, dipakai harian): pecah komponen besar, keyboard shortcut, state kosong/loading/error, virtualisasi, indikator window 24 jam & SLA
- [x] 7.2 **Broadcast wizard**: template picker dengan search + pagination (bug #2), preview template, validasi media header (bug #1)
- [x] 7.3 **Kontak & segmen**: DataTable standar, bulk action, progres import
- [ ] 7.4 **Channel onboarding**: stepper jelas, status readiness
- [ ] 7.5 **Automation**: snippet, test invoke yang nyata (hapus teks "mock")
- [ ] 7.6 **Team & settings**
- [ ] 7.7 Uji usability dengan 3–5 pengguna nyata per modul + audit aksesibilitas (axe)
- DoD tiap modul: e2e test fase 3 tetap hijau, tidak ada komponen di luar design system.

## Fase 8 — Migrasi NATS → Kafka (L)

- [ ] 8.1 **ADR-0008** event bus: Kafka vs Redpanda vs tetap NATS, dengan kriteria retensi, replay, ordering, ops cost, dan kebutuhan integrasi data
- [ ] 8.2 Abstraksi `shared/eventbus` (`Publisher`, `Subscriber`, `Message{Key, Headers, Ack/Nack}`); semua kode pindah ke interface ini (masih dengan implementasi NATS)
- [ ] 8.3 Desain topic: `webhook.meta.events`, `webhook.meta.dlq`, `broadcast.jobs`, `broadcast.batches`; partition key `team_id` (atau `team_id:conversation_id` untuk urutan per percakapan); retensi; skema envelope (pertimbangkan Schema Registry)
- [ ] 8.4 Infra Kafka (KRaft) di compose & deploy + monitoring lag (kafka-exporter) di Grafana fase 4
- [ ] 8.5 Implementasi `eventbus/kafka` (franz-go atau segmentio/kafka-go) + integration test dengan testcontainers
- [ ] 8.6 Migrasi per stream: **dual-publish** → consumer Kafka shadow (hasil dibandingkan, tidak menulis) → cutover → matikan NATS untuk stream itu. Mulai dari broadcast (sudah pakai outbox), lalu webhook ingress.
- [ ] 8.7 Hapus dependensi NATS dan update dokumen (tech stack, alur data, `domain/events/*`)
- DoD fase: tidak ada event hilang selama cutover (dibuktikan dengan rekonsiliasi count per hari) dan lag terpantau.

## Fase 9 — Penyelesaian fitur R3 (L)

- [ ] 9.1 RBAC diterapkan di semua modul (TEAM-04)
- [ ] 9.2 Keyword reply & welcome message
- [ ] 9.3 Analitik inbox (volume, first response time, resolusi) + laporan agent
- [ ] 9.4 Telegram end-to-end (jika dikonfirmasi di PRD)
- [ ] 9.5 Billing (jika masuk lingkup)

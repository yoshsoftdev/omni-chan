# TODO — Omni Chan

Pelacak progres tunggal. Detail, alasan, dan Definition of Done setiap butir ada di
[roadmap/improvement-plan.md](roadmap/improvement-plan.md) (ID sama: `0.3`, `3.1`, dst.).

**Aturan update**
1. Kerjakan berurutan, satu butir per PR.
2. Centang `[x]` hanya bila DoD terpenuhi **dan** `make test-all` hijau.
3. Tulis bukti di kolom akhir: hash commit, file, atau PR.
4. Bila butir di sini berubah, ubah juga di `improvement-plan.md` dalam PR yang sama.

_Terakhir diperbarui: 2026-11-14_

## Ringkasan

| Fase | Topik | Selesai |
|---|---|---|
| T | Tooling & baseline | 3 / 3 ✅ |
| 0 | Stabilisasi darurat | 7 / 7 |
| 1 | Dokumen | 6 / 7 |
| 2 | Daftar fitur & triase | 2 / 5 |
| 3 | Testing & CI | 6 / 6 ✅ |
| 4 | Monitoring | 6 / 7 |
| 5 | Design system | 6 / 6 |
| 6 | Integritas data & pemecahan proses | 5 / 6 |
| 7 | UI/UX revamp | 0 / 7 |
| 8 | Migrasi NATS → Kafka | 0 / 7 |
| 9 | Fitur R3 | 0 / 5 |
| **Total** | | **41 / 66** |

## Perlu tindakan manual (pemilik repo)

- [x] Revoke & rotasi automation API key prefix `oc_auto_9221` (= butir 0.1) — server `nasi` beserta DB-nya sudah dihapus (2026-11-15)
- [x] Push branch: submodule dulu (`frontend@yos`, `backend-service@yos`, `webhook-service@stag`, `storage-service@chore/phase0-hardening`), lalu induk `chore/phase0-hardening`
- [ ] Server deploy (`deploy/`): pastikan Go ≥ 1.26 atau `GOTOOLCHAIN=auto` dengan akses internet
- [ ] Review PRD oleh product owner (= butir 1.5)
- [ ] GitHub: buat secret `SUBMODULES_TOKEN` (repo induk) dan `PARENT_REPO_TOKEN` (repo backend/webhook/storage) bila repo private; read-only cukup
- [ ] Compose kini **mewajibkan** `META_TOKEN_ENCRYPTION_KEY` (base64 32 byte; dulu tidak diteruskan sehingga backend tak bisa start) dan, untuk stack monitoring, `GRAFANA_ADMIN_PASSWORD`
- [ ] Alertmanager: isi receiver notifikasi (Slack/email/webhook) di `deploy/observability/alertmanager.yml`
- [ ] Putuskan Sentry (butir 4.4): pakai SaaS, self-host, atau cukup Loki + alert
- [ ] GitHub: aktifkan branch protection "require status checks" (`ci`) di kelima repo — DoD fase 3

---

## Fase T — Tooling & baseline

- [x] **T1** Playwright 1.64 + harness e2e (backend di-mock, kontrak API dikunci), 4 smoke test auth — `frontend@c296e9d`
- [x] **T2** Baseline hijau: test gagal build/usang diperbaiki, data race `media/metrics.go` — `backend@468a7b4`
- [x] **T3** Plugin **ponytail 5.3.3** terpasang (scope user; skill `/ponytail`, `/ponytail-review`, `/ponytail-audit`, `/ponytail-debt`, `/ponytail-gain`, `/ponytail-help`) — aktif di sesi baru

## Fase 0 — Stabilisasi darurat

- [x] **0.1** Rotasi automation API key — selesai dengan dihapusnya server `nasi` (2026-11-15). Key ber-prefix `oc_auto_9221` hanya berlaku di DB server itu.
  - Key **tidak pernah ada di commit yang ter-push**. Key hanya ada di 2 blob lokal tanpa referensi, dan `docs/bugs/campaign-broadcast-bugs.md` sudah di-redact.
  - **Jangan restore dump DB `nasi`** ke staging/VM baru. Kalau terlanjur, rotate key `oc_auto_9221` di sana.
  - Dev lokal kini memakai container `omni-chan-postgres` (127.0.0.1:5432, DB `omni_chan_dev`, migrasi 52).
- [x] **0.2** Window recontact anti-dobel: diverifikasi + test 16 replika — `backend@468a7b4`
- [x] **0.3** Hardening JetStream + perbaikan kehilangan pesan saat NATS mati — induk `efe1b18`, `webhook@f58848d`, `backend@468a7b4`
- [x] **0.4** `/api/v1/metrics` webhook wajib token internal — `webhook@f58848d`
- [x] **0.5** Kebersihan repo (pnpm saja, `.gitignore`, tabel port, target `make test-*`, Docker build webhook diperbaiki) — induk `efe1b18`
- [x] **0.6** Patch dependensi rentan, Go 1.26.11 (govulncheck & pnpm audit: 0) — semua repo
- [x] **0.7** Bug upload media ganda (single-flight) — `backend@468a7b4`

## Fase 1 — Dokumen

- [x] **1.1** Struktur `docs/` baru + index — induk `efe1b18`
- [x] **1.2** Patch/log/zip dipindah ke `docs/archive/` — induk `efe1b18`
- [x] **1.3** PRD v2 — `product/prd.md`
- [x] **1.4** Tech stack & alur data — `architecture/tech-stack.md`, `architecture/data-flow.md`
- [ ] **1.5** Review PRD oleh product owner (pertanyaan terbuka + tanda [KONFIRMASI])
- [x] **1.6** Tulis ulang `architecture/overview.md` (tabel service, alur utama, port → tabel README) — `architecture/overview.md`
- [x] **1.7** Template PR di 5 repo (test hijau, kontrak API, `features.md`, `TODO.md`, tanpa secret) — `.github/pull_request_template.md`

## Fase 2 — Daftar fitur & triase

- [x] **2.1** Inventaris fitur ✅/🟡/🧪/⬜ — `product/features.md`
- [ ] **2.2** Validasi status + prioritas MoSCoW bersama PO
- [ ] **2.3** Putuskan nasib halaman 🧪 (billing, workflows, integrations, edit jadwal, test invoke)
- [ ] **2.4** Bug list → issue tracker dengan langkah reproduksi
- [x] **2.5** Hapus 24 file `_legacy/`/mock yang tidak di-import (−3.823 baris); 2 file mock yang masih dipakai dipertahankan — `frontend`

## Fase 3 — Testing & CI

- [x] **3.1** GitHub Actions di 5 repo: vet, golangci-lint 2.15 (only-new), `go test -race`, lint, tsc, vitest, build, Playwright; diverifikasi actionlint — `.github/workflows/ci.yml`
- [x] **3.2** Gate coverage = baseline: Go shared 57 / backend 23 / webhook 40 / storage 26 %, FE lines 48 % (vitest thresholds)
- [x] **3.3** Domain prioritas tercakup; test baru resolver tenant + klasifikasi rate limit Meta (2 bug diperbaiki)
- [x] **3.4** Integration: Postgres asli (`-tags=integration`, job CI `integration`) untuk klaim recontact & media; ingress webhook → JetStream → subscriber backend + idempotensi (1 bug race diperbaiki)
- [x] **3.5** Mapper DTO tercakup (+ test pemetaan kanal percakapan); perilaku komponen inbox lewat e2e 3.6
- [x] **3.6** E2E 5 alur: login, hubungkan channel manual, balas di inbox, buat & kirim kampanye, buat endpoint automation (+ regresi template palsu) — `frontend/e2e/*.e2e.ts`, 9 test

## Fase 4 — Monitoring & observability

- [x] **4.1** Log JSON + `trace_id`/`span_id` (slog handler), access log per request dengan `team_id`/`request_id`, log consumer gagal dengan `event_id` — `shared/observability/log.go`, `fiber.go`
- [x] **4.2** OpenTelemetry: HTTP masuk, client Meta & storage (otelhttp), publish/consume NATS (`traceparent`), query Postgres (otelpgx), publish realtime WS — diverifikasi di Tempo: satu trace webhook → consumer → DB
- [x] **4.3** Prometheus di port internal tiap proses: RED per route, lag/redelivery/DLQ/ukuran JetStream, kegagalan broadcast per kode, throughput broadcast, cache media, koneksi WS
- [ ] **4.4** Stack Prometheus + Alertmanager + Grafana + Loki/Alloy + Tempo ✅ (`docker-compose.observability.yml`, diuji end-to-end). **Sisa: Sentry FE/BE — menunggu keputusan** (layanan pihak ketiga, data error keluar dari server)
- [x] **4.5** 4 dashboard ter-provisioning: Ingress webhook, Inbox realtime, Broadcast, Infra — `deploy/observability/grafana/dashboards`
- [x] **4.6** 8 alert rule (service down, DLQ, lag, bus tak terbaca, storage bus, publish webhook gagal, 5xx, gagal kirim broadcast) — lolos promtool; receiver notifikasi perlu diisi (manual)
- [x] **4.7** `/health` & `/ready` di port internal semua proses; healthcheck + `depends_on: service_healthy` di compose

## Fase 5 — Design system

- [x] **5.1** Audit UI semua route (46 halaman; empty/error/loading) — `design/ui-audit.md`
- [x] **5.2** Design token: token `globals.css`/Tailwind dipakai di seluruh UI (hex, palet mentah, `bg-white`, font 10–11 px → token; dark mode Window Recontact/inbox ikut benar) — `frontend` (lihat `design/ui-audit.md`)
- [x] **5.3** Pustaka komponen shadcn/ui (Radix): Dialog, Sheet, Select, Tabs, DropdownMenu, Tooltip, Skeleton, Button `destructive`; + DataTable/Pagination/StatusBadge/EmptyState/ErrorState/Toast yang ada; lint DoD (`eslint.legacy-ui.mjs` hanya menyusut) — `frontend`
- [x] **5.4** Pola layout: `RouteTabs` menyatukan 4 nav section, `DetailPanel` (Sheet), StepWizard semantik, AppShell tanpa data palsu (badge "2", role "Admin") — `frontend`
- [x] **5.5** Storybook 10 (nextjs-vite) + visual regression Playwright, 16 story × light/dark, toleransi 0 px, baseline dari image Docker Playwright; job CI `visual` — `frontend`
- [x] **5.6** Dokumen `design/design-system.md`

## Fase 6 — Integritas data & pemecahan proses

- [x] **6.1** Dedup wamid di DB: tabel klaim `message_wamids(team_id, wamid)` dalam transaksi insert (migrasi `000053`), tanpa pg_partman — [ADR 0008](adr/0008-message-wamid-dedup.md); integration test 16 insert paralel → 1 baris — `backend`
- [x] **6.2** Idempotensi consumer di Postgres (`processed_events`, migrasi `000054`) dengan Redis sebagai cache (`CachedStore`); tak ada lagi fallback memori; purge per jam > MaxAge stream (8 hari). Redis sudah fail-fast saat startup — `backend`
- [x] **6.3** Backend dipecah jadi `cmd/api` (HTTP+WS), `cmd/event-consumer`, `cmd/scheduler` di atas `internal/app.Run(role)`; job scheduler dijaga advisory lock Postgres (`pkg/leader`, test 2 replika); consumer exit bila NATS tak ada. Compose, unit systemd, skrip install, Dockerfile, Makefile (`run-all`), dan scrape Prometheus (9106/9107) ikut diperbarui — `backend`
- [x] **6.4** Subscriber legacy `events.>` dihapus: tidak ada publisher sejak 11.3. Ikut dihapus `LegacyConsumer`, pub/sub legacy di `shared/nats`, `Envelope`/`Subject`, dan webhook handler onboarding yang mati (±500 baris). Test tracing dipindah ke jalur `meta.>`. Pembersihan consumer durable `-legacy` di server: lihat `domain/events/nats-subjects.md` — `backend`, `shared`
- [x] **6.5** WS resync: saat socket tersambung lagi, klien me-refetch state yang tampil (semua query aktif, plus thread/aktivitas/catatan inbox). Ini menggantikan cursor `created_at`, yang tidak menangkap perubahan status pesan. Unit test reconnect + e2e "pesan saat offline muncul setelah reconnect" — `frontend`
- [ ] **6.6** `sqlc` + client TS dari OpenAPI
  - ✅ S5: `sqlc` aktif (`sqlc.yaml`, `make sqlc`, cek drift `sqlc diff` di CI backend); `processed_events` sudah memakai query hasil generate. Modul lain dimigrasi saat disentuh.
  - ✅ Pipeline S6: `pnpm gen:api` menghasilkan `schema.gen.ts`; CI induk gagal bila hasil generate usang.
  - ⏳ **Butuh keputusan:** spec OpenAPI baru mencakup 92/174 route (`python3 -I scripts/openapi_coverage.py --list`), dan sebagian schema usang (mis. `ReadinessResponse`). Tipe/klien hasil generate baru bisa dipakai fitur setelah spec lengkap dan akurat.

## Fase 7 — UI/UX revamp

- [ ] **7.1** Inbox
- [ ] **7.2** Broadcast wizard (template picker search + pagination, preview, validasi header media)
- [ ] **7.3** Kontak & segmen
- [ ] **7.4** Channel onboarding
- [ ] **7.5** Automation (test invoke nyata)
- [ ] **7.6** Team & settings
- [ ] **7.7** Uji usability + audit aksesibilitas

## Fase 8 — Migrasi NATS → Kafka

- [ ] **8.1** ADR-0008 event bus
- [ ] **8.2** Abstraksi `shared/eventbus`
- [ ] **8.3** Desain topic & partition key
- [ ] **8.4** Infra Kafka (KRaft) + monitoring lag
- [ ] **8.5** Implementasi `eventbus/kafka` + integration test
- [ ] **8.6** Migrasi per stream (dual-publish → shadow → cutover)
- [ ] **8.7** Hapus NATS + update dokumen

## Fase 9 — Fitur R3

- [ ] **9.1** RBAC di semua modul
- [ ] **9.2** Keyword reply & welcome message
- [ ] **9.3** Analitik inbox + laporan agent
- [ ] **9.4** Telegram end-to-end (jika dikonfirmasi)
- [ ] **9.5** Billing (jika masuk lingkup)

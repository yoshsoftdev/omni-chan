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
| 0 | Stabilisasi darurat | 6 / 7 |
| 1 | Dokumen | 6 / 7 |
| 2 | Daftar fitur & triase | 2 / 5 |
| 3 | Testing & CI | 6 / 6 ✅ |
| 4 | Monitoring | 0 / 7 |
| 5 | Design system | 0 / 6 |
| 6 | Integritas data & pemecahan proses | 0 / 6 |
| 7 | UI/UX revamp | 0 / 7 |
| 8 | Migrasi NATS → Kafka | 0 / 7 |
| 9 | Fitur R3 | 0 / 5 |
| **Total** | | **23 / 66** |

## Perlu tindakan manual (pemilik repo)

- [ ] Revoke & rotasi automation API key yang bocor di commit `e9d1e1b` (= butir 0.1)
- [ ] Push branch: submodule dulu (`frontend@yos`, `backend-service@yos`, `webhook-service@stag`, `storage-service@chore/phase0-hardening`), lalu induk `chore/phase0-hardening`
- [ ] Server deploy (`deploy/`): pastikan Go ≥ 1.26 atau `GOTOOLCHAIN=auto` dengan akses internet
- [ ] Review PRD oleh product owner (= butir 1.5)
- [ ] GitHub: buat secret `SUBMODULES_TOKEN` (repo induk) dan `PARENT_REPO_TOKEN` (repo backend/webhook/storage) bila repo private; read-only cukup
- [ ] GitHub: aktifkan branch protection "require status checks" (`ci`) di kelima repo — DoD fase 3

---

## Fase T — Tooling & baseline

- [x] **T1** Playwright 1.64 + harness e2e (backend di-mock, kontrak API dikunci), 4 smoke test auth — `frontend@c296e9d`
- [x] **T2** Baseline hijau: test gagal build/usang diperbaiki, data race `media/metrics.go` — `backend@468a7b4`
- [x] **T3** Plugin **ponytail 5.3.3** terpasang (scope user; skill `/ponytail`, `/ponytail-review`, `/ponytail-audit`, `/ponytail-debt`, `/ponytail-gain`, `/ponytail-help`) — aktif di sesi baru

## Fase 0 — Stabilisasi darurat

- [ ] **0.1** Rotasi automation API key (manual) — file sudah di-redact
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

- [ ] **4.1** Standar log JSON dengan `trace_id`, `team_id`, `event_id`
- [ ] **4.2** OpenTelemetry tracing (HTTP, Meta, publish/consume, DB)
- [ ] **4.3** Metrik Prometheus (RED, consumer lag, DLQ, error Meta, throughput broadcast, koneksi WS)
- [ ] **4.4** Stack Prometheus + Grafana + Loki + Tempo; Sentry FE/BE
- [ ] **4.5** Dashboard Grafana (ingress, inbox, broadcast, infra)
- [ ] **4.6** Alert (DLQ, lag, error Meta, heartbeat worker, disk bus)
- [ ] **4.7** Healthcheck compose + `/health` & `/ready` di semua service

## Fase 5 — Design system

- [ ] **5.1** Audit UI semua route (±45 halaman)
- [ ] **5.2** Design token dari `ui-design.mdc` → CSS variables + Tailwind
- [ ] **5.3** Pustaka komponen shadcn/ui
- [ ] **5.4** Pola layout (AppShell, PageHeader, FilterBar, DetailPanel, Wizard)
- [ ] **5.5** Storybook + visual regression
- [ ] **5.6** Dokumen `docs/design/design-system.md`

## Fase 6 — Integritas data & pemecahan proses

- [ ] **6.1** Dedup wamid di DB (pg_partman + tabel dedup, atau hapus partisi — via ADR)
- [ ] **6.2** Idempotensi consumer berbasis DB; Redis jadi cache
- [ ] **6.3** Pecah `cmd/api` → `api`, `event-consumer`, `scheduler`
- [ ] **6.4** Matikan subscriber legacy `events.>`
- [ ] **6.5** WS resync dengan cursor saat reconnect
- [ ] **6.6** `sqlc` + client TS dari OpenAPI

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

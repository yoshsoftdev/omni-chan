# TODO — Omni Chan

Pelacak progres tunggal. Detail, alasan, dan Definition of Done setiap butir ada di
[roadmap/improvement-plan.md](roadmap/improvement-plan.md) (ID sama: `0.3`, `3.1`, dst.).

**Aturan update**
1. Kerjakan berurutan, satu butir per PR.
2. Centang `[x]` hanya bila DoD terpenuhi **dan** `make test-all` hijau.
3. Tulis bukti di kolom akhir: hash commit, file, atau PR.
4. Bila butir di sini berubah, ubah juga di `improvement-plan.md` dalam PR yang sama.

_Terakhir diperbarui: 2026-11-16_

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
| 6 | Integritas data & pemecahan proses | 6 / 6 |
| 7 | UI/UX revamp | 6 / 7 |
| 8 | Migrasi NATS → Kafka | 0 / 7 |
| 9 | Fitur R3 | 0 / 5 |
| **Total** | | **48 / 66** |

## Perlu tindakan manual (pemilik repo)

- [x] Revoke & rotasi automation API key prefix `oc_auto_9221` (= butir 0.1) — server `nasi` beserta DB-nya sudah dihapus (2026-11-15)
- [x] Push branch: submodule dulu (`frontend@yos`, `backend-service@yos`, `webhook-service@stag`, `storage-service@chore/phase0-hardening`), lalu induk `chore/phase0-hardening`
- [ ] Server deploy (`deploy/`): pastikan Go ≥ 1.26 atau `GOTOOLCHAIN=auto` dengan akses internet
- [ ] Review PRD oleh product owner (= butir 1.5)
- [ ] GitHub: buat secret `SUBMODULES_TOKEN` (repo induk) dan `PARENT_REPO_TOKEN` (repo backend/webhook/storage) bila repo private; read-only cukup
- [ ] Compose kini **mewajibkan** `META_TOKEN_ENCRYPTION_KEY` (base64 32 byte; dulu tidak diteruskan sehingga backend tak bisa start) dan, untuk stack monitoring, `GRAFANA_ADMIN_PASSWORD`
- [ ] Alertmanager: isi receiver notifikasi (Slack/email/webhook) di `deploy/observability/alertmanager.yml`
- [ ] Putuskan Sentry (butir 4.4): pakai SaaS, self-host, atau cukup Loki + alert
- [ ] Jalankan uji usability fase 7 dengan 3–5 pengguna nyata per modul (= butir 7.7) — skenario di `design/usability-test-plan.md`
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
- [x] **6.6** `sqlc` + client TS dari OpenAPI
  - ✅ S5: `sqlc` aktif (`sqlc.yaml`, `make sqlc`, cek drift `sqlc diff` di CI backend); `processed_events` sudah memakai query hasil generate. Modul lain dimigrasi saat disentuh.
  - ✅ Pipeline S6: `pnpm gen:api` menghasilkan `schema.gen.ts`; CI induk gagal bila hasil generate usang.
  - ✅ Spec OpenAPI dilengkapi manual (keputusan 2026-11-15): 174/174 operasi, 140 schema dari tipe Go. Ringkasan/deskripsi lama dipertahankan, schema usang diganti. Kecocokan dengan perilaku nyata diuji pada API lokal: 51/51 respons (GET, POST create, GET by id) valid. CI induk job `openapi` gagal bila ada route tak terdokumentasi atau spec tidak valid (redocly).

## Fase 7 — UI/UX revamp

- [x] **7.1** Inbox — `frontend@284391f`
  - Komponen dipecah: `inbox-page.tsx` 755 → ±470 baris; state thread (pesan, aktivitas, catatan, kirim/retry/hapus, event realtime) pindah ke `hooks/use-conversation-thread.ts`, shortcut ke `hooks/use-inbox-shortcuts.ts`. `send-message-form.tsx` (tak dipakai) dihapus.
  - Shortcut: `j`/`k` percakapan berikut/sebelumnya (mengikuti urutan sort yang tampil), `/` cari, `r` fokus composer, `Esc` kembali; petunjuk tampil saat belum ada percakapan dipilih.
  - State: error daftar kini `ErrorState` + "Try again" di panel daftar (dulu menutupi seluruh halaman); thread punya skeleton saat memuat dan `ErrorState` + retry bila gagal (dulu tampil "No messages yet"); kosong pakai `EmptyState`.
  - Virtualisasi: daftar percakapan sudah tervirtualisasi (`@tanstack/react-virtual`); thread tidak, karena dibatasi 100 pesan per muat.
  - Indikator: chip "24h window · 5h left" / "24h window closed" di header (kuning < 2 jam), di samping chip waktu tunggu (SLA) yang sudah ada.
  - Layout: kolom konteks kanan kini baru inline mulai `2xl` (1536 px); di 1280 px dulu thread hanya ±125 px lebar. Di bawah 2xl konteks dibuka sebagai drawer.
  - DoD design system: 5 file inbox keluar dari `eslint.legacy-ui.mjs` (select → `FilterSelect`/`DropdownMenu`, overlay buatan sendiri → `Dialog`, popover quick reply tanpa overlay layar penuh).
  - Test: unit `adjacentId`, `formatWindowRemaining`; e2e baru (j/k, error thread + retry, chip 24h, `r`, dialog Assign, menu More). `pnpm lint`, `typecheck`, `test:coverage`, `build`, `test:e2e` (11/11) hijau.
- [x] **7.2** Broadcast wizard — `frontend@04cb487`, `backend@c07e01f`
  - Bug #1: template hasil sync Meta hanya punya `header_handle` (tidak bisa dipakai kirim). Wizard kini bisa mengunggah file header (dikirim sebagai `header.storage_object_id`, diresolusi backend lewat `EnsureMedia`, mengalahkan file tersimpan di template) atau URL publik. Validasi di `utils/header-media.ts` (+ unit test, + test backend).
  - Header selalu ikut terkirim; centang "Show in preview" dulu ikut membuang header dari payload sehingga kirim gagal di Meta.
  - Bug #2: pencarian template kini debounce 300 ms ke server (pagination sudah ada); error memuat template tampil `ErrorState` + retry (dulu selalu "sesi kedaluwarsa").
  - Preview tampil juga di layar sempit; Review menampilkan sumber media header.
  - Keluar dari `eslint.legacy-ui.mjs` (DataTable, FilterSelect); label form terhubung ke input.
  - e2e baru: template gambar hasil sync terblokir sampai file diunggah, lalu dibuat dengan `storage_object_id`. Frontend lint/typecheck/128 unit/build/12 e2e hijau; backend `go test -race ./...` hijau.
- [x] **7.3** Kontak & segmen — `frontend@ff96b2c`, `backend@6d7c686`
  - Import kontak ternyata tidak berfungsi: endpoint `/contacts/import/{csv,excel}` mengembalikan 501 dan tidak terdaftar di v1; langkah "Map Columns" tidak pernah dikirim; `features.md` salah menandai ✅. Kini: import sinkron (≤ 5.000 baris / 10 MB) dengan pemetaan kolom, upsert per nomor telepon, laporan baris yang dilewati per nomor baris. Wizard menampilkan progres dan laporan hasil. Integration test di Postgres asli.
  - Bulk action: pilih kontak → tambah tag / hapus, kegagalan dihitung dan dilaporkan.
  - Filter sumber & kota kini di server (dulu hanya menyaring halaman yang tampil sehingga total & paginasi salah); filter Source dapat opsi "All Sources".
  - Edit segmen & aturan segmen kini tersimpan lewat `PATCH /segments/:id` (dulu hanya lokal, "API update pending").
  - `ConfirmDialog` menggantikan `window.confirm`; Dialog kini di atas drawer (z 65).
  - Semua file kontak keluar dari `eslint.legacy-ui.mjs`. Kode mati dihapus (cluster `contacts-page`, panel metadata, segment picker, placeholder analitik, tombol "More actions" palsu).
  - Spec OpenAPI: 2 operasi import + parameter `source`/`city` (176/176, redocly valid). Frontend lint/typecheck/132 unit/build/14 e2e/32 visual hijau; backend `go test -race ./...` + integration hijau.
- [x] **7.4** Channel onboarding — `frontend@3adf8de`
  - Readiness kini daftar periksa per syarat (terhubung, token, nomor, verifikasi bisnis, pembayaran, template) beserta cara memperbaiki; dulu hanya satu pesan dan mengabaikan cek token/template yang gagal (`utils/channel-readiness.ts` + test).
  - Dialog Add Channel: `Dialog` + `StepWizard`, Provider → Connect → Done. Langkah "Channel Info" dihapus karena nama & catatan yang diisi tidak pernah dipakai; Messenger tampil "Coming soon" dan tidak bisa dipilih.
  - Hapus channel lewat `ConfirmDialog` + pesan gagal; halaman pakai skeleton & `ErrorState` + retry.
  - `add-channel-modal` keluar dari `eslint.legacy-ui.mjs`; cluster `DevicesPage` yang tidak dirutekan dihapus (8 file).
  - e2e baru (checklist menyebut cek yang gagal; Messenger nonaktif). Lint/typecheck/135 unit/build/15 e2e hijau.
- [x] **7.5** Automation — `frontend@1e8be64`
  - Test invoke memang nyata (invoker produksi, pesan WhatsApp sungguhan terkirim) tetapi UI menyebut "mock executor". Kini teks jujur, wajib konfirmasi sebelum kirim, pemilih API key yang tidak pernah dikirim (dan memblokir tim tanpa key) dihapus, latensi diukur, respons tidak hilang oleh refresh setelah kirim.
  - Snippet kode: payload di-escape per bahasa (shell, PHP, Go, Java) agar nilai berisi kutip/backslash tetap valid (+ test).
  - `useConfirm` (ConfirmDialog yang bisa di-`await`) menggantikan `window.confirm`; rotate/hapus key melaporkan kegagalan.
  - Semua file automation keluar dari `eslint.legacy-ui.mjs`; dialog secret sekali-tampil tidak bisa tertutup tak sengaja.
  - e2e baru (kirim uji butuh konfirmasi, respons tampil). Lint/typecheck/136 unit/build/16 e2e/32 visual hijau.
- [x] **7.6** Team & settings — `frontend@7980366`
  - Semua file team & settings keluar dari `eslint.legacy-ui.mjs`: 12 overlay buatan sendiri → `Dialog` (editor anggota → `Sheet`), tabel → `DataTable`, select → `FilterSelect` (kini mendukung `disabled`), popover aksi baris → `DropdownMenu`. Input, select role, dan tombol izin kini punya nama aksesibel.
  - Hapus anggota & hapus quick reply lewat `useConfirm`; toast halaman users dipindah ke lapisan toast (z 80) agar tidak tertutup dialog.
  - Cluster `TeamPage` yang tidak dirutekan dihapus (3 file).
  - Lanjutan 7.2 (`frontend@9792ec8`): sisa 6 file broadcast juga dimigrasi; **`eslint.legacy-ui.mjs` kini kosong** (DoD fase 7 "tidak ada komponen di luar design system").
  - e2e baru (undang lewat dialog; hapus menunggu konfirmasi). Lint/typecheck/136 unit/build/17 e2e/32 visual hijau.
- [ ] **7.7** Uji usability + audit aksesibilitas — **sebagian**
  - ✅ Audit axe (WCAG 2.1 A/AA) di 9 halaman utama, mode terang & gelap, jalan di CI (`frontend/e2e/a11y.e2e.ts`, 18 test). Temuan diperbaiki: token `--muted-foreground` 46% → 44% (kontras 4,4 → 4,75:1 di atas `--muted`), tab nonaktif tidak lagi di bawah AA, select ukuran halaman diberi label, stepper kampanye memakai `StepWizard` — `frontend@4bc9853`.
  - ⏳ Uji usability 3–5 pengguna nyata per modul: **perlu dijalankan pemilik produk**. Skenario, ambang lulus, dan tabel hasil: [design/usability-test-plan.md](design/usability-test-plan.md).

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

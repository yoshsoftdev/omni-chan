# Audit UI (butir 5.1)

Tanggal: 2026-11-15 · Frontend `yos@24bd392` · Desktop 1280×720, tema light.

## Metode

Skrip Playwright sementara (pakai harness `e2e/fixtures/api-mock.ts`) membuka 46 route dalam keadaan login, dengan tiga kondisi backend:

| Kondisi | Respons mock |
|---|---|
| **empty** | semua `GET` → `{ items: [], total: 0 }` |
| **error** | semua call → `500 Internal server error` |
| **loading** | call tidak pernah dijawab; screenshot setelah 1,5 s |

Untuk tiap halaman dicatat screenshot, `h1`, padding `main`, dan *computed style* (ukuran/berat font, warna teks, background, radius). Ditambah scan statis kelas Tailwind di `src/`.

Batasan:
- Belum mencakup mobile/tablet, dark mode, dan state berisi data.
- `/login` dan `/register` ter-redirect ke `/dashboard` karena sesi aktif, jadi halaman auth belum ter-audit.
- Crash di kondisi **empty** sebagian dipicu bentuk mock generik (objek, padahal halaman mengharapkan array). Temuannya tetap valid: satu respons tak terduga menjatuhkan seluruh halaman.

## Status perbaikan

**P1 diperbaiki (2026-11-15, frontend belum di-commit):**
- `ErrorState` (dengan tombol "Try again") dipasang di semua halaman list/detail yang terdampak.
- Repository `getById` campaign, report, segment, dan endpoint hanya mengembalikan "not found" untuk 404 (`isNotFound`); error lain diteruskan ke halaman.
- `/broadcast/reports` tidak lagi macet di loading.
- Pemilih channel di create-template tidak lagi menampilkan "tidak ada channel" selama loading.
- Window Recontact tidak lagi menampilkan form berisi default saat load gagal, sehingga setting asli tidak tertimpa.
- `error.tsx` ditambahkan di root dan `(dashboard)`, sehingga sidebar tetap tampil saat halaman crash.
- `formatApiErrorMessage` menyembunyikan error runtime (`TypeError` dan sejenisnya), dan 77 tempat yang sebelumnya menampilkan `err.message` mentah kini memakai fungsi ini.

Diverifikasi dengan audit ulang mode error/loading/404 di 32 route, ditambah `tsc`, `eslint` (tanpa warning baru), vitest 118/118, dan e2e 9/9.

**Sisa (belum P1):**
- Team Users/Workspace hanya menampilkan toast saat load gagal.
- Wizard Create Campaign dan Inbox masih punya promise rejection yang tidak ditangani (error tetap tampil di Inbox).
- `getById` di repository template dan channel masih menelan semua error.

## Temuan utama (urut prioritas)

### P1 — Error disamarkan sebagai kosong / tidak ditemukan
Pada respons 500, **11 halaman list menampilkan empty state** ("No campaigns found", "No API keys yet", "Tidak ada template ditemukan", …), dan **5 halaman detail menampilkan "not found"**. User mengira datanya hilang, padahal server gagal. Hanya 9 halaman yang menampilkan pesan error.
- `/broadcast/reports` macet di "Loading reports..." saat error.
- `/broadcast/templates/create` menampilkan "Tidak ada channel WhatsApp terhubung" saat channel **masih loading**. Hal serupa terjadi di `/broadcast/campaigns`, `/broadcast/schedules`, `/team/activity`, `/team/roles`.

### P1 — Tidak ada error boundary
Tidak ada satu pun `error.tsx`, `global-error.tsx`, `loading.tsx`, atau `not-found.tsx` di `src/app`. Satu error render (mis. `keys.map is not a function` di `/automation/keys`) menjatuhkan seluruh halaman termasuk navigasi. Crash juga terjadi di `/automation/endpoints/[id]`, `/automation/webhooks`, `/broadcast/reports/[id]`, dan ada error JS di `/automation` dan `/automation/endpoints`.

### P1 — Pesan teknis tampil ke user
`/dashboard` menampilkan `data.checks is not iterable` mentah di banner error.

### P2 — Data palsu hardcoded
`/team/workspace` menampilkan "Users 6 / 20 (30%)" dari literal di `team-workspace-page.tsx:211`. Angka ini muncul bahkan saat API error.

### P2 — Warna di luar token
Token shadcn (`--primary`, `--muted`, …) sudah ada di `tailwind.config.ts`, tapi sering dilewati:
- **29 file** memakai hex langsung: `text-[#6B7280]` ×65, `border-[#E5E7EB]` ×41, `bg-[#F6F7F9]` ×28, `text-[#111111]` ×19, `#0F9D8A` ×6.
- **69 kelas palet mentah** berbeda (`red-600`, `emerald-600`, `slate-…`).
- Terukur di browser ada 20 warna teks dan 30 background. Contoh hampir-sama: hitam `rgb(18,18,18)` vs `rgb(17,17,17)`; merah `220,40,40` / `220,38,38` / `185,28,28`; hijau `5,150,105` vs `15,157,138`.

### P2 — Tipografi
- 16 kombinasi ukuran/berat font. Skala yang dipakai: 10, 11, 12, 14, 16, 18, 20, 24 px.
- **Teks 10–11 px dipakai 81 kali** (`text-[10px]` ×55, `text-[11px]` ×26). Ini di bawah batas keterbacaan, terutama untuk teks abu-abu `#6B7280`.
- Judul halaman konsisten (24 px/600). Ini baseline yang baik untuk dipertahankan.

### P2 — Judul halaman tidak mencerminkan halaman
- Semua sub-halaman Contacts (`/import`, `/segments`, `/tags`, detail) ber-`h1` **"Contacts"**.
- `/settings/quick-replies` dan `/settings/window-recontact` ber-`h1` **"Settings"**; `/team/users` dan `/team/workspace` ber-`h1` **"Team Management"**.
- `/inbox`, `/settings`, `/team` tidak punya `h1`; `/broadcast/campaigns/[id]` punya `h1` kosong.
- `/onboarding` ber-`h1` nama tim ("E2E Team").

### P2 — Bahasa campur
Mayoritas UI berbahasa Inggris, tetapi beberapa halaman berbahasa Indonesia: `/broadcast/templates` (+ create), `/automation/webhooks`, `/settings/appearance`, `/devices/onboarding/[id]`. Di `/broadcast/templates` bahasanya bahkan campur ("Showing 0 to 0 of 0 templates", "Previous/Next").

### P3 — Navigasi & route ganda
- Sidebar menampilkan *Settings, Quick Replies, Window Recontact, Integrations, Billing* sejajar. Quick Replies dan Window Recontact juga sudah jadi tab di dalam Settings.
- Route alias yang hanya me-redirect: `/broadcast` → `/campaigns`, `/broadcast/campaign` → `/campaigns/create`, `/broadcast/schedule` → `/schedules`, `/integrations` → `/devices`. Menu "Integrations" membuka halaman Channels.
- Nama tidak konsisten: menu "Channels" vs route `/devices` vs teks "Hubungkan WABA di halaman Devices".

### P3 — Komponen & state belum seragam
- `src/components/ui` baru berisi `button`, `input`, `label`. Select memakai `<select>` native, tabel dan pagination ditulis ulang per halaman.
- Empty state ditulis manual di 25 file; komponen `EmptyState` hanya dipakai di 10.
- Loading hanya berupa teks ("Loading X…", "Memuat template..."), tanpa skeleton. Beberapa halaman tidak punya indikator sama sekali (Inbox, Settings, Team users/workspace).
- Radius: 6 px dominan, tapi 4/8/12 px juga dipakai.

### Baik / pertahankan
- Padding konten (`main` 24 px) dan header halaman (judul + subjudul + aksi di kanan) sudah konsisten di semua halaman dashboard.
- Pola tab sub-navigasi (Contacts, Team, Settings, Automation) sudah seragam.

## Matriks per route

Keterangan:
- *kosong (salah)*: error atau loading ditampilkan sebagai data kosong.
- *tak ada sinyal*: tidak ada teks error/loading yang terdeteksi (untuk halaman statis tanpa fetch, ini wajar).
- *→ X*: halaman me-redirect ke X.

| Route | h1 | Empty | Error 500 | Loading |
|---|---|---|---|---|
| `/automation` | Automation | error JS | tak ada sinyal | tak ada sinyal |
| `/automation/endpoints` | API Endpoints | error JS | tak ada sinyal | tak ada sinyal |
| `/automation/endpoints/new` | Create Endpoint | ok | tak ada sinyal | tak ada sinyal |
| `/automation/endpoints/[id]` | — | crash | tak ada sinyal | tak ada sinyal |
| `/automation/keys` | — | crash | kosong (salah) | tak ada sinyal |
| `/automation/webhooks` | — | crash | tak ada sinyal | tak ada sinyal |
| `/automation/workflows` | Workflows | ok | — | — |
| `/billing` | Billing | ok | — | — |
| `/broadcast` | Campaigns | → /broadcast/campaigns | | |
| `/broadcast/campaign` | Create Campaign | → /broadcast/campaigns/create | | |
| `/broadcast/campaigns` | Campaigns | ok | kosong (salah) | kosong (salah) |
| `/broadcast/campaigns/create` | Create Campaign | ok | tak ada sinyal | tak ada sinyal |
| `/broadcast/campaigns/[id]` | (kosong) | ok | "not found" (salah) | teks loading |
| `/broadcast/reports` | Reports | ok | loading terus | teks loading |
| `/broadcast/reports/[id]` | — | crash | "not found" (salah) | teks loading |
| `/broadcast/schedule` | Schedules | → /broadcast/schedules | | |
| `/broadcast/schedules` | Schedules | ok | kosong (salah) | kosong (salah) |
| `/broadcast/templates` | Template WhatsApp | ok | kosong (salah) | teks loading |
| `/broadcast/templates/create` | Buat Template WhatsApp | ok | kosong (salah) | kosong (salah) |
| `/contacts` | Contacts | ok | kosong (salah) | teks loading |
| `/contacts/import` | Contacts | ok | — | — |
| `/contacts/segments` | Contacts | ok | kosong (salah) | teks loading |
| `/contacts/segments/[id]` | Contacts | ok | "not found" (salah) | teks loading |
| `/contacts/tags` | Contacts | ok | kosong (salah) | teks loading |
| `/contacts/tags/[id]` | Contacts | ok | "not found" (salah) | teks loading |
| `/dashboard` | Dashboard | ok | error tampil | teks loading |
| `/devices` | Channels | ok | error tampil | teks loading |
| `/devices/embedded-signup` | About Embedded Signup | ok | error tampil | tak ada sinyal |
| `/devices/instagram/callback` | — | → /devices | | teks loading |
| `/devices/onboarding/[id]` | — | → /devices | error tampil | teks loading |
| `/inbox` | — | ok | error tampil | tak ada sinyal |
| `/integrations` | Channels | → /devices | | |
| `/invite/[id]` | Join | ok | error tampil | teks loading |
| `/login`, `/register` | — | → /dashboard (sesi aktif, belum diaudit) | | |
| `/onboarding` | E2E Team | ok | error tampil | teks loading |
| `/settings` | — | ok | → /settings/quick-replies | tak ada sinyal |
| `/settings/appearance` | Appearance | ok | — | — |
| `/settings/quick-replies` | Settings | ok | kosong (salah) | teks loading |
| `/settings/window-recontact` | Settings | ok | tak ada sinyal | tak ada sinyal |
| `/team` | — | ok | → /team/users | tak ada sinyal |
| `/team/activity` | Workspace Activity | ok | kosong (salah) | kosong (salah) |
| `/team/permissions` | Permission Matrix | ok | error tampil | teks loading |
| `/team/roles` | Roles | ok | kosong (salah) | kosong (salah) |
| `/team/users` | Team Management | ok | tak ada sinyal | tak ada sinyal |
| `/team/workspace` | Team Management | ok | tak ada sinyal (data palsu) | tak ada sinyal |

## Masukan untuk fase berikutnya

- **5.2 Token:** satukan warna ke token yang sudah ada. Ganti `#6B7280` → `muted-foreground`, `#E5E7EB` → `border`, `#F6F7F9` → `muted`/`background`, `#111111` → `foreground`/`primary`, `#0F9D8A` → token `success`/brand. Tetapkan skala font minimum 12 px dan satu radius dasar (6 px). Catatan: `.cursor/rules/ui-design.mdc` yang dirujuk 5.2 sudah di-revert (`26062bf`). Ambil isinya dari `git show e9d1e1b:.cursor/rules/ui-design.mdc`.
- **5.3 Komponen:** prioritaskan `Select`, `Table`/`DataTable`, `Pagination`, `Skeleton`, `EmptyState`, dan **`ErrorState`** dengan tombol "Coba lagi".
- **5.4 Layout:** `PageHeader` wajib menerima judul sub-halaman (perbaiki `h1` ganda), lalu tambahkan `error.tsx`/`loading.tsx` per segmen `(dashboard)`.
- **Bug (masuk 2.4):** error→empty, `/broadcast/reports` macet loading, data palsu di workspace, pesan teknis di dashboard, crash halaman automation saat respons tak terduga.
- **Keputusan produk:** satu bahasa UI (atau i18n), dan struktur sidebar Settings/Integrations.

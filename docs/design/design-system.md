# Design system Omni Chan

Sumber kebenaran untuk UI frontend: aturan warna, token, komponen, pola layout, dan cara menegakkannya. Aturan warna di dokumen ini menggantikan `.cursor/rules/ui-design.mdc`, yang sudah tidak ada di repo (di-revert di `26062bf`).

- Katalog visual: `pnpm storybook` (port 6006) di `frontend/`.
- Hasil audit dan status perbaikan: [ui-audit.md](ui-audit.md).

## 1. Prinsip

1. **Warna hanya untuk makna/status**, bukan dekorasi. Latar selalu netral (putih/abu).
2. **Kontras minimal 4.5:1** (WCAG AA); teks besar/tebal 3:1. Berlaku di light **dan** dark.
3. **Light mode adalah default.** Dark mode opsional, diaktifkan lewat Settings → Appearance.
4. **Status ditampilkan lewat badge, ikon, atau teks**, tidak dengan mewarnai seluruh kartu.
5. **Teks minimal 12 px** (`text-xs`).
6. **Gagal ≠ kosong.** Load yang gagal selalu tampil sebagai error dengan tombol "Try again", bukan "No data" atau "not found".

## 2. Token

Didefinisikan di `src/app/globals.css` (HSL, light + `.dark`) dan dipetakan di `tailwind.config.ts`. **Pakai nama token, jangan hex atau palet Tailwind mentah.**

| Token (kelas) | Light | Untuk |
|---|---|---|
| `foreground`, `primary` | `#111111` | teks utama, tombol primary |
| `card` | `#FFFFFF` | kartu, input, panel |
| `background` | `#F6F7F9` | latar halaman |
| `muted` | `#F3F4F6` | isian halus, hover baris |
| `muted-foreground` | `#666D7A` | teks sekunder, disabled (≥ 4.5:1 di atas `muted`, cek axe 7.7) |
| `border`, `input` | `#E5E7EB` | garis, outline input |
| `success` (+`-foreground`, `-muted`, `-border`) | `#16A34A` | status positif |
| `warning` (…) | `#F59E0B` | perlu perhatian |
| `error` (…) | `#DC2626` | gagal / kritis |
| `info` (…) | `#2563EB` | callout informasi saja |

Pola status yang dipakai: **latar tipis + teks gelap**, yaitu `bg-{status}-muted border-{status}-border text-{status}-foreground`. Jangan pakai teks putih di atas warna status solid, kecuali pada tombol.

Radius dasar `--radius` = 8 px (`rounded-lg`), turunannya `rounded-md` 6 px dan `rounded-sm` 4 px. Skala font memakai skala Tailwind, dengan minimum `text-xs`. Judul halaman: `text-2xl font-semibold`.

**Do**
```tsx
<p className="text-sm text-muted-foreground">…</p>
<div className="rounded-lg border bg-card p-4">…</div>
```
**Don't**
```tsx
<p className="text-[11px] text-[#6B7280]">…</p>          // hex + teks terlalu kecil
<div className="bg-white">…</div>                        // tidak ikut dark mode
<span className="bg-emerald-600 text-white">Active</span> // palet mentah, status solid
```

## 3. Komponen

Kode fitur menyusun UI dari tiga folder ini.

| Kebutuhan | Pakai | Lokasi |
|---|---|---|
| Tombol (primary/outline/ghost/secondary/**destructive**) | `Button` | `ui/button` |
| Input teks, label | `Input`, `Label` | `ui/input`, `ui/label` |
| Pilihan dalam form | `Select` + `SelectTrigger/Content/Item` | `ui/select` |
| Filter di atas tabel | `FilterBar` + `FilterSelect` | `shared/filter-bar` |
| Modal / konfirmasi | `Dialog` + `DialogContent/Header/Title/Description/Footer` | `ui/dialog` |
| Panel samping satu record | `DetailPanel` (di atas `Sheet`) | `layout/detail-panel` |
| Tab di dalam halaman | `Tabs` + `TabsList/Trigger/Content` | `ui/tabs` |
| Tab antar-route section | `RouteTabs` | `layout/route-tabs` |
| Menu aksi (⋯) | `DropdownMenu` + `…Item` (`destructive`) | `ui/dropdown-menu` |
| Petunjuk singkat | `Tooltip` | `ui/tooltip` |
| Tabel | `DataTable` + `DataTableHead/Body/Row/HeaderCell/Cell/Empty` | `shared/data-table` |
| Paginasi | `PaginationFooter` (`start` dimulai dari 1) | `shared/pagination-footer` |
| Status | `StatusBadge` (`tone` atau `status`) | `shared/status-badge` |
| Pesan di dalam kartu | `Alert` (`info/success/warning/error`) | `shared/alert` |
| Loading | `Skeleton` (bentuk mengikuti konten) | `ui/skeleton` |
| Kosong / gagal load | `EmptyState` / `ErrorState` (+`onRetry`) | `shared/empty-state` |
| Notifikasi sesaat | `useToast()` → `notify()` + `<Toast />` | `hooks/use-toast` |
| Langkah wizard | `StepWizard` | `shared/step-wizard` |

Aksesibilitas yang sudah disediakan komponen: focus trap, Escape, dan pelabelan dari Radix (`Dialog`, `Sheet`, `Select`, `DropdownMenu`, `Tooltip`); `aria-current` di `RouteTabs` dan `StepWizard`; `role="alert"` di `ErrorState`; serta `role="status"` di toast.

## 4. Pola layout

- **AppShell** (`layout/dashboard-shell`): sidebar, header, dan `main` ber-padding 24 px. Halaman tidak menambah padding luar sendiri.
- **Halaman section:**
  ```tsx
  <div className="space-y-6">
    <PageHeader title="Contacts" description="…" right={<Button>…</Button>} />
    <RouteTabs tabs={TABS} />
    {/* konten */}
  </div>
  ```
  `h1` (lewat `PageHeader`) adalah nama section. Sub-halaman memakai `h2`.
- **List:** `PageHeader` → kartu `rounded-lg border bg-card p-4` berisi `FilterBar` → `DataTable` → `PaginationFooter`.
- **Detail di tempat:** `DetailPanel`. Konfirmasi di dalamnya memakai `Dialog`. Overlay buatan sendiri tidak bisa diklik saat panel modal sedang terbuka.
- **Wizard:** `StepWizard` di atas, satu kartu per langkah, tombol Back/Next di kanan bawah.
- **Error render:** `app/error.tsx` dan `app/(dashboard)/error.tsx` menangkap crash. Sidebar tetap tampil.

## 5. State data

Setiap halaman yang memuat data harus menangani keempat state berikut:

| State | Tampilkan |
|---|---|
| loading | `Skeleton` (atau `DataTableEmpty message="Loading …"`) |
| gagal | `ErrorState message={formatApiErrorMessage(err)} onRetry={…}` |
| 404 | "… not found", hanya bila `isNotFound(err)` |
| kosong | `EmptyState` / `DataTableEmpty` dengan langkah berikutnya |

Pesan error selalu lewat `formatApiErrorMessage`, karena fungsi ini menyembunyikan error runtime seperti `TypeError`. Repository mengembalikan `undefined` hanya untuk 404; error lain dilempar ke atas.

## 6. Penegakan

- **ESLint** (`frontend/eslint.config.mjs`) menolak kode berikut di luar `src/components/`:
  - import `@radix-ui/*` langsung;
  - elemen `<select>` dan `<table>` mentah;
  - kelas `fixed inset-0` (overlay buatan sendiri).
- File lama yang masih melanggar didaftar di `frontend/eslint.legacy-ui.mjs`. **Daftar ini hanya boleh menyusut:** hapus entri saat file dimigrasi (fase 7), dan jangan menambah entri baru.
- **Visual regression:** setiap story Storybook difoto dalam light dan dark, lalu dibandingkan dengan baseline di `frontend/visual/__screenshots__` dengan toleransi 0 piksel.
  - Jalankan `pnpm build-storybook && pnpm test:visual`. Di CI, job `visual` menjalankan ini.
  - Setelah perubahan visual yang disengaja, jalankan `pnpm test:visual:update` lalu commit PNG-nya. Perintah ini memakai Docker image Playwright yang sama dengan CI, supaya rendering font identik.
- Komponen baru di `src/components/**` wajib punya story (`*.stories.tsx`). Story baru otomatis ikut visual test.

## 7. Masih terbuka

- Migrasi file di `eslint.legacy-ui.mjs`: 21 modal buatan sendiri, `<select>`, dan `<table>` mentah (fase 7).
- `RightDrawer` lama belum memakai `Sheet`, karena `ContactDetailDrawer` merender modal sebagai sibling. Konversi modal itu ke `Dialog` dulu, baru ganti drawer-nya dengan `DetailPanel`.
- `FilterSelect` masih `<select>` native, karena beberapa filter memakai nilai `""` yang tidak didukung Radix Select.
- Warna violet/pink di role badge, audience cell, devices, dan team workspace menunggu keputusan produk.
- Satu bahasa UI (Inggris vs Indonesia) atau i18n menunggu keputusan produk.

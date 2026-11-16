# Daftar Fitur — Omni Chan

Satu-satunya sumber kebenaran untuk status fitur. Dokumen ini menggantikan [`archive/tasks-legacy.md`](../archive/tasks-legacy.md).
Status disusun dari kode per 2026-11-13 (route frontend + modul `services/backend-service/internal/*`).

**Legenda status**

| Simbol | Arti |
|---|---|
| ✅ | Jalan end-to-end (FE + BE) |
| 🟡 | Jalan, tapi ada bug terbuka atau celah yang diketahui |
| 🧪 | UI ada, tapi data mock/placeholder atau backend belum ada |
| ⬜ | Belum dibuat |

Cara update: ubah status di tabel ini dalam PR yang sama dengan perubahan kodenya.

---

## F-AUTH — Auth & Workspace
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| AUTH-01 | Login / register / logout | ✅ | `(auth)/login`, `(auth)/register` | `internal/auth`, `internal/user` |
| AUTH-02 | Refresh token | ✅ | `services/api` | `pkg/jwt`, `internal/user` |
| AUTH-03 | Undangan anggota | ✅ | `invite/[id]` | `internal/team` |
| AUTH-04 | Workspace readiness / onboarding checklist | ✅ | `dashboard`, `onboarding` | `internal/workspace`, `internal/onboarding` |

## F-TEAM — Team & RBAC
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| TEAM-01 | Kelola user tim | ✅ | `team/users` | `internal/team` |
| TEAM-02 | Role & permission matrix | ✅ | `team/roles`, `team/permissions` | `internal/rbac` |
| TEAM-03 | Activity log | ✅ | `team/activity` | `internal/team` |
| TEAM-04 | Permission dipakai di **semua** modul | ✅ | semua modul (aksi disembunyikan per permission) | Service + `/auth/me` `permissions` — lihat [rbac-audit](../security/rbac-audit.md) |

## F-CHAN — Channel / Device
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| CHAN-01 | WhatsApp Embedded Signup | ✅ | `devices/embedded-signup`, `devices/onboarding/[id]` | `internal/meta` |
| CHAN-02 | Channel health & readiness | ✅ | `devices` | `internal/meta`, `internal/channel` |
| CHAN-03 | Coexistence (WA Business app) + import riwayat chat | 🟡 | `devices` | `internal/channelsync`, `eventconsumer` (coex) |
| CHAN-04 | Instagram Business Login | ✅ | `devices/instagram/callback` | `internal/instagram`, `pkg/meta/instagram_oauth.go` |
| CHAN-05 | Telegram | 🧪 | – | Baru normalizer di `shared/normalizer/telegram.go` |
| CHAN-06 | Channel management (rename, disconnect) | ✅ | `devices` | `internal/channelmgmt` |

## F-INBOX — Team Inbox
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| INBOX-01 | Daftar percakapan realtime (WebSocket) | ✅ | `inbox` | `internal/inbox`, `internal/websocket` |
| INBOX-02 | Balas teks / media / template | ✅ | `inbox` | `internal/message` (outbound) |
| INBOX-03 | Assignment otomatis (least-loaded) & manual | ✅ | `inbox` | `internal/inbox` (AssignmentEngine) |
| INBOX-04 | Status & lifecycle percakapan | ✅ | `inbox` | `internal/inbox` |
| INBOX-05 | SLA tracking | ✅ | `inbox` | `internal/inbox` |
| INBOX-06 | Collision detection & presence agent | ✅ | `inbox` | `internal/inbox` (Redis) |
| INBOX-07 | Catatan & tag percakapan | ✅ | `inbox` | `internal/inbox`, `internal/tag` |
| INBOX-08 | Quick replies | ✅ | `settings/quick-replies` | `internal/quickreply` |
| INBOX-09 | Window recontact otomatis | ✅ | `settings/window-recontact` | `internal/inbox` (klaim ledger mencegah kirim ganda antar replika) |
| INBOX-10 | Media player (audio/video/gambar terautentikasi) | ✅ | `features/inbox` | `internal/media`, storage-service |

## F-CONT — Kontak
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| CONT-01 | CRUD kontak + consent | ✅ | `contacts` | `internal/contact` |
| CONT-02 | Import CSV / Excel (sinkron, ≤ 5.000 baris / 10 MB, pemetaan kolom, laporan per baris) | ✅ | `contacts/import` | `internal/contact/import.go` (`internal/importjob` belum dipakai) |
| CONT-03 | Segmen | ✅ | `contacts/segments` | `internal/segment` |
| CONT-04 | Tag | ✅ | `contacts/tags` | `internal/tag` |
| CONT-05 | Provider profile (satu kontak, banyak akun) | ✅ | `contacts` | ADR [0003](../adr/0003-provider-profile.md), [0004](../adr/0004-contact-linking.md) |
| CONT-06 | Enrich profil Instagram | ✅ | – | `cmd/enrich-ig-profiles` |

## F-BCAST — Broadcast
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| BCAST-01 | Sinkronisasi template dari Meta | ✅ | `broadcast/templates` | `internal/template` |
| BCAST-02 | Buat / hapus template | ✅ | `broadcast/templates/create` | `internal/template` |
| BCAST-03 | Pilih template dengan search + pagination | 🟡 | `broadcast/campaigns/create` | Bug terbuka #2 di [bug list](../bugs/campaign-broadcast-bugs.md) |
| BCAST-04 | Wizard kampanye + mapping variabel | 🟡 | `broadcast/campaigns/create` | `internal/campaign`, Bug #1 (header image), #3 (template EN), #4 (struktur kontak) |
| BCAST-05 | Eksekusi broadcast (orchestrator + worker, rate limit, retry) | ✅ | – | `cmd/broadcast-orchestrator`, `cmd/broadcast-worker` |
| BCAST-06 | Jadwal kampanye | 🟡 | `broadcast/schedules` | Edit jadwal di UI masih **mock** |
| BCAST-07 | Laporan kampanye & status per penerima | ✅ | `broadcast/reports`, `broadcast/reports/[id]` | `internal/report`, `internal/campaign/delivery` |
| BCAST-08 | Pembatalan kampanye | ✅ | `broadcast/campaigns/[id]` | `pkg/broadcastcancel` |
| BCAST-09 | Rekonsiliasi counter harian | ✅ | – | `internal/campaign/delivery/reconcile.go` |

## F-AUTO — Automation & Integrasi
| ID | Fitur | Status | Frontend | Backend |
|---|---|---|---|---|
| AUTO-01 | Endpoint automation (wizard, detail) | ✅ | `automation/endpoints*` | `internal/automation` |
| AUTO-02 | API key | ✅ | `automation/keys` | `internal/automation` |
| AUTO-03 | Invoke publik lewat webhook-service | 🟡 | – | Validasi mapping tombol URL masih membingungkan (lihat bug list) |
| AUTO-04 | Forward `message.received` ke URL pelanggan | ✅ | `automation/webhooks` | `internal/automation` |
| AUTO-05 | Test invoke dari UI | 🧪 | `automation/endpoints/[id]` | Teks UI masih menyebut "mock executor" |
| AUTO-06 | OTP via template | ✅ | – | `internal/otp` |
| AUTO-07 | Workflow builder | 🧪 | `automation/workflows` (placeholder) | ⬜ |
| AUTO-08 | Halaman integrations | 🧪 | `integrations` (redirect) | ⬜ |

## F-REPORT — Analitik
| ID | Fitur | Status | Catatan |
|---|---|---|---|
| REP-01 | Dashboard workspace | 🟡 | Saat ini hanya readiness, belum ada metrik percakapan |
| REP-02 | Analitik inbox (volume, FRT, resolusi) | ⬜ | R3 |
| REP-03 | Laporan performa agent | ⬜ | R3 |

## R3 — Otomasi percakapan
| ID | Fitur | Status |
|---|---|---|
| R3-01 | Keyword reply | ⬜ |
| R3-02 | Welcome message | ⬜ |

## Ditunda
| ID | Fitur | Status | Catatan |
|---|---|---|---|
| BILL-01 | Billing & subscription | 🧪 | `billing` placeholder; `pkg/planlimits` dan `internal/quota` sudah ada |

---

## Utang teknis di frontend yang terkait fitur
- ✅ `_legacy/` dan file mock yang tidak terpakai sudah dihapus (TODO 2.5). Masih dipakai: `broadcast/mock/template-legacy-adapters.ts` (oleh `utils/template-consumer.ts`) dan `broadcast/mock/template-mock-data.ts` (oleh wizard automation) — ganti dengan data API sebelum dihapus.
- Route redirect lama: `broadcast/campaign`, `broadcast/schedule`, `integrations`.

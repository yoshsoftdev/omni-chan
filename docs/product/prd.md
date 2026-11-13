# PRD — Omni Chan (Omnichannel Team Inbox & Broadcast)

| | |
|---|---|
| Status | **Draft v2**: perlu review product owner |
| Menggantikan | [`archive/prd-legacy.md`](../archive/prd-legacy.md) |
| Terkait | [Daftar fitur](features.md) · [Tech stack](../architecture/tech-stack.md) · [Alur data](../architecture/data-flow.md) · [Roadmap perbaikan](../roadmap/improvement-plan.md) |

> Bagian yang ditandai **[KONFIRMASI]** adalah asumsi yang disusun dari kode dan dokumen yang ada. Bagian itu harus divalidasi oleh pemilik produk.

---

## 1. Ringkasan

Omni Chan adalah SaaS multi-tenant untuk tim bisnis yang melayani pelanggan dan mengirim kampanye lewat **WhatsApp Business Platform (Cloud API)**, **Instagram DM**, dan (sebagian) **Telegram**. Semua percakapan masuk ke satu **Team Inbox** realtime. Kampanye massal dikirim memakai template Meta yang sudah disetujui. Sistem eksternal (ERP, n8n) dapat memicu pesan lewat **Automation API**.

## 2. Masalah

1. Tim CS memakai satu nomor WhatsApp di banyak HP/WA Web. Akibatnya tidak ada pembagian tugas, tidak ada riwayat bersama, dan percakapan sering terlewat.
2. Kampanye broadcast manual rawan diblokir Meta, tidak terukur (sent/delivered/read), dan tidak bisa dijadwalkan.
3. Sistem internal (ERP, OTP, notifikasi transaksi) tidak punya jalur resmi dan aman untuk mengirim pesan WhatsApp.

## 3. Target pengguna & persona **[KONFIRMASI]**

| Persona | Kebutuhan utama |
|---|---|
| **Owner / Admin workspace** | Menghubungkan channel (Embedded Signup), mengatur tim & role, melihat performa |
| **Agent CS** | Membalas cepat, melihat konteks kontak, quick reply, assignment yang adil |
| **Supervisor** | Memantau SLA, beban agent, mengalihkan percakapan |
| **Marketer** | Membuat segmen, memilih template, menjadwalkan kampanye, membaca laporan |
| **Developer integrator** | API key, endpoint automation, webhook forward pesan masuk |

Segmen awal: UMKM sampai bisnis menengah di Indonesia yang sudah memakai WhatsApp sebagai kanal utama.

## 4. Tujuan & metrik keberhasilan **[KONFIRMASI target angka]**

| Tujuan | Metrik | Target usulan |
|---|---|---|
| Respons pelanggan lebih cepat | Median first response time | < 5 menit di jam kerja |
| Tidak ada pesan hilang | Pesan masuk yang tidak muncul di inbox | 0 (diukur dari DLQ + rekonsiliasi) |
| Broadcast andal | Delivery rate kampanye (delivered/sent) | ≥ 95% (di luar error nomor invalid) |
| Realtime | Latensi webhook → tampil di inbox | p95 < 2 detik |
| Adopsi | Workspace aktif mingguan dengan ≥ 1 channel terhubung | ditentukan PO |

## 5. Ruang lingkup

### 5.1 Dalam lingkup (v1)
- Auth, workspace, undangan, team & RBAC
- Channel: WhatsApp Cloud API (Embedded Signup, coexistence), Instagram Business Login
- Team Inbox realtime: assignment, status & lifecycle, SLA, tag, catatan, quick reply, window recontact
- Kontak: CRUD, import CSV/Excel, segmen, tag, consent, provider profile (satu kontak bisa punya banyak akun channel)
- Broadcast: sinkronisasi & pembuatan template, kampanye, jadwal, laporan pengiriman
- Automation: endpoint invoke (template/teks), API key, forward webhook pesan masuk, OTP
- Media: upload/simpan file, proxy media untuk Instagram

### 5.2 Di luar lingkup v1
- Billing/subscription (sudah di-rollback, halaman masih placeholder)
- Workflow builder visual (halaman placeholder)
- Chatbot/AI auto-reply
- Kanal lain (email, SMS, Facebook Messenger)

## 6. Kebutuhan fungsional

Rincian per modul beserta status implementasi ada di **[features.md](features.md)**. Ringkasannya sebagai berikut.

| Kode | Modul | Inti kebutuhan |
|---|---|---|
| F-AUTH | Auth & Workspace | Login/register, refresh token, undangan, workspace readiness checklist |
| F-TEAM | Team & RBAC | Role kustom, permission matrix, activity log; **semua modul wajib cek permission** |
| F-CHAN | Channel | Embedded Signup WA, health/readiness, sinkronisasi template, Instagram OAuth |
| F-INBOX | Team Inbox | Daftar percakapan realtime, balas teks/media/template, assignment otomatis (least-loaded), SLA, collision detection, presence |
| F-CONT | Kontak | Import, dedup by phone, segmen dinamis, tag, consent opt-in/out |
| F-BCAST | Broadcast | Template → audience → mapping variabel → jadwal → kirim bertahap dengan rate limit → laporan status per penerima |
| F-AUTO | Automation | Endpoint per use-case, API key ber-scope, invoke publik lewat webhook-service, forward `message.received` ke URL pelanggan |
| F-MEDIA | Media | Upload, penyimpanan (local/MinIO), token media berumur pendek untuk Meta |
| F-REPORT | Laporan & Analitik | Laporan kampanye (ada); analitik inbox & laporan agent (**belum**) |

## 7. Kebutuhan non-fungsional

| Area | Kebutuhan |
|---|---|
| Multi-tenant | Semua query di-scope `team_id`; tidak boleh ada kebocoran data lintas tenant (lihat [multi-tenant audit](../archive/reviews/multi-tenant-audit.md)) |
| Keandalan | Event webhook **at-least-once** dan diproses **idempoten**; pesan keluar tidak boleh terkirim ganda |
| Skalabilitas | API stateless dan bisa di-scale horizontal; worker broadcast bisa di-scale terpisah; job singleton tidak boleh dobel |
| Kinerja | p95 API < 300 ms untuk list endpoint; inbox realtime p95 < 2 detik |
| Keamanan | Hanya webhook-service yang terbuka ke internet; verifikasi signature Meta; token Meta terenkripsi; secret tidak pernah masuk repo |
| Observability | Log terstruktur, metrik, tracing lintas service, alert (lihat roadmap fase 4) |
| Kualitas | CI wajib hijau; unit test untuk domain logic; e2e untuk alur utama (lihat roadmap fase 3) |
| Aksesibilitas UI | Kontras WCAG AA sesuai `.cursor/rules/ui-design.mdc` |

## 8. Batasan eksternal (Meta)
- Pesan bebas hanya dalam **customer service window 24 jam**; di luar window wajib memakai template yang sudah disetujui.
- Template harus disetujui Meta. Template dengan header media tetap butuh link media saat dikirim (lihat [bug list](../bugs/campaign-broadcast-bugs.md)).
- Ada rate limit dan tier pengiriman per nomor. Kualitas nomor bisa turun jika banyak di-report.

## 9. Rilis bertahap

| Rilis | Isi | Status |
|---|---|---|
| R1 | Auth, Team, Channel, Realtime Inbox | Sebagian besar selesai |
| R2 | Kontak, Broadcast, Template | Selesai, ada bug terbuka |
| R2.5 | Automation API, OTP, Instagram | Selesai, perlu hardening |
| R3 | Keyword reply, welcome message, analitik, laporan agent | Belum mulai |
| R4 | Billing & subscription | Ditunda |

## 10. Pertanyaan terbuka
1. Apakah Telegram masuk lingkup v1? Saat ini baru ada normalizer, belum ada alur end-to-end.
2. Target angka di §4 dan model harga (mempengaruhi `quota`/`planlimits`).
3. Apakah migrasi ke Kafka didorong kebutuhan nyata (retensi panjang, replay, integrasi data) atau preferensi tim? Lihat roadmap fase 8.
4. Siapa pemilik design system dan apakah ada desain Figma sebagai acuan?

# Dokumentasi Omni Chan

Mulai dari sini. Urutan baca untuk anggota baru: **PRD → Fitur → Tech stack → Alur data → Roadmap**.

## Struktur

```
docs/
├── README.md                  ← index ini
├── TODO.md                    ← checklist progres (centang tiap butir selesai)
├── product/                   apa yang dibangun
│   ├── prd.md
│   └── features.md            status fitur (sumber kebenaran tunggal)
├── architecture/              bagaimana sistem bekerja
│   ├── overview.md
│   ├── tech-stack.md
│   ├── data-flow.md
│   ├── api-v1-contract.md
│   ├── database-schema-v1.md
│   └── diagrams/              mermaid sistem
├── adr/                       keputusan arsitektur (immutable, bernomor)
├── design/                    audit UI & design system
├── domain/                    desain per domain
│   ├── broadcast/  contacts/  events/  inbox/  meta/
├── openapi/v1.yaml            kontrak API mesin-baca
├── security/                  review keamanan & strategi token
├── runbooks/                  prosedur operasional
├── roadmap/
│   └── improvement-plan.md    rencana perbaikan bertahap
├── bugs/                      catatan bug sebelum masuk issue tracker
└── archive/                   historis, read-only
    ├── phases/                catatan per fase implementasi (9.x–10.x, sprint 6)
    ├── reviews/               review & audit fase 10–12
    ├── patches/  samples/  zips/
    ├── prd-legacy.md
    └── tasks-legacy.md
```

## Dokumen utama

| Topik | Dokumen |
|---|---|
| Produk | [PRD](product/prd.md) · [Daftar fitur](product/features.md) |
| Arsitektur | [Overview](architecture/overview.md) · [Tech stack](architecture/tech-stack.md) · [Alur data](architecture/data-flow.md) |
| Kontrak | [API v1](architecture/api-v1-contract.md) · [OpenAPI](openapi/v1.yaml) · [Skema DB](architecture/database-schema-v1.md) |
| Keputusan | [ADR index](adr/README.md) |
| Desain UI | [Design system](design/design-system.md) · [Audit UI](design/ui-audit.md) |
| Event | [Envelope](domain/events/event-envelope.md) · [Subjects](domain/events/nats-subjects.md) · [Idempotensi](domain/events/idempotency-strategy.md) · [DLQ](domain/events/dead-letter-strategy.md) · [Tenant resolution](domain/events/tenant-resolution.md) |
| Broadcast | [Execution architecture](domain/broadcast/execution-architecture.md) · [Failure classification](domain/broadcast/failure-classification.md) · [Reconciliation](domain/broadcast/reconciliation.md) · [Runbook recovery](runbooks/broadcast-recovery.md) |
| Inbox | [Assignment](domain/inbox/assignment-rules.md) · [Lifecycle](domain/inbox/conversation-lifecycle.md) · [SLA](domain/inbox/sla-tracking.md) |
| Meta | [Embedded signup](domain/meta/embedded-signup-flow.md) · [Template sync](domain/meta/template-sync-strategy.md) · [Webhook catalog](domain/meta/webhook-event-catalog.md) |
| Keamanan | [Webhook](security/webhook-security-review.md) · [Token Meta](security/meta-token-strategy.md) · [Embedded signup](security/embedded-signup-security-review.md) |
| Rencana | [**TODO / progres**](TODO.md) · [Roadmap perbaikan](roadmap/improvement-plan.md) |

## Aturan menulis dokumen

1. **Satu topik, satu tempat.** Update dokumen yang sudah ada, jangan buat salinan baru.
2. Keputusan arsitektur → ADR baru di `adr/` (nomor berikutnya: **0008**). Jangan mengedit ADR yang sudah diterima; buat ADR pengganti.
3. Status fitur hanya di `product/features.md`, progres pekerjaan di `TODO.md`; keduanya di-update dalam PR yang sama dengan kodenya.
4. Catatan fase/review yang sudah selesai → `archive/`.
5. **Jangan pernah** menulis token, API key, password, atau nomor pelanggan asli di dokumen. Pakai `<REDACTED>` atau data contoh.
6. Nama file: `kebab-case.md`, bahasa Indonesia atau Inggris secara konsisten dalam satu dokumen.

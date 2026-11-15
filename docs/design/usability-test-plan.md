# Uji usability fase 7 (butir 7.7)

Tujuan: memastikan modul yang direvamp di 7.1–7.6 bisa dipakai pengguna nyata tanpa bantuan.
Audit aksesibilitas otomatis (axe) sudah berjalan di CI: `frontend/e2e/a11y.e2e.ts`.

## Cara menjalankan

- **Peserta:** 3–5 orang per modul. Pakai agent/admin sungguhan; jangan tim developer.
- **Lingkungan:** staging dengan data contoh (kontak, template yang sudah disetujui, satu channel WhatsApp). Jangan pakai nomor pelanggan asli. Test invoke automation mengirim pesan **sungguhan**.
- **Sesi:** 30 menit per orang, berbagi layar, minta peserta berpikir keras. Fasilitator tidak membantu, kecuali peserta macet lebih dari 3 menit (catat sebagai "dibantu").
- **Catat per tugas:** berhasil / dibantu / gagal, waktu, dan kutipan saat peserta bingung.
- **Setelah sesi:** SEQ (1–7, "seberapa mudah tugas ini?") per tugas, lalu SUS di akhir.

## Tugas per modul

| Modul | Tugas | Berhasil bila |
|---|---|---|
| Inbox (7.1) | Buka percakapan belum dibaca berikutnya tanpa mouse, lalu balas. | Memakai `j`/`r` atau menemukannya dari petunjuk; balasan terkirim. |
| Inbox | Sebutkan berapa lama lagi Anda bisa membalas pelanggan ini dengan teks bebas. | Membaca chip "24h window". |
| Broadcast (7.2) | Buat kampanye dari template bergambar hasil sync Meta, jadwalkan besok pukul 09.00. | Mengunggah gambar header; kampanye tersimpan terjadwal. |
| Broadcast | Temukan template "promo" di antara banyak template. | Memakai pencarian/paginasi. |
| Kontak (7.3) | Impor CSV contoh (kolom berbahasa Indonesia), lalu perbaiki baris yang dilewati. | Memetakan kolom; membaca laporan per baris. |
| Kontak | Beri tag "VIP" ke tiga kontak sekaligus. | Memakai bulk action. |
| Channel (7.4) | Channel ini belum bisa mengirim. Apa yang harus diperbaiki? | Menyebut item merah di checklist Readiness. |
| Automation (7.5) | Kirim pesan uji dari endpoint ke nomor Anda sendiri, lalu salin contoh cURL. | Mengonfirmasi kirim nyata; snippet tersalin. |
| Team & settings (7.6) | Undang anggota baru sebagai supervisor, lalu hapus anggota lain. | Undangan terkirim; penghapusan dikonfirmasi. |

## Ambang lulus

- Setiap tugas ≥ 80% berhasil tanpa bantuan, dan median SEQ ≥ 5.
- SUS rata-rata ≥ 68.
- Temuan dengan dampak "gagal" menjadi issue (butir 2.4) sebelum 7.7 dicentang.

## Hasil

| Modul | Tanggal | Peserta | Berhasil tanpa bantuan | Median SEQ | Temuan utama |
|---|---|---|---|---|---|
| | | | | | |

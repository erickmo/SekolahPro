# Cerebrum

> OpenWolf's learning memory. Updated automatically as the AI learns from interactions.
> Do not edit manually unless correcting an error.
> Last updated: 2026-04-26

## User Preferences

<!-- How the user likes things done. Code style, tools, patterns, communication. -->

## Key Learnings

- **Project:** SekolahPro boilerplate — sistem manajemen sekolah (siswa, guru, orang tua, staf, antar-sekolah). Arsitektur hybrid CQRS + Vernon read-cache pattern. Go backend, PostgreSQL JSONB. Branch aktif: `feat/vernon-hybrid`.
- **Cross-domain boundary:** Sekolah dan Koperasi TIDAK boleh JOIN langsung ke tabel domain lain. Semua integrasi via domain events (NATS JetStream) atau via `school_entity_id` FK yang di-resolve melalui API call.
- **Source of truth:** Sekolah memiliki data person (students, teachers, guardians). Koperasi hanya menyimpan `school_entity_id` sebagai referensi — tidak menduplikasi data person.
- **Critical events:** `StudentActivated`, `TeacherPayrollDataReady`, `InvoicePaid`, `TransaksiCreated`, `NasabahApproved`, `PinjamanDisbursed`, `SHUDistributed`, `AngsuranOverdue` harus menggunakan NATS JetStream durable consumer dengan DLQ.
- **Kantin vs Toko:** Kantin Sekolah (S036-S037) menggunakan canteen_wallet terpisah, dikelola bendahara sekolah. Toko Koperasi (K019) menggunakan rekening tabungan nasabah, dikelola koperasi. Tidak direkomendasikan menjalankan keduanya bersamaan untuk satu sekolah.
- **SPP vs Simpanan:** SPP adalah biaya pendidikan (milik sekolah, tidak bisa ditarik). Simpanan Koperasi adalah tabungan (milik nasabah, bisa ditarik). Secara legal dan akuntansi keduanya terpisah meski orang tua yang sama yang membayar.
- **Portal orang tua:** Single JWT Phase 2 menyertakan permissions dari kedua domain. Frontend menampilkan Tab Sekolah dan Tab Koperasi di dashboard yang sama — implementasinya di sisi frontend, bukan backend.
- **DAPODIK mode utama:** `manual_export` (bukan API) karena DAPODIK tidak memiliki public REST API yang stabil di Phase 1. Engineer tidak boleh menghabiskan waktu untuk integrasi API Dapodik yang rapuh.
- **Dokumentasi cross-domain:** tersimpan di `docs/spec/cross-domain/` (01-05). Selalu cek sebelum implementasi fitur yang melibatkan lebih dari satu domain.

## Do-Not-Repeat

<!-- Mistakes made and corrected. Each entry prevents the same mistake recurring. -->
<!-- Format: [YYYY-MM-DD] Description of what went wrong and what to do instead. -->

## Decision Log

<!-- Significant technical decisions with rationale. Why X was chosen over Y. -->
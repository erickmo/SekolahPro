# 10 — Parent, Student, Member Experience

**Status**: Draft  
**Perspective**: Product + System Analyst  
**Primary Inputs**: `ADR-S042`, `ADR-S043`, `ADR-S044`, `ADR-S045`, `ADR-S009`, `ADR-S018`, cooperative member-facing flows

## Purpose

Mendefinisikan pengalaman pengguna eksternal utama: orang tua, siswa, dan anggota koperasi/member agar engagement, trust, dan retention meningkat.

## Experience Principles

- mobile-first
- read-mostly, action-when-needed
- scoped dan aman
- komunikasi terstruktur, bukan chaos grup informal
- relevansi tinggi berdasarkan anak/anggota dan event penting

## Personas

### Parent / Guardian
- memantau akademik, kehadiran, keuangan, rapor, pengumuman
- mengelola preferensi notifikasi dan multi-child view

### Student
- melihat jadwal, tugas, nilai, kehadiran, pengumuman
- akses sesuai umur dan kebijakan institusi

### Cooperative Member
- melihat saldo, transaksi, pembiayaan, angsuran, status layanan anggota
- menerima notifikasi transaksi dan kewajiban penting

## Capability Areas

### Dashboard
- summary per user dan per linked subject (child/member account)
- quick insights, outstanding actions, and recent events

### Communication
- direct messaging, class/group messaging, institutional broadcast
- read receipt, mute, archive, escalation path

### Notifications
- push, SMS, email, WhatsApp, in-app
- quiet hours, channel preference, urgency policy

### Documents & Downloads
- rapor, invoice, bukti bayar, surat, laporan yang boleh diakses user

## Business Rules

- portal parent bersifat read-only untuk data akademik inti
- akses selalu di-scope ke child/member yang sah
- messaging dan notification dipisahkan secara konseptual dan teknis
- data yang ditampilkan harus berasal dari source of truth, bukan salinan liar

## Acceptance Criteria

- `AC-FUNC`: pengguna eksternal dapat menjalankan kebutuhan utama tanpa bantuan operator
- `AC-AUTH`: tidak ada akses ke child/member yang tidak terhubung sah
- `AC-DATA`: dashboard summary konsisten dengan domain sumber dalam batas SLA consistency
- `AC-AUDIT`: perubahan preference, messaging action, dan access-sensitive event tercatat
- `AC-INT`: push/SMS/WA/email terhubung dengan preference dan event source
- `AC-NFR`: portal mobile-first, responsif, dan dapat dipakai pada koneksi umum pengguna Indonesia

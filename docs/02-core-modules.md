# 02 — Core Modules (M01–M12)

Modul-modul inti yang menjadi tulang punggung EDS. Semua sekolah minimal menggunakan modul ini.

---

## M01 — Administrasi Siswa (SIMS)
**Tier**: FREE | **Integrasi**: M02, M04, M09, M12, M13

Manajemen data siswa dari pendaftaran hingga kelulusan.

**Entitas**: `Student`, `Guardian`, `Enrollment`, `Attendance`, `AcademicRecord`

**Fitur**:
- [ ] Profil siswa lengkap (NIK, NISN, foto, data wali, riwayat)
- [ ] Import massal via Excel/CSV (gunakan skill `xlsx`)
- [ ] Mutasi siswa (masuk/keluar/pindah antar sekolah)
- [ ] Manajemen kelas & rombongan belajar
- [ ] Absensi digital: manual / QR / fingerprint / face recognition
- [ ] Riwayat akademik per semester
- [ ] Export laporan ke PDF/Excel

---

## M02 — Kartu Siswa Digital
**Tier**: FREE | **Integrasi**: M01, M32

- [ ] Generate kartu otomatis dari data M01
- [ ] QR Code unik per siswa (enkripsi HMAC-SHA256, anti-pemalsuan)
- [ ] Verifikasi QR via mobile app petugas
- [ ] Perpanjangan otomatis tiap tahun ajaran
- [ ] Template kartu kustomisasi per sekolah (logo, warna, layout)
- [ ] Download PDF kartu (skill `pdf`)
- [ ] Kartu virtual di mobile app siswa

**Spesifikasi QR**: `{studentId}:{schoolId}:{academicYear}:{hmac}` — di-encode base64url, diverifikasi server-side.

---

## M03 — Koperasi & Tabungan Siswa
**Tier**: BASIC | **Integrasi**: M13, M18, M19

**Entitas**: `SavingsAccount`, `Transaction`, `CashierSession`, `Product`

- [ ] Buku tabungan digital per siswa
- [ ] Setoran & penarikan (kasir atau self-service via portal ortu)
- [ ] Notifikasi ke orang tua saat transaksi (via M14)
- [ ] POS (Point of Sale) untuk produk koperasi
- [ ] Rekap keuangan harian/bulanan/tahunan
- [ ] Saldo minimum Rp 10.000 (configurable per sekolah)
- [ ] Export laporan untuk audit bendahara (skill `xlsx`)

**Rules**: Penarikan > batas tertentu perlu approval bendahara. Log transaksi immutable.

---

## M04 — Akademik & Kurikulum
**Tier**: BASIC | **Integrasi**: M05, M23, M24, M48, M49, M50

**Entitas**: `Subject`, `Curriculum`, `Schedule`, `Grade`, `ReportCard`

- [ ] Manajemen kurikulum: K13 & Merdeka Belajar
- [ ] Penjadwalan otomatis (constraint satisfaction algorithm)
- [ ] Input nilai per KD oleh guru
- [ ] Kalkulasi otomatis nilai akhir & predikat
- [ ] Generate rapor PDF (format Kemendikbud) — skill `pdf`
- [ ] Kalender akademik & agenda sekolah
- [ ] Narasi komentar wali kelas via Claude AI

---

## M05 — Ujian & AI Generate Soal ⭐
**Tier**: PRO (500 soal AI/bulan, bisa tambah add-on) | **Integrasi**: M04, M23, M50

Fitur unggulan EDS. Kepala kurikulum mengatur parameter, AI generate soal, siswa mengerjakan online.

### Konfigurasi Ujian (ExamConfig)

```typescript
interface ExamConfig {
  subject: string;
  topic: string;
  gradeLevel: string;
  curriculum: 'K13' | 'MERDEKA';
  difficulty: { easy: number; medium: number; hard: number };  // % masing-masing
  questionTypes: {
    multipleChoice: number; essay: number;
    trueOrFalse: number;    matching: number;
  };
  totalQuestions: number;
  timeLimit: number;          // menit
  bloomLevel?: 'C1'|'C2'|'C3'|'C4'|'C5'|'C6';
}
```

### Prompt Engineering

```typescript
const EXAM_SYSTEM_PROMPT = `
Kamu adalah pembuat soal ujian profesional untuk sekolah Indonesia.
Buat soal sesuai Taksonomi Bloom level {bloomLevel}.
Kurikulum: {curriculum}. Jenjang: {gradeLevel}.
Output HANYA JSON valid dengan struktur:
{
  "questions": [{
    "id": "string",
    "type": "multiple_choice|essay|true_false|matching",
    "bloomLevel": "C1-C6",
    "difficulty": "easy|medium|hard",
    "question": "...",
    "options": ["A. ...", "B. ...", "C. ...", "D. ..."],
    "answer": "...",
    "explanation": "...",
    "indicators": "indikator capaian pembelajaran"
  }]
}`;
```

### Fitur Lengkap

- [ ] Dashboard konfigurasi ujian (kepala kurikulum)
- [ ] Generate soal via Claude API dengan parameter kustom
- [ ] Bank soal — simpan & re-use soal yang sudah digenerate
- [ ] Review & edit soal sebelum publish
- [ ] Jadwal ujian online (auto open/close berdasarkan waktu)
- [ ] Anti-cheat: randomize urutan soal & pilihan jawaban per siswa
- [ ] Auto-grading soal objektif
- [ ] Manual grading essay dengan rubrik
- [ ] Analisis hasil: per siswa, per kelas, per topik, per KD
- [ ] Laporan kualitas pendidikan untuk kepsek (M12)
- [ ] Export soal ke PDF untuk ujian offline (skill `pdf`)
- [ ] Mode CBT standar nasional

### Quota & Rate Limiting

```typescript
// packages/api/modules/exam/quota.ts
export async function checkAIQuota(schoolId: string): Promise<QuotaStatus> {
  const config = await getFeatureConfig(schoolId, 'M05');
  const used = await getMonthlyAIUsage(schoolId, 'M05');
  return {
    limit: config?.aiQuotaMonthly ?? 500,
    used,
    remaining: (config?.aiQuotaMonthly ?? 500) - used,
    resetAt: endOfMonth(),
  };
}
```

---

## M06 — Ekstrakurikuler
**Tier**: BASIC | **Integrasi**: M04, M37, M57

- [ ] Daftar & deskripsi program ekskul
- [ ] Pendaftaran online oleh siswa (batas kuota per ekskul)
- [ ] Manajemen anggota & pembina
- [ ] Jadwal kegiatan & kalender
- [ ] Absensi kegiatan ekskul
- [ ] Penilaian ekskul (masuk rapor M04)
- [ ] Portofolio & prestasi ekskul → M37

---

## M07 — Website Sekolah & Portal Publik
**Tier**: FREE | **Integrasi**: M08, M17, M57

- [ ] Profil sekolah (visi, misi, akreditasi, sejarah)
- [ ] Berita, pengumuman & agenda
- [ ] Galeri foto & video
- [ ] Data guru & staff
- [ ] Prestasi sekolah
- [ ] Google Maps embed
- [ ] SEO-optimized (Next.js SSG)
- [ ] CMS untuk admin (TipTap rich text)
- [ ] Mobile-first, responsif

---

## M08 — PPDB Online
**Tier**: BASIC | **Integrasi**: M01, M07, M18, M56

**Alur**: Pengumuman → Pendaftaran → Upload Berkas → Seleksi → Pengumuman → Daftar Ulang → Import ke M01

- [ ] Formulir pendaftaran online multi-langkah
- [ ] Upload dokumen persyaratan (KK, akta, ijazah, foto)
- [ ] Verifikasi berkas oleh panitia
- [ ] Sistem seleksi: nilai / zonasi / prestasi (configurable)
- [ ] Pengumuman hasil via email & WA otomatis (M14)
- [ ] Pembayaran biaya pendaftaran (Midtrans/Xendit)
- [ ] Dashboard monitoring real-time untuk panitia
- [ ] Auto-import data ke M01 setelah daftar ulang selesai
- [ ] Integrasi listing seragam & perlengkapan (M56, via Listing)

---

## M09 — Laporan Dinas & Pemerintah
**Tier**: BASIC | **Integrasi**: M01, M04, M31, M40

- [ ] Dapodik — format CSV/Excel standar Kemendikbud
- [ ] Laporan Bulanan Kepala Sekolah
- [ ] Laporan BOS & e-RKAM/ARKAS
- [ ] Statistik siswa (jenis kelamin, agama, kecamatan)
- [ ] Laporan kehadiran guru & siswa
- [ ] Rapor mutu sekolah

Untuk sinkronisasi otomatis, lihat M40 di `docs/11-integrasi.md`.

---

## M10 — Stakeholder & Publisher
**Tier**: PRO | **Integrasi**: Listing Marketplace

- [ ] Portal mitra/penerbit dengan login terpisah (role `PUBLISHER`)
- [ ] Katalog buku & materi dari publisher
- [ ] Pemesanan buku/LKS digital
- [ ] Kontrak & dokumen kerja sama digital
- [ ] Notifikasi update materi baru

**Catatan SaaS**: Penerbit mendaftar via Listing Marketplace (kategori `PENERBIT`). Sekolah tier PRO bisa akses katalog penerbit yang sudah terdaftar.

---

## M11 — Guru Les / Bimbingan Belajar
**Tier**: LISTING (guru les daftar via Listing Marketplace) | **Integrasi**: M24, M04

- [ ] Profil guru les tampil di portal siswa (dari Listing M11)
- [ ] Booking sesi les oleh siswa/orang tua
- [ ] Jadwal & kalender sesi les
- [ ] Pembayaran in-app (Midtrans)
- [ ] Rating & ulasan
- [ ] Rekomendasi guru berdasarkan nilai lemah siswa (M24 + M04)

**Catatan SaaS**: Guru les bukan user EDS — mereka vendor di Listing Marketplace (kategori `GURU_LES`). Sekolah bisa toggle apakah listing ini muncul di portal mereka.

---

## M12 — Dashboard Kepala Sekolah
**Tier**: BASIC | **Integrasi**: semua modul

Command center — satu layar kondisi sekolah real-time.

**Widgets**:
```
┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Total Siswa  │ Hadir Hari   │ Rata2 Nilai  │ Saldo Kop.   │
├──────────────┴──────────────┴──────────────┴──────────────┤
│ Grafik kehadiran 30 hari terakhir                          │
├──────────────────────────┬─────────────────────────────────┤
│ Top siswa berprestasi    │ Agenda minggu ini               │
├──────────────────────────┴─────────────────────────────────┤
│ Alert EWS siswa berisiko (M23) — hanya jika tier PRO       │
├────────────────────────────────────────────────────────────┤
│ AI Weekly Insight — rangkuman kondisi sekolah (Claude API) │
└────────────────────────────────────────────────────────────┘
```

Widget AI Insight hanya muncul di tier PRO ke atas. Di tier BASIC, widget ini digantikan dengan teks statis.

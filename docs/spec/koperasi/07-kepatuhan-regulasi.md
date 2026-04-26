# Kepatuhan & Laporan Regulasi

Koperasi dan BMT di Indonesia wajib menyampaikan laporan berkala ke regulator — Dinas Koperasi dan OJK (jika terdaftar sebagai LKM). Modul ini menyediakan mesin pelaporan terintegrasi yang menghasilkan laporan secara otomatis dari data akuntansi dan operasional, mengurangi beban manual dan risiko human error.

---

## ADR References

- **ADR-K017** — Laporan Regulasi (Regulatory Reporting)

---

## Domain Entities

### Tabel `laporan_config` (Konfigurasi per Jenis Laporan)

| Field | Tipe | Keterangan |
|-------|------|------------|
| tenant_id | UUID (FK) | |
| report_type | VARCHAR | `neraca`, `laba_rugi`, `kolektibilitas`, `car`, `bmpk`, `arus_kas`, `keanggotaan`, `zakat`, `harian_kas`, dll |
| report_name | VARCHAR | Nama tampilan |
| report_category | ENUM | dinas_koperasi, ojk, islamic, internal |
| is_active | BOOLEAN | |
| frequency | ENUM | daily, weekly, monthly, quarterly, annually |
| auto_generate | BOOLEAN | |
| generate_cron | VARCHAR nullable | Cron expression untuk auto-generate |
| deadline_rule | VARCHAR nullable | e.g., `T+10`, `D+90` |
| reminder_days | JSONB | Array: `[30, 7, 3, 1]` — hari sebelum deadline |
| template_config | JSONB | Header/logo, footer/signatory, columns, coa_mapping |
| email_recipients | JSONB | List `{email, name, role}` |
| thresholds | JSONB | Threshold anomaly: npl_ratio_max, car_min, bmpk_max |

**Unique:** `(tenant_id, report_type)`

### Tabel `laporan` (Per Laporan yang Di-Generate)

| Field | Tipe | Keterangan |
|-------|------|------------|
| tenant_id | UUID (FK) | |
| branch_id | UUID nullable | null = konsolidasi |
| laporan_config_id | UUID (FK) | |
| report_type | VARCHAR | Denormalized dari config |
| period_start / period_end | DATE | |
| period_label | VARCHAR | e.g., `Januari 2026` |
| status | ENUM | draft, reviewed, final, submitted, overdue |
| version_number | INTEGER | Increment jika revisi |
| revision_of | UUID nullable | FK → laporan yang direvisi |
| report_data | JSONB | Snapshot data + rasio keuangan + validation results |
| pdf_url / excel_url | VARCHAR nullable | Output files |
| consolidation_level | ENUM | branch, company, tenant |
| source_branch_ids | JSONB | Branch yang di-konsolidasi |
| generated_at / generated_by | TIMESTAMPTZ / UUID | |
| reviewed_at / reviewed_by | TIMESTAMPTZ / UUID | |
| review_notes | TEXT nullable | |
| finalized_at / finalized_by | TIMESTAMPTZ / UUID | |
| submitted_at / submitted_by | TIMESTAMPTZ / UUID | |
| submit_proof_url | VARCHAR nullable | Screenshot/nomor registrasi |
| deadline_date | DATE nullable | |
| is_overdue | BOOLEAN | |

### Tabel `laporan_versi`

| Field | Tipe | Keterangan |
|-------|------|------------|
| laporan_id | UUID (FK) | |
| version_number | INTEGER | |
| report_data | JSONB | Snapshot data pada saat versi ini dibuat |
| validation_result | JSONB | `{errors, warnings, checks_passed, checks_failed}` |
| pdf_url / excel_url | VARCHAR nullable | |
| created_reason | VARCHAR | `auto_generate`, `manual_trigger`, `regenerate`, `revision` |

---

## Jenis Laporan per Regulator

## Landasan Regulasi

### Regulasi Pokok

- **UU No. 25 Tahun 1992** tentang Perkoperasian: Landasan hukum utama seluruh kegiatan koperasi di Indonesia.
- **PP No. 9 Tahun 1995** tentang Pelaksanaan Kegiatan Usaha Simpan Pinjam oleh Koperasi: Mengatur batasan kegiatan usaha KSPS, kewajiban pembukuan, persyaratan pengurus, dan kewajiban laporan tahunan ke dinas koperasi. Merupakan peraturan turunan wajib dari UU No. 25/1992.
- **POJK No. 62/POJK.05/2014** tentang Penyelenggaraan Usaha Lembaga Keuangan Mikro: wajib untuk koperasi yang mendaftar sebagai LKM (Lembaga Keuangan Mikro). Mencakup: rasio kecukupan modal, BMPK, kolektibilitas, PPAP.
- **POJK No. 14/POJK.05/2014** tentang Pembinaan dan Pengawasan LKM: mengatur pengawasan oleh OJK/dinas koperasi.
- **Fatwa DSN-MUI No. 17/DSN-MUI/IX/2000** tentang Sanksi (Ta'zir) atas Nasabah yang Mampu Menunda-nunda Pembayaran *(berlaku untuk BMT/koperasi syariah)*.

### Kepatuhan Akuntansi

- **PSAK 109** (Akuntansi Zakat, Infaq, dan Sedekah) — wajib untuk BMT yang menghimpun dan menyalurkan ZIS.
- **PSAK 101** (Penyajian Laporan Keuangan Syariah) — untuk BMT yang menyajikan laporan keuangan syariah.
- **PSAK 105** (Akuntansi Mudharabah) dan **PSAK 106** (Akuntansi Musyarakah) — untuk produk pembiayaan syariah.

---

## Jenis Laporan per Regulator

### Dinas Koperasi (Wajib untuk Semua Koperasi)

| Laporan | Frekuensi | Deadline | Keterangan |
|---------|-----------|----------|------------|
| Neraca (Balance Sheet) | Tahunan | Perlu RAT | Aset = Kewajiban + Ekuitas |
| Laporan Laba Rugi / SHU | Tahunan | Perlu RAT | Total Pendapatan - Total Beban = SHU |
| Laporan Arus Kas | Tahunan | Perlu RAT | Operasi + Investasi + Pendanaan |
| Laporan Perubahan Ekuitas | Tahunan | Perlu RAT | |
| Laporan RAT | Tahunan | D+90 setelah RAT | Notulen + Keputusan RAT |
| Data Keanggotaan | Tahunan | Perlu RAT | Total, baru, keluar, per kategori |
| Laporan Perkembangan Usaha | Tahunan | Perlu RAT | |

### OJK (Hanya Jika `is_lkm = true`)

| Laporan | Frekuensi | Deadline | Keterangan |
|---------|-----------|----------|------------|
| Laporan Keuangan Bulanan | Bulanan | T+10 | Neraca + L/R + Arus Kas |
| Laporan Kolektibilitas | Bulanan | T+10 | NPL per kelas Kol 1-5 |
| Laporan BMPK | Bulanan | T+10 | Eksposur per peminjam vs modal |
| Laporan CAR | Bulanan | T+10 | Modal Inti / ATMR |
| Laporan Profil Risiko | Triwulanan | T+10 | Risiko kredit, operasional, likuiditas |
| Laporan Perlindungan Konsumen | Tahunan | D+30 | |

### Islamic Mode (Otomatis Aktif Jika `coop_type = "islamic"`)

| Laporan | Frekuensi | Dasar Hukum |
|---------|-----------|-------------|
| Laporan Sumber & Penyaluran Dana Zakat | Bulanan/Tahunan | PSAK 109 |
| Laporan Dana Kebajikan (Qardhul Hasan) | Bulanan/Tahunan | PSAK 109 |
| Laporan Rekonsiliasi Pendapatan Bagi Hasil | Bulanan | PSAK 105 |
| Laporan Perubahan Dana Investasi Terikat | Tahunan | PSAK 105 |

### Internal

| Laporan | Frekuensi |
|---------|-----------|
| Laporan Harian Kas | Harian (auto) |
| Laporan Bulanan per Cabang | Bulanan (auto) |
| Dashboard Ringkasan Eksekutif | Real-time |
| Custom Report Builder | On-demand |

---

## Business Rules

### Pembuatan Laporan

1. Laporan di-generate **otomatis** dari data akuntansi (K015) dan operasional — bukan input manual
2. Scheduled auto-generation di akhir periode — berjalan via background job
3. Manual trigger oleh authorized user kapan saja
4. Setiap generation menyimpan **snapshot data** saat itu — perubahan data setelah generation tidak mempengaruhi laporan yang sudah dibuat
5. **Versioning**: setiap regenerasi/revisi menyimpan versi baru — versi lama tetap tersimpan
6. Laporan Dinas Koperasi **wajib** untuk semua tenant
7. Laporan OJK **aktif** hanya jika `is_lkm = true` di tenant config
8. Laporan Islamic Mode otomatis **aktif** jika `coop_type = "islamic"`

### Approval Flow

```
DRAFT → REVIEWED → FINAL → SUBMITTED
```

9. **DRAFT**: bisa di-regenerate (data diambil ulang); draft lama tersimpan sebagai versi
10. **REVIEWED**: Manager+ validasi angka — bisa dikembalikan ke DRAFT dengan catatan
11. **FINAL**: immutable — jika ada kesalahan, buat **revisi** (version baru), bukan edit
12. **SUBMITTED**: catat tanggal submit + bukti (screenshot/nomor registrasi)
13. Setiap transisi status menghasilkan audit trail

### Validasi Otomatis Sebelum Finalisasi

**Blocking (tidak bisa finalisasi jika gagal):**
- Semua field wajib terisi
- Neraca balance: Total Aset = Total Kewajiban + Ekuitas
- Laporan Laba Rugi konsisten dengan neraca
- Arus Kas: Saldo akhir = Saldo awal + net flow

**Warning (non-blocking — laporan bisa disubmit tapi warning dicatat):**
- NPL ratio > threshold (default 5%)
- CAR < minimum (default 8%)
- BMPK > limit (default 20% modal)
- Perubahan signifikan vs periode lalu (> 30%)
- Akun dengan saldo negatif abnormal

### Multi-Branch Consolidation

| Level | Scope | Eliminasi |
|-------|-------|-----------|
| Branch | Per cabang | - |
| Company | Gabungan semua cabang | Transaksi antar cabang dieliminasi |
| Tenant | Gabungan semua company | Transaksi antar company dieliminasi |

14. Laporan ke regulator **selalu** menggunakan level konsolidasi
15. Eliminasi menggunakan akun inter-branch (1901/2901) dari COA K015
16. Reconciliation check: total saldo per cabang (setelah eliminasi) = saldo konsolidasi

### Jadwal & Reminder

| Laporan | Auto-Generate | Deadline | Reminder |
|---------|---------------|----------|----------|
| Harian Kas | Ya, 00:30 | T+1 | - |
| Keuangan Bulanan (OJK) | Ya, tgl 1 02:00 | T+10 | 30d, 7d, 3d, 1d |
| Kolektibilitas (OJK) | Ya, tgl 1 02:30 | T+10 | 7d, 3d, 1d |
| RAT (Dinas Koperasi) | Tidak | D+90 | 60d, 30d, 14d, 7d, 1d |

17. Jika laporan belum disubmit setelah deadline, status → **OVERDUE** — alert ke Admin dan Manager
18. Holiday calendar awareness: deadline di hari libur mundur ke hari kerja sebelumnya

---

## Rasio Keuangan yang Dihitung Otomatis

| Rasio | Formula | Threshold |
|-------|---------|-----------|
| **NPL Ratio** | Total NPL (Kol 3+4+5) / Total Outstanding Loans | Alert: > 5% |
| **CAR** | Modal Inti (Tier 1) / ATMR × 100% | DANGER: < 8%; WARNING: 8-12%; OK: > 12% |
| **BMPK** | Max Single Borrower Exposure / Modal Sendiri | Alert: > 20% (single), > 25% (group) |
| **ROA** | Net Income / Total Assets | Informasi |
| **ROE** | Net Income / Equity | Informasi |
| **Liquidity Ratio** | Liquid Assets (1100+1200) / Short-term Liabilities | Informasi |
| **LDR/FDR** | Total Pinjaman/Pembiayaan / Total Simpanan | Alert: > 90% (gen) / > 95% (Islamic) |

### Formula CAR Detail

**Modal Inti (Tier 1):**
```
Modal Inti = 3100 Simpanan Pokok
           + 3200 Simpanan Wajib
           + 3300 Cadangan Umum
           + 3400 Cadangan Risiko
           + 3600 Donasi/Hibah
           (EXCLUDE: 3500 SHU Tahun Berjalan — belum disetujui RAT)
```

**ATMR Bobot Risiko:**

| Akun | Keterangan | Bobot |
|------|------------|-------|
| 1101 | Kas | 0% |
| 1102 | Bank — Giro | 0% |
| 1103 | Bank — Tabungan | 20% |
| 1301 | Piutang Pinjaman/Pembiayaan | 100% |
| 1302 | Piutang Bunga/Margin | 100% |
| 1309 | Cadangan Kerugian (kontra) | 0% |
| 1500 | Aktiva Tetap | 100% |

```
ATMR = SUM(saldo_akun × bobot_risiko) untuk semua akun aset
CAR = Modal Inti / ATMR × 100%
```

---

## Arsip & Distribusi

### Retention

19. Semua laporan dan versi disimpan **minimal 10 tahun** (sesuai regulasi retensi data koperasi)
20. Searchable by: type, period, status

### Distribusi

21. Auto-email ke Pengurus, Pengawas, dan custom recipients hanya untuk status FINAL atau SUBMITTED
22. Email berisi summary + link download — file tidak di-attach langsung
23. Download role-based access dari dashboard

---

## RBAC Summary

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View laporan (daftar + detail) | - | v | v | v |
| Download laporan | - | v | v | v |
| Generate laporan manual | - | v | v | v |
| Review laporan (DRAFT → REVIEWED) | - | - | v | v |
| Finalize laporan (REVIEWED → FINAL) | - | - | - | v |
| Submit laporan (FINAL → SUBMITTED) | - | - | - | v |
| Revise laporan | - | - | - | v |
| Configure laporan (template, schedule) | - | - | - | v |
| Configure anomaly thresholds | - | - | - | v |
| Configure email recipients | - | - | v | v |
| View dashboard rasio keuangan | - | v | v | v |
| Custom report builder | - | - | v | v |

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Auto-generate dari data akuntansi | Konsistensi; menghilangkan manual error; dari hari menjadi menit |
| Snapshot data saat generation | Laporan per 31 Desember harus konsisten meskipun ada transaksi di Januari |
| Versioning per generation | Tidak ada data yang hilang; audit trail untuk setiap regenerasi |
| FINAL = immutable | Laporan yang diterbitkan tidak berubah; revisi → versi baru |
| Multi-level consolidation dengan eliminasi | Laporan ke regulator menghilangkan transaksi antar cabang sesuai standar konsolidasi |
| Threshold anomaly non-blocking | Laporan tetap bisa disubmit; warning dicatat untuk review Manager/Admin |
| Retention 10 tahun | Ketentuan OJK tentang retensi data lembaga keuangan |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `jurnal` (K015) | Laporan ← Jurnal | Semua laporan keuangan dari data jurnal POSTED |
| `nasabah` (K001) | Laporan ← Keanggotaan | Data anggota untuk laporan keanggotaan |
| `pinjaman` (K007) | Laporan ← Pinjaman | Kolektibilitas, BMPK, NPL |
| `shu_periode` (K016) | Laporan ← SHU | Laporan SHU dan distribusi untuk RAT |
| `kas_harian` (K014) | Laporan ← Kas | Laporan kas harian |
| `zakat_collection` (K018) | Laporan ← Zakat | Laporan sumber dan penyaluran zakat *(Islamic)* |
| `laporan_config` | Scheduling → Background Job | Auto-generate dengan cron expression |

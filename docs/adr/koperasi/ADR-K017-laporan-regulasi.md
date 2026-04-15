# ADR-K017: Laporan Regulasi (Regulatory Reporting)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi dan BMT di Indonesia **wajib** menyampaikan laporan berkala ke berbagai regulator — Dinas Koperasi, OJK (jika terdaftar sebagai LKM), dan internal manajemen. Laporan ini meliputi laporan keuangan (neraca, laba rugi, arus kas), laporan keanggotaan, laporan kolektibilitas pinjaman, dan laporan khusus syariah untuk BMT.

Tanpa mesin pelaporan terintegrasi, pembuatan laporan regulasi dilakukan manual di spreadsheet — proses yang lambat, rawan error, dan sulit diaudit. Sistem harus mengakomodasi:

- **Auto-generate** dari data akuntansi (K015 jurnal/COA) dan data operasional
- **Multi-regulator**: format dan frekuensi berbeda per regulator (Dinas Koperasi vs OJK)
- **Dual-mode**: laporan tambahan untuk koperasi syariah (PSAK 101-110)
- **Multi-branch consolidation**: laporan cabang → konsolidasi company → tenant
- **Approval flow**: draft → reviewed → final → submitted — setiap tahap tercatat
- **Scheduling**: auto-generate di akhir periode, reminder sebelum deadline
- **Rasio keuangan otomatis**: NPL, CAR, BMPK, ROA, ROE, LDR/FDR

Referensi terkait:
- [ADR-K015](./ADR-K015-jurnal-coa.md) — Jurnal & COA sebagai sumber data akuntansi
- [ADR-K016](./ADR-K016-shu.md) — Perhitungan SHU untuk laporan laba rugi / SHU
- [ADR-K001](./ADR-K001-nasabah.md) — Data keanggotaan untuk laporan anggota
- [ADR-K007](./ADR-K007-pinjaman.md) — Data pinjaman untuk laporan kolektibilitas

## Decision

### 1. Jenis Laporan — Kategori per Regulator

Sistem mengakomodasi empat kategori laporan sesuai tujuan dan regulatornya:

```
DINAS KOPERASI (wajib untuk semua koperasi):
├── Laporan RAT (Annual Member Meeting report)
├── Neraca (Balance Sheet)
├── Laporan Laba Rugi / Perhitungan SHU
├── Laporan Arus Kas (Cash Flow Statement)
├── Laporan Perubahan Ekuitas
├── Data Keanggotaan (total, new, exit, by category)
└── Laporan Perkembangan Usaha (business development)

OJK (wajib untuk LKM — Lembaga Keuangan Mikro, jika terdaftar):
├── Laporan Keuangan Bulanan
├── Laporan Kolektibilitas (NPL report, klasifikasi 1-5)
├── Laporan BMPK (Batas Maksimum Pemberian Kredit)
├── Laporan CAR (Capital Adequacy Ratio)
├── Laporan Profil Risiko
└── Laporan Perlindungan Konsumen

ISLAMIC MODE ADDITIONAL (per PSAK Syariah 101-110):
├── Laporan Sumber & Penyaluran Dana Zakat
├── Laporan Dana Kebajikan (Qardhul Hasan)
├── Laporan Perubahan Dana Investasi Terikat
└── Laporan Rekonsiliasi Pendapatan Bagi Hasil

INTERNAL (manajemen):
├── Laporan Harian Kas (daily cash report)
├── Laporan Bulanan per Cabang (monthly branch report)
├── Dashboard Ringkasan Eksekutif
└── Custom report builder
```

**Aturan:**
- Laporan Dinas Koperasi **wajib** untuk semua tenant — tidak bisa dinonaktifkan
- Laporan OJK hanya **aktif** jika tenant menandai diri sebagai LKM terdaftar (`is_lkm = true` di tenant config)
- Laporan Islamic Mode otomatis **aktif** jika `coop_type = "islamic"` — tidak perlu konfigurasi tambahan
- Laporan Internal selalu tersedia untuk semua tenant
- Custom report builder memungkinkan tenant membuat laporan ad-hoc dari data yang tersedia

### 2. Report Generation Engine — Auto-Generate dari Data Akuntansi

Laporan di-generate **otomatis** dari data akuntansi (K015) dan data operasional — bukan input manual.

**Sumber data:** K015 (jurnal/COA untuk neraca, laba rugi, arus kas, ekuitas), K001 (keanggotaan), K007 (pinjaman/kolektibilitas/BMPK), K016 (SHU), K014 (kas harian).

**Mekanisme generation:**

```
Trigger (scheduled / manual)
        │
        v
┌──────────────────────────────────────┐
│ 1. Ambil data dari sumber terkait   │
│    (query akuntansi, nasabah, dll)   │
│                                      │
│ 2. Apply report template             │
│    (format sesuai jenis laporan)     │
│                                      │
│ 3. Validasi completeness             │
│    (semua field terisi, balance OK)  │
│                                      │
│ 4. Hitung rasio keuangan otomatis   │
│    (NPL, CAR, ROA, ROE, dll)        │
│                                      │
│ 5. Generate output (PDF/Excel/JSON) │
│                                      │
│ 6. Simpan sebagai draft              │
│    (laporan_versi status: DRAFT)     │
└──────────────────────────────────────┘
```

**Aturan:**
- Configurable period: **daily, monthly, quarterly, annually** — per jenis laporan
- Scheduled auto-generation di akhir periode (e.g., laporan bulanan auto-generate tanggal 1 bulan berikutnya)
- Manual trigger oleh authorized user kapan saja (e.g., laporan ad-hoc)
- Report versioning: **draft → reviewed → final → submitted** — setiap versi tersimpan
- Data snapshot: saat generation, data di-freeze sebagai snapshot — perubahan data setelah generation tidak mempengaruhi laporan yang sudah dibuat

### 3. Report Format & Output

| Format | Kegunaan | Template |
|--------|----------|----------|
| PDF | Pengajuan formal ke regulator, arsip cetak | Configurable per tenant (logo, header, footer, tanda tangan pejabat) |
| Excel/CSV | Analisis data, exchange ke sistem eksternal | Mengikuti format standar regulator jika ada |
| On-screen | Quick review di dashboard, drill-down ke detail | Built-in template |
| JSON | API response untuk integrasi sistem lain | Standard schema |

**Aturan:**
- Setiap output tersimpan di storage dan bisa diunduh kembali kapan saja
- Template PDF dikustomisasi oleh Admin: logo koperasi, alamat, nama pejabat penandatangan

### 4. Regulatory Compliance Validation — Auto-Validate Sebelum Submit

Sebelum laporan bisa di-finalisasi dan disubmit, sistem menjalankan **validasi otomatis** untuk memastikan kepatuhan.

```
Validation Engine:
├── Field Completeness
│   ├── Semua field wajib terisi                    ✓ / ✗
│   ├── Periode laporan sesuai                      ✓ / ✗
│   └── Tanda tangan pejabat tersedia               ✓ / ✗
│
├── Cross-Check Totals
│   ├── Neraca: Total Aset = Total Kewajiban + Ekuitas   ✓ / ✗
│   ├── Laba Rugi: Total consistent dengan neraca         ✓ / ✗
│   └── Arus Kas: Saldo akhir = Saldo awal + net flow    ✓ / ✗
│
├── Anomaly Detection (warning, tidak blocking)
│   ├── NPL ratio > threshold (default 5%)          ⚠ WARNING
│   ├── CAR < minimum (default 8%)                   ⚠ WARNING
│   ├── BMPK > limit (default 20% modal)             ⚠ WARNING
│   ├── Perubahan signifikan vs periode lalu (>30%)  ⚠ WARNING
│   └── Akun dengan saldo negatif abnormal           ⚠ WARNING
│
└── Compliance Checklist
    ├── Laporan dibuat dalam periode yang benar       ✓ / ✗
    ├── Approver memiliki wewenang yang sesuai        ✓ / ✗
    └── Dokumen pendukung terlampir (jika wajib)      ✓ / ✗
```

**Aturan:**
- **Cross-check errors** bersifat **blocking** — laporan tidak bisa di-finalisasi jika neraca tidak balance
- **Anomaly warnings** bersifat **non-blocking** — laporan tetap bisa disubmit, tapi warning ditampilkan dan dicatat
- Threshold anomaly **configurable per tenant** — default mengikuti standar OJK
- Compliance checklist otomatis di-generate per jenis laporan — reviewer tinggal centang
- Semua hasil validasi tersimpan bersama laporan (audit trail)

### 5. Rasio Keuangan Kunci — Auto-Calculate

Sistem menghitung rasio keuangan secara **otomatis** dari data akuntansi dan operasional:

| Kategori | Rasio | Formula | Sumber | Threshold |
|----------|-------|---------|--------|-----------|
| Kualitas Aset | NPL Ratio | Total NPL (kol 3+4+5) / Total Outstanding Loans | K007 | > 5% |
| Permodalan | CAR | Modal (3xxx) / ATMR | K015 | < 8% |
| Permodalan | BMPK | Max Single Borrower Exposure / Modal | K007, K015 | > 20% |
| Profitabilitas | ROA | Net Income / Total Assets | K015 | - |
| Profitabilitas | ROE | Net Income / Equity | K015 | - |
| Likuiditas | Liquidity Ratio | Liquid Assets (1100+1200) / Short-term Liabilities (2100+2200) | K015 | - |
| Likuiditas | FDR/LDR | Total Financing / Total Deposits | K007, K015 | > 90% (general) / > 95% (islamic) |

**Aturan:**
- Rasio dihitung **otomatis** setiap kali laporan di-generate — tidak perlu input manual
- Threshold warning **configurable per tenant** — default mengikuti standar regulasi
- Dual-mode: koperasi konvensional menggunakan **LDR**, BMT menggunakan **FDR** — formula sama, terminologi berbeda
- Tren rasio disimpan **historis** — ditampilkan di dashboard eksekutif secara real-time
- ATMR dihitung dengan bobot standar OJK: Kas & SBI 0%, Antar bank 20%, Pinjaman anggota 100%, Aktiva tetap 100%

### 6. Multi-Branch Consolidation

Laporan mendukung tiga level aggregasi:

| Level | Scope | Eliminasi | Kegunaan |
|-------|-------|-----------|----------|
| **Branch** | Per cabang | - | Monitoring cabang |
| **Company** | Gabungan semua cabang | Transfer, hutang-piutang, pendapatan/beban antar cabang | Laporan ke regulator |
| **Tenant** | Gabungan semua company | Transaksi antar company | Konsolidasi holding |

**Aturan:**
- Konsolidasi **otomatis** mengeliminasi transaksi antar cabang — menggunakan kode akun inter-branch (1901/2901) dari K015
- Laporan regulator **selalu** menggunakan level konsolidasi (bukan per cabang)
- Drill-down dari konsolidasi ke cabang tersedia di on-screen view
- Reconciliation check: total saldo per cabang (setelah eliminasi) harus sama dengan saldo konsolidasi

### 7. Report Distribution — Distribusi Laporan

Laporan yang sudah final didistribusikan ke pihak terkait:

- **Auto-email**: ke Pengurus, Pengawas, Dinas Koperasi (jika tersedia), dan custom recipients per tenant. Hanya untuk status **FINAL** atau **SUBMITTED**. Email berisi summary + link download (file tidak di-attach langsung)
- **Download dari dashboard**: PDF/Excel, role-based access
- **Archive permanen**: semua laporan dan versi tersimpan permanent (minimal 10 tahun sesuai regulasi). Searchable by type, period, status

### 8. Report Scheduling — Penjadwalan & Reminder

Setiap jenis laporan memiliki jadwal yang **configurable**:

| Laporan | Frequency | Auto-Generate | Deadline | Reminders |
|---------|-----------|---------------|----------|-----------|
| Harian Kas | Daily | Ya, 00:30 | T+1 | - |
| Keuangan Bulanan (OJK) | Monthly | Ya, tgl 1 02:00 | T+10 | 30d, 7d, 3d, 1d |
| Kolektibilitas (OJK) | Monthly | Ya, tgl 1 02:30 | T+10 | 7d, 3d, 1d |
| RAT (Dinas Koperasi) | Annually | Tidak (manual) | D+90 | 60d, 30d, 14d, 7d, 1d |
| Custom Reports | Configurable | Configurable | Configurable | Configurable |

**Aturan:**
- Auto-generate berjalan via **background job** (scheduled task) — tidak memblokir operasi harian
- Reminder dikirim via notifikasi in-app dan email ke **designated approver**
- Jika laporan belum disubmit setelah deadline, status otomatis menjadi **OVERDUE** — alert ke Admin dan Manager
- Jadwal **configurable per tenant** — setiap tenant bisa menyesuaikan jadwal sesuai kebutuhan internal
- Holiday calendar awareness: jika deadline jatuh di hari libur, mundur ke hari kerja sebelumnya (configurable)

### 9. Approval Flow — Draft sampai Submitted

Setiap laporan melalui proses approval bertahap:

```
DRAFT → REVIEWED → FINAL → SUBMITTED
  │         │         │         │
  │         │         │         └─ Admin: catat tanggal submit + bukti
  │         │         └─ Admin: lock laporan, generate PDF final
  │         └─ Manager+: validasi angka, catatan review
  └─ Auto/Manual: generate snapshot, validation check
```

**Aturan:**
- **DRAFT** — bisa di-regenerate ulang (data diambil ulang). Draft lama tetap tersimpan sebagai versi
- **REVIEWED** — jika ada masalah, bisa dikembalikan ke DRAFT dengan catatan
- **FINAL** — immutable. Jika ada kesalahan, harus buat laporan **revisi** (versi baru, bukan edit)
- **SUBMITTED** — mencatat tanggal submit dan bukti (screenshot/nomor registrasi)
- Setiap transisi status menghasilkan **audit trail** (who, when, notes)

### 10. Audit Trail — Jejak Lengkap

Setiap aktivitas terkait laporan tercatat di field audit pada tabel `laporan`: `generated_at/by`, `reviewed_at/by` + `review_notes`, `finalized_at/by`, `submitted_at/by` + `submit_proof_url`, dan `revision_of`.

**Aturan:**
- Audit trail **immutable** — tidak bisa di-edit atau di-delete
- Setiap regenerasi (DRAFT ulang) menghasilkan **versi baru** di `laporan_versi` — versi lama tetap tersimpan
- Semua aksi (generate, review, finalize, submit, reject, revise) tercatat dengan timestamp dan actor

### 11. Data Model

```
laporan_config (konfigurasi per jenis laporan per tenant)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
│
├── ── Identitas ──
├── report_type           VARCHAR NOT NULL
│                         (e.g., "neraca", "laba_rugi", "kolektibilitas",
│                          "car", "bmpk", "arus_kas", "keanggotaan",
│                          "zakat", "dana_kebajikan", "harian_kas", dll)
├── report_name           VARCHAR NOT NULL
│                         (display name, e.g., "Neraca / Balance Sheet")
├── report_category       ENUM (dinas_koperasi, ojk, islamic, internal)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Scheduling ──
├── frequency             ENUM (daily, weekly, monthly, quarterly, annually)
├── auto_generate         BOOLEAN DEFAULT false
├── generate_cron         VARCHAR (nullable, cron expression for auto-generate)
├── deadline_rule         VARCHAR (nullable, e.g., "T+10", "D+90")
├── reminder_days         JSONB DEFAULT '[]'
│                         (e.g., [30, 7, 3, 1] — days before deadline)
│
├── ── Template ──
├── template_config       JSONB NOT NULL DEFAULT '{}'
│                         (header/logo, footer/signatory, columns, coa_mapping)
│
├── ── Distribution ──
├── email_recipients      JSONB DEFAULT '[]'
│                         (list {email, name, role})
│
├── ── Anomaly Thresholds ──
├── thresholds            JSONB DEFAULT '{}'
│                         (npl_ratio_max, car_min, bmpk_max, variance_max_pct)
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

```
laporan (record per laporan yang di-generate)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── branch_id             UUID (nullable, FK → branch, null = konsolidasi)
├── laporan_config_id     UUID (FK → laporan_config) NOT NULL
│
├── ── Identitas ──
├── report_type           VARCHAR NOT NULL (denormalized dari config)
├── report_category       ENUM (dinas_koperasi, ojk, islamic, internal)
├── period_type           ENUM (daily, weekly, monthly, quarterly, annually)
├── period_start          DATE NOT NULL
├── period_end            DATE NOT NULL
├── period_label          VARCHAR NOT NULL
│                         (e.g., "Januari 2026", "Q1 2026", "2026")
│
├── ── Status & Workflow ──
├── status                ENUM (draft, reviewed, final, submitted, overdue)
├── version_number        INTEGER NOT NULL DEFAULT 1
│                         (increment jika revisi)
├── revision_of           UUID (nullable, FK → laporan, jika ini revisi)
│
├── ── Data Snapshot ──
├── report_data           JSONB NOT NULL
│                         {generated_from, snapshot_at, data,
│                          ratios: {npl_ratio, car, roa, roe, ldr},
│                          validation: {errors[], warnings[]}}
│
├── ── Output Files ──
├── pdf_url               VARCHAR (nullable, URL file PDF)
├── excel_url             VARCHAR (nullable, URL file Excel)
│
├── ── Consolidation ──
├── consolidation_level   ENUM (branch, company, tenant)
├── source_branch_ids     JSONB DEFAULT '[]'
│                         (list branch_id yang di-konsolidasi)
│
├── ── Audit Trail ──
├── generated_at          TIMESTAMPTZ NOT NULL
├── generated_by          UUID NOT NULL
│                         (user atau 'SYSTEM' untuk auto-generate)
├── reviewed_at           TIMESTAMPTZ (nullable)
├── reviewed_by           UUID (nullable)
├── review_notes          TEXT (nullable)
├── finalized_at          TIMESTAMPTZ (nullable)
├── finalized_by          UUID (nullable)
├── submitted_at          TIMESTAMPTZ (nullable)
├── submitted_by          UUID (nullable)
├── submit_proof_url      VARCHAR (nullable)
│
├── ── Deadline ──
├── deadline_date         DATE (nullable)
├── is_overdue            BOOLEAN DEFAULT false
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

```
laporan_versi (versioning — setiap regenerasi/revisi menyimpan versi)
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant) NOT NULL
├── laporan_id            UUID (FK → laporan) NOT NULL
│
├── ── Versi ──
├── version_number        INTEGER NOT NULL
├── report_data           JSONB NOT NULL
│                         (snapshot data pada saat versi ini dibuat)
├── validation_result     JSONB NOT NULL DEFAULT '{}'
│   {
│     "errors": [],
│     "warnings": [],
│     "checks_passed": 12,
│     "checks_failed": 0
│   }
│
├── ── Output ──
├── pdf_url               VARCHAR (nullable)
├── excel_url             VARCHAR (nullable)
│
├── ── Metadata ──
├── created_reason        VARCHAR NOT NULL
│                         (e.g., "auto_generate", "manual_trigger",
│                          "regenerate", "revision")
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Unique constraints:**
- `laporan_config`: `(tenant_id, report_type)` — satu konfigurasi per jenis laporan per tenant
- `laporan`: `(tenant_id, report_type, period_start, period_end, version_number)` — satu laporan per jenis per periode per versi
- `laporan_versi`: `(laporan_id, version_number)` — satu versi number per laporan

### 12. Vernon _rels dan _data Structure

**laporan_config:**
- `_rels`: `{ "tenant_id": "018f..." }`
- `_data`: `{ "tenant": { "id": "...", "name": "Koperasi Sekolah Harapan" } }`

**laporan:**
- `_rels`: `{ "tenant_id", "branch_id", "laporan_config_id", "revision_of" }`
- `_data`: `{ "tenant": {id, name}, "branch": {id, name, code}, "laporan_config": {id, report_type, report_name} }`

**laporan_versi:**
- `_rels`: `{ "tenant_id", "laporan_id" }`
- `_data`: `{ "laporan": {id, report_type, period_label, status} }`

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di semua `laporan` cabang tersebut
- `LaporanConfigUpdatedEvent` → update `_data.laporan_config` di semua `laporan` yang mereferensikan config tersebut
- `LaporanStatusChangedEvent` → update `_data.laporan` di semua `laporan_versi` terkait

### 13. Dual-Mode Terminology

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| Neraca | Neraca | Laporan Posisi Keuangan |
| Laba Rugi | Laporan Laba Rugi | Laporan Laba Rugi (+ Laporan Pendapatan Bagi Hasil) |
| Rasio LDR/FDR | LDR (Loan to Deposit Ratio) | FDR (Financing to Deposit Ratio) |
| Pinjaman | Pinjaman | Pembiayaan |
| Bunga | Pendapatan Bunga | Pendapatan Margin/Bagi Hasil |
| Laporan Tambahan | - | Laporan Zakat, Dana Kebajikan, Investasi Terikat |
| Standar Akuntansi | PSAK / SAK ETAP | PSAK 101-110 (Syariah) |
| Kolektibilitas | Kredit Bermasalah (NPL) | Pembiayaan Bermasalah (NPF) |

**Aturan:**
- Terminologi otomatis diterapkan berdasarkan `coop_type` tenant — template menyediakan dua versi label
- Laporan Islamic Mode tambahan otomatis aktif jika `coop_type = "islamic"`
- Format numerik dan penanggalan mengikuti standar Indonesia

### 14. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View laporan (daftar) | - | v | v | v |
| View laporan detail (data + rasio) | - | v | v | v |
| Download laporan (PDF/Excel) | - | v | v | v |
| Generate laporan (manual trigger) | - | v | v | v |
| Review laporan (DRAFT → REVIEWED) | - | - | v | v |
| Finalize laporan (REVIEWED → FINAL) | - | - | - | v |
| Submit laporan (FINAL → SUBMITTED) | - | - | - | v |
| Revise laporan (buat revisi baru) | - | - | - | v |
| Configure laporan (template, schedule) | - | - | - | v |
| Configure anomaly thresholds | - | - | - | v |
| Configure email recipients | - | - | v | v |
| View dashboard rasio keuangan | - | v | v | v |
| Custom report builder | - | - | v | v |

**Catatan:**
- **Teller** tidak memiliki akses ke laporan regulasi — bukan operasional harian
- **Supervisor** bisa view dan generate untuk monitoring cabang
- **Manager** bisa review — memastikan kebenaran data sebelum finalisasi
- **Admin** kontrol penuh — finalisasi, submit, konfigurasi

## Consequences

### Positif

- **Otomatis** — laporan di-generate dari data akuntansi (K015), bukan input manual. Mengurangi human error dan waktu pembuatan dari hari menjadi menit
- **Compliant** — semua jenis laporan yang diwajibkan regulator (Dinas Koperasi, OJK) tersedia out-of-the-box
- **Auditable** — approval flow bertahap (draft → review → final → submit) dengan audit trail lengkap di setiap tahap
- **Proaktif** — scheduling dan reminder memastikan laporan dibuat dan disubmit tepat waktu, mengurangi risiko keterlambatan
- **Dual-mode ready** — laporan syariah (zakat, dana kebajikan) otomatis tersedia untuk BMT tanpa konfigurasi tambahan
- **Multi-branch** — konsolidasi otomatis dengan eliminasi transaksi antar cabang
- **Configurable** — template, threshold anomaly, schedule, dan recipients bisa disesuaikan per tenant
- **Versioned** — setiap regenerasi/revisi tersimpan sebagai versi terpisah, tidak ada data yang hilang

### Negatif

- **Complexity tinggi** — mesin validasi, konsolidasi multi-branch, dan auto-calculate rasio membutuhkan logika bisnis yang rumit
- **Storage** — setiap versi laporan menyimpan JSONB snapshot data yang bisa berukuran besar, terutama untuk laporan detail per anggota
- **Template maintenance** — perubahan format laporan regulator membutuhkan update template, yang mungkin berbeda per tenant
- **Scheduling dependency** — auto-generate bergantung pada background job infrastructure yang harus reliable (tidak boleh miss schedule)
- **Year-end dependency** — laporan tahunan (RAT, neraca akhir tahun) bergantung pada year-end closing K015 yang mungkin tertunda

### Mitigasi

- Complexity ditangani dengan **modular report engine** — setiap jenis laporan adalah module terpisah dengan interface standar
- Storage dioptimalkan dengan **compression** pada JSONB — versi draft lama di-archive setelah laporan final
- Template menggunakan **versioning** — format baru tidak menghapus template lama (laporan historis tetap konsisten)
- Background job menggunakan **retry mechanism** dengan dead-letter queue — retry 3x sebelum alert ke Admin
- Year-end closing dependency dimitigasi dengan **pre-check** dan reminder jika closing belum selesai saat deadline mendekat

## Alternatives Considered

### A. Laporan Manual (Export Data, Buat di Excel)

Sistem hanya menyediakan export data mentah (CSV/JSON), pembuatan laporan dilakukan manual di spreadsheet.

**Ditolak** karena: terlalu lambat untuk koperasi dengan banyak cabang, rawan inkonsistensi antar cabang, dan tidak bisa auto-validate. Koperasi kecil mungkin cukup dengan Excel, tapi solusi ini tidak scalable dan tidak memenuhi standar pelaporan OJK yang semakin ketat.

### B. Third-Party Reporting Tool (Jasper, BIRT, Metabase)

Integrasi dengan reporting tool eksternal untuk generate laporan.

**Ditolak** karena: menambah dependency infrastruktur dan biaya lisensi. Format laporan regulasi koperasi Indonesia sangat spesifik dan sering berubah — maintenance template di tool eksternal justru menambah complexity. Lebih baik built-in dengan format yang dikontrol penuh oleh sistem. Namun, JSON export tersedia jika tenant ingin integrasi dengan tool BI eksternal.

### C. Real-Time Reporting (tanpa snapshot)

Laporan selalu di-generate on-the-fly dari data terkini, tanpa menyimpan snapshot.

**Ditolak** karena: laporan regulasi membutuhkan **data point-in-time** — angka neraca per 31 Desember harus tetap sama meskipun ada transaksi di Januari. Tanpa snapshot, angka laporan akan berubah setiap kali di-view, yang melanggar prinsip akuntansi dan membuat audit impossible. Snapshot approach memastikan konsistensi dan auditability.

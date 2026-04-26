# ADR-S031: Staff Payroll (Penggajian Guru & Staff)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Penggajian guru dan staff di sekolah Indonesia sangat bervariasi tergantung status kepegawaian:

1. **Guru PNS/ASN**: Gaji pokok dari APBN/APBD melalui Dinas Pendidikan. Sekolah tidak mengelola gaji pokok, tapi mengelola tunjangan sertifikasi (TPG), tunjangan khusus daerah, dan insentif dari BOS/APBD.
2. **Guru Honorer**: Gaji dari dana BOS, APBD, atau yayasan. Besaran bervariasi — bisa per jam mengajar, per bulan tetap, atau kombinasi.
3. **Guru/Staff Yayasan (Swasta)**: Gaji sepenuhnya dari yayasan/sekolah. Struktur mirip karyawan swasta: gaji pokok + tunjangan + potongan.
4. **Pesantren**: Ustadz/ustadzah sering mendapat honorarium + tunjangan makan/tempat tinggal, bukan gaji formal.

Komponen penggajian:

| Komponen | PNS | Honorer | Yayasan |
|----------|-----|---------|---------|
| Gaji Pokok | Dari APBN (di luar sistem) | Dari BOS/Yayasan | Dari Yayasan |
| Tunjangan Sertifikasi (TPG) | 1x gaji pokok | - | - |
| Tunjangan Struktural | Jika jabatan struktural | - | Jika ada |
| Tunjangan Fungsional | Berdasarkan golongan | - | Jika ada |
| Tunjangan Transport | Dari sekolah/BOS | Dari BOS | Dari yayasan |
| Insentif Kehadiran | Dari BOS | Dari BOS | Dari yayasan |
| Potongan PPh 21 | Ya | Ya (jika > PTKP) | Ya |
| Potongan BPJS | Ya | Opsional | Ya |
| Potongan Ketidakhadiran | Per hari/kebijakan | Per hari | Per hari |

Regulasi terkait:
- **PP No. 15 Tahun 2019**: Penghasilan PNS guru (gaji pokok + tunjangan).
- **Permendikbud tentang BOS**: Dana BOS bisa untuk honor guru honorer (max 50% alokasi).
- **UU PPh Pasal 21**: Pajak penghasilan atas gaji karyawan.
- **PP No. 7 Tahun 2021**: BPJS Kesehatan dan Ketenagakerjaan.

### Mengapa Vernon Pattern?

- has_many dari teacher — satu record payroll per guru per bulan.
- Relasi ke teacher (ADR-012), academic_year, payroll_period.
- Read-heavy: slip gaji, rekap bulanan, laporan pajak tahunan.
- Write batch: payroll di-calculate sekali per bulan untuk semua guru.
- Eventually consistent acceptable — payroll bersifat bulanan, bukan real-time.

## Decision

Menggunakan **Vernon Pattern** untuk 4 tabel: `payroll_configs` (konfigurasi komponen gaji per sekolah), `payroll_periods` (periode penggajian), `payroll_entries` (slip gaji per guru per bulan), dan `payroll_components` (detail komponen per slip gaji).

### Table Schema

```sql
-- Konfigurasi komponen gaji per sekolah
CREATE TABLE payroll_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas komponen
    code            VARCHAR(50) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    component_type  VARCHAR(20) NOT NULL,
    category        VARCHAR(30) NOT NULL,

    -- Aturan perhitungan
    calculation_method VARCHAR(20) NOT NULL DEFAULT 'fixed',
    default_amount  NUMERIC(15,2),
    percentage_base VARCHAR(50),
    percentage_value NUMERIC(5,2),
    is_taxable      BOOLEAN NOT NULL DEFAULT true,

    -- Aplikabilitas
    applicable_to   VARCHAR(20) NOT NULL DEFAULT 'all',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_payroll_config_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_component_type CHECK (component_type IN ('earning', 'deduction')),
    CONSTRAINT chk_category CHECK (category IN (
        'gaji_pokok', 'tunjangan_sertifikasi', 'tunjangan_struktural',
        'tunjangan_fungsional', 'tunjangan_transport', 'tunjangan_makan',
        'insentif_kehadiran', 'honor_mengajar', 'lembur',
        'potongan_pph', 'potongan_bpjs_kesehatan', 'potongan_bpjs_ketenagakerjaan',
        'potongan_ketidakhadiran', 'potongan_pinjaman', 'potongan_lain'
    )),
    CONSTRAINT chk_calc_method CHECK (calculation_method IN ('fixed', 'percentage', 'per_hour', 'per_day', 'formula')),
    CONSTRAINT chk_applicable CHECK (applicable_to IN ('all', 'pns', 'honorer', 'yayasan')),
    CONSTRAINT chk_default_amount CHECK (default_amount IS NULL OR default_amount >= 0),
    CONSTRAINT chk_percentage CHECK (percentage_value IS NULL OR (percentage_value > 0 AND percentage_value <= 100))
);

-- Indexes
CREATE INDEX idx_payroll_cfg_tenant_company ON payroll_configs (tenant_id, company_id);
CREATE INDEX idx_payroll_cfg_type ON payroll_configs (component_type);
CREATE INDEX idx_payroll_cfg_category ON payroll_configs (category);
CREATE INDEX idx_payroll_cfg_active ON payroll_configs (is_active) WHERE is_active = true;
CREATE INDEX idx_payroll_cfg_rels ON payroll_configs USING GIN (_rels);
CREATE INDEX idx_payroll_cfg_data ON payroll_configs USING GIN (_data);

-- Periode penggajian (bulanan)
CREATE TABLE payroll_periods (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Periode
    year            INT NOT NULL,
    month           INT NOT NULL,
    period_name     VARCHAR(30) NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,

    -- Status workflow
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    total_gross     NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_deductions NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_net       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_entries   INT NOT NULL DEFAULT 0,

    -- Approval
    calculated_at   TIMESTAMPTZ,
    calculated_by   UUID,
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     UUID,
    approved_at     TIMESTAMPTZ,
    approved_by     UUID,
    disbursed_at    TIMESTAMPTZ,
    disbursed_by    UUID,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_payroll_period UNIQUE (tenant_id, company_id, year, month),
    CONSTRAINT chk_period_status CHECK (status IN (
        'draft', 'calculating', 'calculated', 'reviewed',
        'approved', 'disbursing', 'disbursed', 'cancelled'
    )),
    CONSTRAINT chk_period_month CHECK (month >= 1 AND month <= 12),
    CONSTRAINT chk_period_year CHECK (year >= 2020 AND year <= 2100),
    CONSTRAINT chk_period_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_total_gross CHECK (total_gross >= 0),
    CONSTRAINT chk_total_deductions CHECK (total_deductions >= 0),
    CONSTRAINT chk_total_net CHECK (total_net >= 0)
);

-- Indexes
CREATE INDEX idx_payroll_period_tenant_company ON payroll_periods (tenant_id, company_id);
CREATE INDEX idx_payroll_period_year_month ON payroll_periods (year, month);
CREATE INDEX idx_payroll_period_status ON payroll_periods (status);
CREATE INDEX idx_payroll_period_rels ON payroll_periods USING GIN (_rels);
CREATE INDEX idx_payroll_period_data ON payroll_periods USING GIN (_data);

-- Slip gaji per guru per bulan
CREATE TABLE payroll_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    payroll_period_id UUID NOT NULL,
    teacher_id      UUID NOT NULL,

    -- Teacher snapshot (saat payroll dihitung)
    employee_type   VARCHAR(20) NOT NULL,
    golongan        VARCHAR(10),
    masa_kerja_tahun INT,

    -- Attendance summary (from S026)
    days_present    INT NOT NULL DEFAULT 0,
    days_absent     INT NOT NULL DEFAULT 0,
    days_sick       INT NOT NULL DEFAULT 0,
    days_leave      INT NOT NULL DEFAULT 0,
    total_late_minutes INT NOT NULL DEFAULT 0,
    working_days    INT NOT NULL DEFAULT 0,

    -- Workload summary (from S027)
    teaching_hours_per_week NUMERIC(5,1) NOT NULL DEFAULT 0,
    total_workload_hours NUMERIC(5,1) NOT NULL DEFAULT 0,

    -- Totals
    total_earnings  NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_deductions NUMERIC(15,2) NOT NULL DEFAULT 0,
    net_pay         NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- Tax (PPh 21)
    taxable_income  NUMERIC(15,2) NOT NULL DEFAULT 0,
    pph21_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- Payment
    bank_name       VARCHAR(50),
    bank_account    VARCHAR(30),
    payment_status  VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    paid_at         TIMESTAMPTZ,

    -- Slip
    slip_number     VARCHAR(50),
    note            TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_payroll_entry UNIQUE (payroll_period_id, teacher_id),
    CONSTRAINT chk_entry_employee_type CHECK (employee_type IN ('pns', 'p3k', 'honorer', 'yayasan', 'kontrak')),
    CONSTRAINT chk_payment_status CHECK (payment_status IN ('unpaid', 'processing', 'paid', 'failed')),
    CONSTRAINT chk_total_earnings CHECK (total_earnings >= 0),
    CONSTRAINT chk_total_deductions_entry CHECK (total_deductions >= 0),
    CONSTRAINT chk_taxable_income CHECK (taxable_income >= 0),
    CONSTRAINT chk_pph21 CHECK (pph21_amount >= 0),
    CONSTRAINT chk_days_present CHECK (days_present >= 0),
    CONSTRAINT chk_days_absent CHECK (days_absent >= 0),
    CONSTRAINT chk_working_days CHECK (working_days >= 0)
);

-- Indexes
CREATE INDEX idx_payroll_entry_tenant_company ON payroll_entries (tenant_id, company_id);
CREATE INDEX idx_payroll_entry_period ON payroll_entries (payroll_period_id);
CREATE INDEX idx_payroll_entry_teacher ON payroll_entries (teacher_id);
CREATE INDEX idx_payroll_entry_type ON payroll_entries (employee_type);
CREATE INDEX idx_payroll_entry_payment ON payroll_entries (payment_status);
CREATE INDEX idx_payroll_entry_unpaid ON payroll_entries (payment_status) WHERE payment_status = 'unpaid';
CREATE INDEX idx_payroll_entry_rels ON payroll_entries USING GIN (_rels);
CREATE INDEX idx_payroll_entry_data ON payroll_entries USING GIN (_data);

-- Detail komponen per slip gaji
CREATE TABLE payroll_components (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    payroll_entry_id UUID NOT NULL,
    payroll_config_id UUID NOT NULL,

    -- Komponen
    component_type  VARCHAR(20) NOT NULL,
    category        VARCHAR(30) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    sort_order      INT NOT NULL DEFAULT 0,

    -- Perhitungan
    calculation_method VARCHAR(20) NOT NULL,
    base_amount     NUMERIC(15,2),
    quantity        NUMERIC(10,2),
    rate            NUMERIC(15,2),
    amount          NUMERIC(15,2) NOT NULL DEFAULT 0,

    -- Tax
    is_taxable      BOOLEAN NOT NULL DEFAULT true,

    -- Keterangan
    note            TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_comp_type CHECK (component_type IN ('earning', 'deduction')),
    CONSTRAINT chk_comp_category CHECK (category IN (
        'gaji_pokok', 'tunjangan_sertifikasi', 'tunjangan_struktural',
        'tunjangan_fungsional', 'tunjangan_transport', 'tunjangan_makan',
        'insentif_kehadiran', 'honor_mengajar', 'lembur',
        'potongan_pph', 'potongan_bpjs_kesehatan', 'potongan_bpjs_ketenagakerjaan',
        'potongan_ketidakhadiran', 'potongan_pinjaman', 'potongan_lain'
    )),
    CONSTRAINT chk_comp_calc CHECK (calculation_method IN ('fixed', 'percentage', 'per_hour', 'per_day', 'formula')),
    CONSTRAINT chk_comp_amount CHECK (amount >= 0)
);

-- Indexes
CREATE INDEX idx_payroll_comp_tenant_company ON payroll_components (tenant_id, company_id);
CREATE INDEX idx_payroll_comp_entry ON payroll_components (payroll_entry_id);
CREATE INDEX idx_payroll_comp_config ON payroll_components (payroll_config_id);
CREATE INDEX idx_payroll_comp_type ON payroll_components (component_type);
CREATE INDEX idx_payroll_comp_rels ON payroll_components USING GIN (_rels);
CREATE INDEX idx_payroll_comp_data ON payroll_components USING GIN (_data);
```

### Field Design Rationale

**payroll_configs:**

| Field | Keputusan | Alasan |
|---|---|---|
| `component_type` | `earning` / `deduction` | Binary classification — pendapatan atau potongan |
| `category` | 15 kategori | Mencakup semua komponen gaji PNS, honorer, dan yayasan |
| `calculation_method` | 5 metode | `fixed` (nominal tetap), `percentage` (% dari base), `per_hour` (honor mengajar), `per_day` (insentif kehadiran), `formula` (custom) |
| `applicable_to` | 4 tipe | Tunjangan sertifikasi hanya PNS, honor mengajar hanya honorer, dll |
| `percentage_base` | VARCHAR(50) | Referensi field base: `gaji_pokok`, `total_earnings`, dll |

**payroll_periods:**

| Field | Keputusan | Alasan |
|---|---|---|
| `status` | 8 status | Full lifecycle: draft -> calculating -> calculated -> reviewed -> approved -> disbursing -> disbursed |
| `total_gross` / `total_net` | Denormalized totals | Summary cepat untuk dashboard — tidak perlu SUM dari entries |
| `calculated_by` / `approved_by` / `disbursed_by` | UUID per step | Audit trail lengkap — siapa melakukan apa di setiap tahap |

**payroll_entries:**

| Field | Keputusan | Alasan |
|---|---|---|
| `employee_type` | Snapshot, bukan FK | Di-snapshot saat payroll dihitung — jika guru pindah status, payroll historis tidak berubah |
| `golongan` | VARCHAR(10), nullable | Golongan PNS (III/a, IV/b, dll) — nullable karena honorer tidak punya golongan |
| `days_present` / `days_absent` | Snapshot dari S026 | Diambil dari teacher_attendances saat perhitungan — basis potongan ketidakhadiran |
| `teaching_hours_per_week` | Snapshot dari S027 | Beban mengajar — basis honor per jam untuk guru honorer |
| `taxable_income` | NUMERIC(15,2) | Penghasilan kena pajak setelah dikurangi PTKP — basis PPh 21 |
| `bank_account` | VARCHAR(30) | Nomor rekening untuk disbursement — di entry, bukan di teacher, karena bisa berubah |

**payroll_components:**

| Field | Keputusan | Alasan |
|---|---|---|
| `base_amount` | NUMERIC(15,2) | Basis perhitungan: gaji pokok untuk percentage, jam mengajar untuk per_hour |
| `quantity` | NUMERIC(10,2) | Jumlah: hari hadir untuk insentif, jam mengajar untuk honor |
| `rate` | NUMERIC(15,2) | Tarif: Rp per jam, Rp per hari, atau percentage |
| `amount` | NUMERIC(15,2) | Hasil akhir: `base_amount * percentage / 100` atau `quantity * rate` |

### Payroll Calculation Workflow

```
1. SETUP (status: draft)
   - Create payroll_period untuk bulan/tahun
   - Pastikan payroll_configs sudah lengkap

2. CALCULATE (status: calculating -> calculated)
   - Untuk setiap guru aktif:
     a. Ambil attendance summary dari S026 (hari hadir, absen, sakit, cuti, terlambat)
     b. Ambil workload dari S027 (jam mengajar per minggu)
     c. Tentukan komponen yang applicable berdasarkan employee_type
     d. Hitung setiap komponen:
        - Gaji pokok: fixed amount
        - Tunjangan sertifikasi: 1x gaji pokok (PNS bersertifikat, jam >= 24)
        - Honor mengajar: teaching_hours * rate_per_hour (honorer)
        - Insentif kehadiran: days_present * rate_per_day
        - Potongan ketidakhadiran: days_absent * rate_per_day
        - Potongan keterlambatan: late_minutes * rate_per_minute (opsional)
        - BPJS Kesehatan: 1% dari gaji (ditanggung karyawan)
        - BPJS Ketenagakerjaan: 2% JHT + 1% JP (ditanggung karyawan)
     e. Hitung PPh 21 (simplified):
        - Penghasilan bruto = total_earnings
        - Biaya jabatan = 5% * bruto (max Rp 500.000/bulan)
        - Penghasilan neto = bruto - biaya jabatan - BPJS karyawan
        - Penghasilan neto setahun = neto * 12
        - PTKP (sesuai status kawin + tanggungan)
        - PKP = neto setahun - PTKP
        - PPh 21 setahun = PKP * tarif progresif
        - PPh 21 sebulan = PPh 21 setahun / 12
     f. Net pay = total_earnings - total_deductions

3. REVIEW (status: reviewed)
   - Bendahara/admin review perhitungan
   - Koreksi manual jika ada anomali

4. APPROVE (status: approved)
   - Kepala sekolah / ketua yayasan approve payroll

5. DISBURSE (status: disbursing -> disbursed)
   - Transfer ke rekening masing-masing guru
   - Update payment_status per entry
```

### PPh 21 Calculation

```sql
-- Tarif progresif PPh 21 (UU HPP No. 7/2021)
-- PKP 0 - 60 juta: 5%
-- PKP 60 - 250 juta: 15%
-- PKP 250 - 500 juta: 25%
-- PKP 500 juta - 5 miliar: 30%
-- PKP > 5 miliar: 35%

-- PTKP (Penghasilan Tidak Kena Pajak) per tahun:
-- TK/0 (belum kawin, tanpa tanggungan): Rp 54.000.000
-- K/0 (kawin, tanpa tanggungan): Rp 58.500.000
-- K/1 (kawin, 1 tanggungan): Rp 63.000.000
-- K/2 (kawin, 2 tanggungan): Rp 67.500.000
-- K/3 (kawin, 3 tanggungan): Rp 72.000.000
```

### Vernon Relationships

**payroll_configs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| — | — | — | Standalone config, no FK relations |

**payroll_periods:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| — | — | — | Standalone period, no FK relations |

**payroll_entries:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `payroll_period` | belongs_to | **Ya** | Periode bulan/tahun |
| `teacher` | belongs_to | **Ya** | Nama guru, NIP, employee_type |

**payroll_components:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `payroll_entry` | belongs_to | **Ya** | Parent slip gaji |
| `payroll_config` | belongs_to | **Ya** | Konfigurasi komponen |

### _rels / _data Structure

```json
// payroll_entries
{
  "_rels": {
    "payroll_period_id": "018f...",
    "teacher_id": "018f..."
  },
  "_data": {
    "payroll_period": {
      "id": "018f...",
      "period_name": "April 2026",
      "year": 2026,
      "month": 4,
      "status": "approved"
    },
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns",
      "golongan": "III/c",
      "role": "guru_mapel"
    }
  }
}

// payroll_components
{
  "_rels": {
    "payroll_entry_id": "018f...",
    "payroll_config_id": "018f..."
  },
  "_data": {
    "payroll_entry": {
      "id": "018f...",
      "teacher_name": "Bu Siti Aminah",
      "period_name": "April 2026"
    },
    "payroll_config": {
      "id": "018f...",
      "code": "tunjangan_sertifikasi",
      "name": "Tunjangan Profesi Guru",
      "component_type": "earning",
      "calculation_method": "percentage"
    }
  }
}
```

### Integration with S026 (Attendance) & S027 (Workload)

```sql
-- Step 1: Ambil attendance summary dari S026
SELECT
    teacher_id,
    COUNT(*) FILTER (WHERE status = 'present')   AS days_present,
    COUNT(*) FILTER (WHERE status = 'absent')     AS days_absent,
    COUNT(*) FILTER (WHERE status = 'sick')       AS days_sick,
    COUNT(*) FILTER (WHERE status = 'cuti')       AS days_leave,
    SUM(late_minutes)                              AS total_late_minutes,
    COUNT(*) FILTER (WHERE status IN ('present','sick','permitted','dinas_luar','cuti','libur')) AS working_days
FROM teacher_attendances
WHERE attendance_date >= $period_start AND attendance_date <= $period_end
  AND tenant_id = $tenant_id AND company_id = $company_id
GROUP BY teacher_id;

-- Step 2: Ambil workload summary dari S027
SELECT
    teacher_id,
    teaching_hours,
    total_hours
FROM teacher_workloads
WHERE academic_year_id = $academic_year_id
  AND semester = $semester
  AND tenant_id = $tenant_id AND company_id = $company_id;

-- Step 3: Potongan ketidakhadiran
-- Guru honorer: potong per hari absen
-- Formula: days_absent * (gaji_pokok / working_days_in_month)

-- Step 4: Tunjangan sertifikasi check
-- Hanya diberikan jika:
-- a. Guru sudah bersertifikat (teacher.is_certified = true)
-- b. Beban mengajar >= 24 JP/minggu (dari S027)
-- c. Kehadiran minimal (misal: max 3 hari absen per bulan)
```

### API Endpoints

```
# Payroll Configs
GET    /api/v1/payroll-configs                            -- List komponen gaji
POST   /api/v1/payroll-configs                            -- Tambah komponen
PUT    /api/v1/payroll-configs/{id}                       -- Update komponen
DELETE /api/v1/payroll-configs/{id}                       -- Soft delete

# Payroll Periods
GET    /api/v1/payroll-periods                            -- List periode
POST   /api/v1/payroll-periods                            -- Buat periode baru
GET    /api/v1/payroll-periods/{id}                       -- Detail periode + summary
DELETE /api/v1/payroll-periods/{id}                       -- Cancel periode (only draft)

# Payroll Lifecycle
POST   /api/v1/payroll-periods/{id}/calculate             -- Hitung payroll seluruh guru
POST   /api/v1/payroll-periods/{id}/review                -- Mark as reviewed
POST   /api/v1/payroll-periods/{id}/approve               -- Approve payroll
POST   /api/v1/payroll-periods/{id}/disburse              -- Proses pembayaran

# Payroll Entries (Slip Gaji)
GET    /api/v1/payroll-periods/{id}/entries                -- List slip gaji per periode
GET    /api/v1/payroll-entries/{id}                        -- Detail slip gaji + komponen
PUT    /api/v1/payroll-entries/{id}/adjust                 -- Manual adjustment komponen
GET    /api/v1/teachers/{id}/payslips                      -- Riwayat slip gaji guru
GET    /api/v1/payroll-entries/{id}/print                  -- Generate PDF slip gaji

# Reports
GET    /api/v1/payroll-periods/{id}/summary                -- Rekap total per komponen
GET    /api/v1/payroll-reports/annual?year=2026            -- Laporan tahunan (untuk SPT)
GET    /api/v1/payroll-reports/pph21?year=2026&month=4     -- Laporan PPh 21
GET    /api/v1/payroll-reports/bpjs?year=2026&month=4      -- Laporan BPJS
```

## Consequences

### Positive

- **Multi-status support**: PNS, honorer, yayasan, kontrak — masing-masing punya komponen dan aturan berbeda.
- **Full workflow**: Draft -> calculate -> review -> approve -> disburse — dengan audit trail di setiap tahap.
- **S026/S027 integration**: Potongan ketidakhadiran otomatis dari attendance, honor per jam dari workload.
- **PPh 21 compliant**: Perhitungan pajak sesuai UU HPP dengan tarif progresif dan PTKP.
- **Payslip generation**: Slip gaji per guru dengan breakdown komponen lengkap.
- **Immutable history**: Payroll entry di-snapshot — perubahan status guru tidak mengubah payroll historis.
- **Pesantren compatible**: Komponen bisa dikonfigurasi untuk honorarium ustadz/ustadzah.

### Negative / Trade-offs

- **PNS gaji pokok di luar sistem**: Gaji pokok PNS dari APBN/APBD — SekolahPro hanya mengelola tunjangan dan honor dari sekolah.
- **PPh 21 simplified**: Perhitungan pajak di-simplify — untuk kasus kompleks (PPh 21 final, penghasilan tidak teratur) perlu enhancement.
- **Manual adjustment**: Koreksi manual bisa override calculation — perlu audit trail ketat.
- **Bank transfer integration**: Disbursement awal manual — integrasi bank API bisa ditambah nanti.
- **Regulasi berubah**: Tarif PPh, PTKP, aturan BOS berubah periodik — perlu maintenance konfigurasi.

## Alternatives Considered

### 1. Payroll tanpa detail komponen (hanya total)
- Ditolak: guru dan kepsek perlu melihat breakdown. Audit pajak memerlukan detail per komponen.

### 2. Payroll real-time (hitung setiap hari)
- Ditolak: payroll bersifat bulanan. Menghitung setiap hari waste resources dan membingungkan (angka berubah terus).

### 3. Integrasi langsung ke sistem payroll eksternal
- Deferred: MVP mengelola payroll internal. Export ke format BPJS/pajak bisa ditambah. Integrasi API bank untuk disbursement di fase berikutnya.

### 4. Flat table tanpa komponen terpisah
- Ditolak: tidak bisa menambah komponen baru tanpa alter table. Komponen bervariasi antar sekolah — perlu configurable.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `PayrollConfigDescriptor.TableName()` | -- | `"payroll_configs"` |
| U02 | `PayrollPeriodDescriptor.TableName()` | -- | `"payroll_periods"` |
| U03 | `PayrollEntryDescriptor.TableName()` | -- | `"payroll_entries"` |
| U04 | `PayrollComponentDescriptor.TableName()` | -- | `"payroll_components"` |
| U05 | Validate rejects invalid `component_type` | `"bonus"` | Error |
| U06 | Validate rejects invalid `period.status` | `"pending"` | Error |
| U07 | Validate rejects invalid `employee_type` | `"freelance"` | Error |
| U08 | Calculate fixed earning | method=fixed, amount=5000000 | amount=5000000 |
| U09 | Calculate percentage earning | method=percentage, base=5000000, pct=100 | amount=5000000 |
| U10 | Calculate per_hour honor | method=per_hour, hours=24, rate=50000 | amount=1200000 |
| U11 | Calculate per_day insentif | method=per_day, days=22, rate=25000 | amount=550000 |
| U12 | Calculate absence deduction | absent=3, daily_rate=250000 | deduction=750000 |
| U13 | PPh 21: below PTKP | annual_net=48000000, PTKP=54000000 | pph21=0 |
| U14 | PPh 21: first bracket | PKP=50000000 | pph21=2500000/year, 208333/month |
| U15 | PPh 21: second bracket | PKP=100000000 | pph21=(60M*5%)+(40M*15%)=9000000/year |
| U16 | Tunjangan sertifikasi: eligible | certified=true, hours>=24, attendance OK | Amount = 1x gaji pokok |
| U17 | Tunjangan sertifikasi: ineligible (hours) | certified=true, hours=20 | Amount = 0 |
| U18 | Net pay calculation | earnings=8000000, deductions=1500000 | net=6500000 |

### Integration Tests -- Payroll Lifecycle

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create payroll period | POST /payroll-periods | 201, status=draft |
| I02 | Calculate payroll | POST /calculate | 200, status=calculated, entries created |
| I03 | Entries created for all active teachers | After calculate | One entry per active teacher |
| I04 | Components generated per entry | After calculate | Components matching applicable configs |
| I05 | Attendance data integrated | After calculate | days_present/absent from S026 |
| I06 | Workload data integrated | After calculate | teaching_hours from S027 |
| I07 | Review payroll | POST /review | 200, status=reviewed |
| I08 | Approve payroll | POST /approve | 200, status=approved |
| I09 | Disburse payroll | POST /disburse | 200, status=disbursed |
| I10 | Cannot approve without review | POST /approve on calculated | 422, must review first |
| I11 | Cannot modify disbursed payroll | PUT entry on disbursed period | 422, immutable |

### Integration Tests -- Payslip

| # | Test Case | Action | Expected |
|---|---|---|---|
| I12 | Get payslip detail | GET /payroll-entries/{id} | 200, entry + all components |
| I13 | Get teacher payslip history | GET /teachers/{id}/payslips | 200, list of past entries |
| I14 | Manual adjustment | PUT /adjust with override | 200, component amount updated, totals recalculated |
| I15 | Generate PDF slip | GET /payroll-entries/{id}/print | 200, PDF binary |

### Integration Tests -- Tax & Reports

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | PPh 21 monthly report | GET /payroll-reports/pph21 | 200, all teachers with pph21 |
| I17 | Annual summary for SPT | GET /payroll-reports/annual | 200, 12-month summary per teacher |
| I18 | BPJS report | GET /payroll-reports/bpjs | 200, BPJS contributions |

### Integration Tests -- SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | TeacherUpdated syncs to entries | Update teacher full_name | `_data.teacher.full_name` updated |
| I20 | PeriodStatusChanged syncs to entries | Period status change | `_data.payroll_period.status` updated |
| I21 | ConfigUpdated syncs to components | Update config name | `_data.payroll_config.name` updated |

### Integration Tests -- Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's payroll | GET with wrong tenant | 404 |
| I23 | Cannot calculate other company's payroll | POST /calculate with wrong company | Error, scope violation |
| I24 | Cannot view other tenant's payslip | GET payslip with wrong tenant | 404 |

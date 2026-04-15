# ADR-S030: Leave Management (Manajemen Cuti & Izin Guru/Staff)

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Manajemen cuti dan izin guru/staff di sekolah Indonesia memiliki regulasi yang ketat, terutama untuk guru berstatus PNS/ASN:

1. **PP No. 11 Tahun 2017** tentang Manajemen PNS: Cuti tahunan 12 hari kerja, cuti besar 3 bulan setelah 6 tahun masa kerja berturut-turut, cuti sakit sesuai keterangan dokter, cuti melahirkan 3 bulan.
2. **Permenpan No. 24 Tahun 2014**: Aturan teknis pelaksanaan cuti PNS — termasuk mekanisme persetujuan berjenjang.
3. **Guru Honorer/Yayasan**: Aturan cuti ditentukan oleh pihak yayasan atau sekolah — lebih fleksibel tapi perlu tetap tercatat.
4. **Dampak operasional**: Guru yang cuti harus digantikan (S032 Teacher Substitution), absensi otomatis terupdate (S026), dan payroll terpengaruh (S031).
5. **Pesantren**: Ustadz/ustadzah mengikuti aturan yayasan pondok, dengan tambahan periode cuti khusus (misal: setelah Ramadhan).

Jenis cuti yang harus didukung:

| Jenis Cuti | Durasi PNS | Durasi Non-PNS | Approval Level |
|------------|-----------|----------------|----------------|
| Cuti Tahunan | Max 12 hari/tahun | Sesuai kebijakan yayasan | Kepala Sekolah |
| Cuti Sakit | Sesuai surat dokter (max 18 bulan) | Sesuai surat dokter | Kepala Sekolah |
| Cuti Melahirkan | 3 bulan | Sesuai kebijakan | Kepala Sekolah |
| Cuti Besar | 3 bulan setelah 6 tahun | Tidak ada | Kepsek + Yayasan |
| Izin | 1-3 hari (potong cuti tahunan atau tanpa potongan) | Fleksibel | Kepala Sekolah |
| Tugas Belajar | 6 bulan - 4 tahun | Sesuai kebijakan | Kepsek + Yayasan + Dinas |

### Mengapa Vernon Pattern?

- has_many dari teacher — banyak record cuti per guru per tahun.
- Relasi ke teacher (ADR-012), academic_year (ADR-010), approver (ADR-012).
- Read-heavy: dashboard kepsek, saldo cuti, laporan absensi bulanan.
- Write moderate: pengajuan cuti tidak setiap hari.
- Eventually consistent acceptable — approval workflow bersifat async.

## Decision

Menggunakan **Vernon Pattern** untuk 3 tabel: `leave_types` (konfigurasi jenis cuti per sekolah), `leave_balances` (saldo cuti per guru per tahun), dan `leave_requests` (pengajuan cuti).

### Table Schema

```sql
-- Konfigurasi jenis cuti per sekolah
CREATE TABLE leave_types (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    code            VARCHAR(30) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,

    -- Aturan
    max_days_per_year INT,
    requires_attachment BOOLEAN NOT NULL DEFAULT false,
    deducts_balance BOOLEAN NOT NULL DEFAULT true,
    applicable_to   VARCHAR(20) NOT NULL DEFAULT 'all',
    min_service_years INT,

    -- Approval
    approval_levels INT NOT NULL DEFAULT 1,
    approval_chain  JSONB NOT NULL DEFAULT '[]',

    is_active       BOOLEAN NOT NULL DEFAULT true,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_leave_type_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_leave_type_code CHECK (code IN (
        'cuti_tahunan', 'cuti_sakit', 'cuti_melahirkan',
        'cuti_besar', 'izin', 'tugas_belajar'
    )),
    CONSTRAINT chk_applicable_to CHECK (applicable_to IN ('all', 'pns', 'honorer', 'yayasan')),
    CONSTRAINT chk_approval_levels CHECK (approval_levels >= 1 AND approval_levels <= 4),
    CONSTRAINT chk_max_days CHECK (max_days_per_year IS NULL OR max_days_per_year > 0),
    CONSTRAINT chk_min_service CHECK (min_service_years IS NULL OR min_service_years >= 0)
);

-- Indexes
CREATE INDEX idx_leave_type_tenant_company ON leave_types (tenant_id, company_id);
CREATE INDEX idx_leave_type_code ON leave_types (code);
CREATE INDEX idx_leave_type_active ON leave_types (is_active) WHERE is_active = true;
CREATE INDEX idx_leave_type_rels ON leave_types USING GIN (_rels);
CREATE INDEX idx_leave_type_data ON leave_types USING GIN (_data);

-- Saldo cuti per guru per tahun
CREATE TABLE leave_balances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    leave_type_id   UUID NOT NULL,
    academic_year_id UUID NOT NULL,

    -- Saldo
    year            INT NOT NULL,
    initial_balance INT NOT NULL DEFAULT 0,
    used            INT NOT NULL DEFAULT 0,
    remaining       INT NOT NULL DEFAULT 0,
    carried_over    INT NOT NULL DEFAULT 0,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_leave_balance UNIQUE (teacher_id, leave_type_id, year),
    CONSTRAINT chk_balance_initial CHECK (initial_balance >= 0),
    CONSTRAINT chk_balance_used CHECK (used >= 0),
    CONSTRAINT chk_balance_remaining CHECK (remaining >= 0),
    CONSTRAINT chk_balance_carried CHECK (carried_over >= 0),
    CONSTRAINT chk_balance_year CHECK (year >= 2020 AND year <= 2100)
);

-- Indexes
CREATE INDEX idx_leave_bal_tenant_company ON leave_balances (tenant_id, company_id);
CREATE INDEX idx_leave_bal_teacher ON leave_balances (teacher_id);
CREATE INDEX idx_leave_bal_type ON leave_balances (leave_type_id);
CREATE INDEX idx_leave_bal_year ON leave_balances (year);
CREATE INDEX idx_leave_bal_teacher_year ON leave_balances (teacher_id, year);
CREATE INDEX idx_leave_bal_rels ON leave_balances USING GIN (_rels);
CREATE INDEX idx_leave_bal_data ON leave_balances USING GIN (_data);

-- Pengajuan cuti
CREATE TABLE leave_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    teacher_id      UUID NOT NULL,
    leave_type_id   UUID NOT NULL,
    leave_balance_id UUID,
    academic_year_id UUID NOT NULL,

    -- Periode cuti
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_days      INT NOT NULL,

    -- Detail
    reason          TEXT NOT NULL,
    attachment_url  TEXT,
    emergency_contact VARCHAR(100),
    emergency_phone VARCHAR(20),

    -- Alamat selama cuti
    address_during_leave TEXT,

    -- Status approval
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    current_approval_level INT NOT NULL DEFAULT 0,

    -- Approval history (denormalized for quick access)
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    rejected_by     UUID,
    rejected_at     TIMESTAMPTZ,
    rejection_reason TEXT,

    -- Cancellation
    cancelled_by    UUID,
    cancelled_at    TIMESTAMPTZ,
    cancellation_reason TEXT,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_leave_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_leave_total_days CHECK (total_days > 0),
    CONSTRAINT chk_leave_status CHECK (status IN (
        'draft', 'submitted', 'pending_approval',
        'approved', 'rejected', 'cancelled', 'completed'
    )),
    CONSTRAINT chk_approval_level CHECK (current_approval_level >= 0 AND current_approval_level <= 4)
);

-- Indexes
CREATE INDEX idx_leave_req_tenant_company ON leave_requests (tenant_id, company_id);
CREATE INDEX idx_leave_req_teacher ON leave_requests (teacher_id);
CREATE INDEX idx_leave_req_type ON leave_requests (leave_type_id);
CREATE INDEX idx_leave_req_balance ON leave_requests (leave_balance_id);
CREATE INDEX idx_leave_req_status ON leave_requests (status);
CREATE INDEX idx_leave_req_dates ON leave_requests (start_date, end_date);
CREATE INDEX idx_leave_req_teacher_status ON leave_requests (teacher_id, status);
CREATE INDEX idx_leave_req_pending ON leave_requests (status) WHERE status IN ('submitted', 'pending_approval');
CREATE INDEX idx_leave_req_year ON leave_requests (academic_year_id);
CREATE INDEX idx_leave_req_rels ON leave_requests USING GIN (_rels);
CREATE INDEX idx_leave_req_data ON leave_requests USING GIN (_data);

-- Riwayat approval per level
CREATE TABLE leave_approval_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    leave_request_id UUID NOT NULL,
    approver_id     UUID NOT NULL,

    -- Approval detail
    approval_level  INT NOT NULL,
    approver_role   VARCHAR(30) NOT NULL,
    action          VARCHAR(20) NOT NULL,
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

    CONSTRAINT chk_approval_action CHECK (action IN ('approved', 'rejected', 'returned')),
    CONSTRAINT chk_approver_role CHECK (approver_role IN (
        'wakil_kepsek', 'kepala_sekolah', 'yayasan', 'dinas_pendidikan'
    )),
    CONSTRAINT chk_log_level CHECK (approval_level >= 1 AND approval_level <= 4)
);

-- Indexes
CREATE INDEX idx_leave_log_tenant_company ON leave_approval_logs (tenant_id, company_id);
CREATE INDEX idx_leave_log_request ON leave_approval_logs (leave_request_id);
CREATE INDEX idx_leave_log_approver ON leave_approval_logs (approver_id);
CREATE INDEX idx_leave_log_rels ON leave_approval_logs USING GIN (_rels);
CREATE INDEX idx_leave_log_data ON leave_approval_logs USING GIN (_data);
```

### Field Design Rationale

**leave_types:**

| Field | Keputusan | Alasan |
|---|---|---|
| `code` | VARCHAR(30), 6 kode standar | Enum cuti sesuai regulasi PNS + tambahan `izin` dan `tugas_belajar` |
| `max_days_per_year` | INT, nullable | NULL untuk cuti tanpa batas tahunan (cuti sakit = sesuai surat dokter) |
| `requires_attachment` | BOOLEAN | Cuti sakit wajib lampiran surat dokter, cuti tahunan tidak |
| `deducts_balance` | BOOLEAN | Cuti sakit tidak potong saldo cuti tahunan |
| `applicable_to` | VARCHAR(20) | Cuti besar hanya untuk PNS, cuti tahunan untuk semua |
| `min_service_years` | INT, nullable | Cuti besar memerlukan 6 tahun masa kerja |
| `approval_chain` | JSONB | Array role yang harus approve: `["kepala_sekolah"]` atau `["wakil_kepsek","kepala_sekolah","yayasan"]` |

**leave_balances:**

| Field | Keputusan | Alasan |
|---|---|---|
| `year` | INT | Saldo per tahun kalender (bukan tahun ajaran) — sesuai aturan PNS |
| `initial_balance` | INT | Saldo awal: 12 untuk PNS cuti tahunan, configurable untuk lainnya |
| `used` | INT | Jumlah hari sudah terpakai — auto-increment saat cuti diapprove |
| `remaining` | INT | `initial_balance + carried_over - used` — denormalisasi untuk query cepat |
| `carried_over` | INT | Sisa tahun lalu yang dibawa — PNS: max 6 hari carry over |

**leave_requests:**

| Field | Keputusan | Alasan |
|---|---|---|
| `total_days` | INT, calculated | Jumlah hari kerja (exclude weekend/libur) antara start-end date |
| `status` | VARCHAR(20), 7 status | Full lifecycle: draft -> submitted -> pending -> approved/rejected -> completed |
| `current_approval_level` | INT | Track posisi di approval chain — level berapa yang sedang pending |
| `emergency_contact` | VARCHAR(100) | Wajib untuk cuti panjang (cuti besar, melahirkan) — kontak darurat selama cuti |
| `address_during_leave` | TEXT | Alamat selama cuti — requirement dari form cuti PNS |

### Approval Workflow

```
Cuti biasa (tahunan, sakit, izin):
  Guru → [submit] → Kepala Sekolah → [approve/reject]

Cuti besar / tugas belajar:
  Guru → [submit] → Wakil Kepsek → [approve] → Kepala Sekolah → [approve] → Yayasan → [approve/reject]

State transitions:
  draft → submitted → pending_approval → approved → completed
                                       → rejected
  draft → cancelled (by teacher)
  submitted → cancelled (by teacher sebelum approval)
  approved → cancelled (pembatalan cuti yang sudah disetujui — refund saldo)
```

### Vernon Relationships

**leave_types:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| — | — | — | Standalone config, no FK relations |

**leave_balances:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Nama guru, NIP, employee_type untuk validasi aturan PNS |
| `leave_type` | belongs_to | **Ya** | Jenis cuti, max days, aturan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |

**leave_requests:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `teacher` | belongs_to | **Ya** | Pemohon cuti |
| `leave_type` | belongs_to | **Ya** | Jenis cuti, aturan approval |
| `leave_balance` | belongs_to | **Ya** | Saldo terkait — untuk cek remaining |
| `academic_year` | belongs_to | **Ya** | Tahun ajaran |

**leave_approval_logs:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `leave_request` | belongs_to | **Ya** | Parent request |
| `approver` | belongs_to | **Ya** | Siapa yang approve/reject |

### _rels / _data Structure

```json
// leave_balances
{
  "_rels": {
    "teacher_id": "018f...",
    "leave_type_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns"
    },
    "leave_type": {
      "id": "018f...",
      "code": "cuti_tahunan",
      "name": "Cuti Tahunan",
      "max_days_per_year": 12
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// leave_requests
{
  "_rels": {
    "teacher_id": "018f...",
    "leave_type_id": "018f...",
    "leave_balance_id": "018f...",
    "academic_year_id": "018f..."
  },
  "_data": {
    "teacher": {
      "id": "018f...",
      "full_name": "Bu Siti Aminah",
      "nip": "198501012010012001",
      "employee_type": "pns",
      "role": "guru_mapel"
    },
    "leave_type": {
      "id": "018f...",
      "code": "cuti_tahunan",
      "name": "Cuti Tahunan",
      "max_days_per_year": 12,
      "approval_chain": ["kepala_sekolah"]
    },
    "leave_balance": {
      "id": "018f...",
      "initial_balance": 12,
      "used": 3,
      "remaining": 9
    },
    "academic_year": {
      "id": "018f...",
      "name": "2025/2026"
    }
  }
}

// leave_approval_logs
{
  "_rels": {
    "leave_request_id": "018f...",
    "approver_id": "018f..."
  },
  "_data": {
    "leave_request": {
      "id": "018f...",
      "teacher_name": "Bu Siti Aminah",
      "leave_type": "Cuti Tahunan",
      "start_date": "2026-04-20",
      "end_date": "2026-04-22",
      "total_days": 3
    },
    "approver": {
      "id": "018f...",
      "full_name": "Pak Ahmad",
      "role": "kepala_sekolah"
    }
  }
}
```

### Integration with S026 (Teacher Attendance)

Saat leave_request disetujui (status = `approved`):

```sql
-- Auto-create attendance records dengan status 'cuti' untuk setiap hari cuti
INSERT INTO teacher_attendances (
    tenant_id, company_id, teacher_id, academic_year_id,
    attendance_date, status, note
)
SELECT
    lr.tenant_id, lr.company_id, lr.teacher_id, lr.academic_year_id,
    d::date, 'cuti',
    'Auto-generated dari cuti: ' || lt.name
FROM leave_requests lr
JOIN leave_types lt ON lt.id = lr.leave_type_id
CROSS JOIN generate_series(lr.start_date, lr.end_date, '1 day'::interval) d
WHERE lr.id = $1
  AND EXTRACT(DOW FROM d::date) NOT IN (0, 6) -- exclude weekend
  AND NOT EXISTS (
      SELECT 1 FROM teacher_attendances ta
      WHERE ta.teacher_id = lr.teacher_id AND ta.attendance_date = d::date
  );
```

### Balance Deduction Logic

```sql
-- Saat cuti diapprove dan leave_type.deducts_balance = true
UPDATE leave_balances SET
    used = used + $total_days,
    remaining = initial_balance + carried_over - (used + $total_days),
    updated_at = now()
WHERE id = $leave_balance_id
  AND remaining >= $total_days; -- prevent negative balance

-- Jika cancelled setelah diapprove — refund balance
UPDATE leave_balances SET
    used = used - $total_days,
    remaining = initial_balance + carried_over - (used - $total_days),
    updated_at = now()
WHERE id = $leave_balance_id;
```

### API Endpoints

```
# Leave Types (Config)
GET    /api/v1/leave-types                               -- List jenis cuti aktif
POST   /api/v1/leave-types                               -- Create jenis cuti
PUT    /api/v1/leave-types/{id}                           -- Update konfigurasi
DELETE /api/v1/leave-types/{id}                           -- Soft delete

# Leave Balances
GET    /api/v1/teachers/{id}/leave-balances               -- Saldo cuti guru per tahun
GET    /api/v1/teachers/{id}/leave-balances?year=2026     -- Filter per tahun
POST   /api/v1/leave-balances/initialize                  -- Init saldo awal tahun (bulk all teachers)
PUT    /api/v1/leave-balances/{id}/adjust                 -- Manual adjustment oleh admin

# Leave Requests
POST   /api/v1/leave-requests                             -- Ajukan cuti baru
GET    /api/v1/leave-requests/{id}                        -- Detail pengajuan
PUT    /api/v1/leave-requests/{id}                        -- Edit draft
DELETE /api/v1/leave-requests/{id}                        -- Cancel request
GET    /api/v1/teachers/{id}/leave-requests               -- Riwayat cuti guru
GET    /api/v1/leave-requests/pending                     -- Daftar cuti pending approval (untuk kepsek)

# Approval
POST   /api/v1/leave-requests/{id}/submit                 -- Submit untuk approval
POST   /api/v1/leave-requests/{id}/approve                -- Approve oleh approver
POST   /api/v1/leave-requests/{id}/reject                 -- Reject oleh approver
POST   /api/v1/leave-requests/{id}/cancel                 -- Batalkan cuti yang sudah disetujui

# Reports
GET    /api/v1/leave-requests/report?month=2026-04        -- Rekap cuti bulanan seluruh guru
GET    /api/v1/leave-requests/calendar                    -- Kalender cuti (siapa cuti kapan)
```

## Consequences

### Positive

- **Regulasi PNS compliant**: Mendukung aturan cuti PNS (12 hari tahunan, 3 bulan melahirkan, cuti besar 6 tahun).
- **Multi-level approval**: Workflow berjenjang sesuai jenis cuti — simple untuk cuti biasa, kompleks untuk cuti besar.
- **Auto-integration**: Approval otomatis membuat record attendance (S026) dan mengurangi saldo.
- **Balance tracking**: Saldo cuti real-time — guru dan kepsek bisa melihat remaining balance.
- **Audit trail**: Leave approval logs mencatat setiap langkah approval — siapa, kapan, keputusan apa.
- **Pesantren compatible**: Ustadz/ustadzah menggunakan tabel yang sama — terminologi di presentation layer.

### Negative / Trade-offs

- **Weekend/holiday exclusion**: Perhitungan `total_days` harus exclude weekend dan hari libur — perlu tabel kalender libur.
- **Carry-over manual**: Perpindahan saldo antar tahun perlu di-trigger manual (atau cron job awal tahun).
- **Cancellation refund**: Pembatalan cuti yang sudah disetujui perlu refund saldo — bisa jadi edge case jika sudah lewat tanggal.
- **Multi-approval complexity**: Workflow 3-4 level approval menambah kompleksitas state management.
- **PNS rules hardcoded**: Aturan 12 hari, 3 bulan, 6 tahun di-encode sebagai konfigurasi tapi tetap perlu maintenance jika regulasi berubah.

## Alternatives Considered

### 1. Cuti tanpa balance tracking (hanya log)
- Ditolak: tanpa saldo, tidak bisa enforce max 12 hari PNS. Kepsek tidak punya visibility sisa cuti guru.

### 2. Approval di luar sistem (manual surat)
- Ditolak: tidak bisa auto-integrate ke attendance dan payroll. Proses manual rawan error dan slow.

### 3. Saldo cuti sebagai field di tabel teachers
- Ditolak: per jenis cuti per tahun — perlu tabel terpisah. Tidak bisa track history.

### 4. Single table untuk semua (request + balance + approval log)
- Ditolak: melanggar SRP. Balance adalah state per tahun, request adalah event, approval log adalah audit trail.

## Test Cases

### Unit Tests

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `LeaveTypeDescriptor.TableName()` | -- | `"leave_types"` |
| U02 | `LeaveBalanceDescriptor.TableName()` | -- | `"leave_balances"` |
| U03 | `LeaveRequestDescriptor.TableName()` | -- | `"leave_requests"` |
| U04 | Validate rejects invalid `leave_type.code` | `"cuti_liburan"` | Error: invalid code |
| U05 | Validate rejects invalid `status` | `"pending"` | Error: invalid status |
| U06 | Validate rejects `end_date < start_date` | start=2026-04-25, end=2026-04-20 | Error: end_date must >= start_date |
| U07 | Validate rejects `total_days <= 0` | `0` | Error: total_days must > 0 |
| U08 | Calculate working days excludes weekend | Mon-Fri (5 hari kalender) | total_days = 5 |
| U09 | Calculate working days excludes weekend | Thu-Mon (4 hari kalender, 1 weekend) | total_days = 2 |
| U10 | Balance deduction: normal | remaining=9, request=3 | remaining=6, used+=3 |
| U11 | Balance deduction: insufficient | remaining=2, request=5 | Error: insufficient balance |
| U12 | PNS max 12 cuti tahunan | employee_type=pns | initial_balance=12 |
| U13 | Cuti besar requires 6 years | service_years=4, code=cuti_besar | Error: min 6 tahun masa kerja |
| U14 | Approval chain validation | approval_level > approval_levels | Error: all levels completed |
| U15 | Cancellation refunds balance | cancelled after approved, days=3 | remaining+=3, used-=3 |

### Integration Tests -- Leave Request Lifecycle

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create draft leave request | POST /leave-requests | 201, status=draft |
| I02 | Submit leave request | POST /leave-requests/{id}/submit | 200, status=submitted |
| I03 | Cannot submit without sufficient balance | Submit with remaining=0 | 422, insufficient balance |
| I04 | Approve by kepala sekolah | POST /approve by kepsek | 200, status=approved |
| I05 | Multi-level: approve level 1 | POST /approve by wakil_kepsek | 200, status=pending_approval, level=2 |
| I06 | Multi-level: approve level 2 | POST /approve by kepsek | 200, status=approved |
| I07 | Reject leave request | POST /reject with reason | 200, status=rejected |
| I08 | Cancel approved leave | POST /cancel | 200, status=cancelled, balance refunded |
| I09 | Cannot approve if not authorized | POST /approve by guru_mapel | 403, forbidden |
| I10 | Cannot edit non-draft request | PUT /leave-requests/{id} (status=submitted) | 422, cannot edit |

### Integration Tests -- Balance Management

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Initialize balances for new year | POST /leave-balances/initialize?year=2026 | 201, all teachers get initial balance |
| I12 | Auto-deduct on approval | Approve request with 3 days | Balance: used+=3, remaining-=3 |
| I13 | Refund on cancellation | Cancel approved request | Balance: used-=3, remaining+=3 |
| I14 | Manual adjustment by admin | PUT /adjust with +2 days | Balance adjusted |
| I15 | Prevent negative remaining | Approve request exceeding remaining | 422, insufficient balance |

### Integration Tests -- S026 Integration

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Approval creates attendance records | Approve 3-day leave | 3 attendance records with status='cuti' |
| I17 | Cancellation removes attendance | Cancel approved leave | Attendance records soft-deleted |
| I18 | Weekend excluded from attendance | Leave covers Mon-Sun | Only Mon-Fri attendance created |

### Integration Tests -- SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I19 | TeacherUpdated syncs to balances | Update teacher full_name | `_data.teacher.full_name` updated |
| I20 | TeacherUpdated syncs to requests | Update teacher full_name | `_data.teacher.full_name` updated |
| I21 | LeaveTypeUpdated syncs to requests | Update leave_type name | `_data.leave_type.name` updated |

### Integration Tests -- Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I22 | Cannot access other tenant's leave | GET with wrong tenant scope | 404 |
| I23 | Cannot approve other company's leave | POST /approve with wrong company | Error, scope violation |
| I24 | Cannot view other tenant's balance | GET balances with wrong tenant | 404 |

# ADR-S012: Student Discipline / Tata Tertib

**Status**: Approved
**Date**: 2026-04-15
**Deciders**: SekolahPro Engineering Team

## Context

Pencatatan pelanggaran dan penghargaan tata tertib siswa diperlukan untuk:

1. **Pembinaan**: Tracking perilaku siswa untuk intervensi dini.
2. **Poin pelanggaran**: Sistem poin kumulatif — akumulasi poin menentukan sanksi.
3. **Peringatan bertingkat**: Teguran lisan → tertulis → panggilan orang tua → skorsing → dikeluarkan.
4. **Rapor**: Catatan perilaku dan kedisiplinan masuk rapor.
5. **Reward**: Penghargaan untuk perilaku positif (poin positif).
6. **Audit**: Dokumentasi untuk transparansi keputusan disiplin.

Karakteristik:
- Setiap kejadian = satu record (event-based, bukan state-based).
- Poin bersifat **kumulatif** per semester/tahun ajaran.
- Ada dua arah: **violation** (pelanggaran, poin negatif) dan **merit** (penghargaan, poin positif).
- Eskalasi otomatis berdasarkan akumulasi poin.

### Mengapa Vernon Pattern?

- has_many dari student (banyak records per siswa).
- Read-heavy: dashboard siswa, laporan disiplin, cetak surat peringatan.
- Relasi ke student, academic_year.
- Business logic moderate: poin akumulasi, threshold eskalasi.

## Decision

Menggunakan **Vernon Pattern** untuk 2 tabel: `discipline_rules` (master aturan tata tertib) dan `student_disciplines` (catatan per kejadian).

### Table Schema

```sql
-- Master aturan tata tertib
CREATE TABLE discipline_rules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Identitas
    code            VARCHAR(20) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,

    -- Konfigurasi
    rule_type       VARCHAR(10) NOT NULL,
    category        VARCHAR(30) NOT NULL,
    default_points  INT NOT NULL,

    -- Status
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

    CONSTRAINT uq_discipline_rule_code UNIQUE (tenant_id, company_id, code),
    CONSTRAINT chk_rule_type CHECK (rule_type IN ('violation', 'merit')),
    CONSTRAINT chk_rule_category CHECK (category IN (
        'ringan', 'sedang', 'berat', 'sangat_berat',
        'akademik', 'sosial', 'kepemimpinan'
    ))
);

-- Indexes
CREATE INDEX idx_disc_rule_tenant_company ON discipline_rules (tenant_id, company_id);

-- Catatan disiplin per kejadian
CREATE TABLE student_disciplines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Foreign keys
    student_id      UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    rule_id         UUID NOT NULL,

    -- Kejadian
    incident_date   DATE NOT NULL,
    semester        VARCHAR(10) NOT NULL,
    record_type     VARCHAR(10) NOT NULL,
    points          INT NOT NULL,
    description     TEXT,

    -- Sanksi / penghargaan
    action_taken    VARCHAR(30),
    action_note     TEXT,

    -- Pencatat
    reported_by     UUID NOT NULL,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    -- Pemberitahuan orang tua
    parent_notified BOOLEAN NOT NULL DEFAULT false,
    notified_at     TIMESTAMPTZ,

    -- Vernon read-cache
    _rels           JSONB NOT NULL DEFAULT '{}',
    _data           JSONB NOT NULL DEFAULT '{}',
    _sync_status    TEXT NOT NULL DEFAULT 'synced',
    _sync_version   BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_disc_semester CHECK (semester IN ('ganjil', 'genap')),
    CONSTRAINT chk_disc_record_type CHECK (record_type IN ('violation', 'merit')),
    CONSTRAINT chk_disc_action CHECK (action_taken IS NULL OR action_taken IN (
        'teguran_lisan', 'teguran_tertulis', 'panggilan_ortu',
        'skorsing_1_hari', 'skorsing_3_hari', 'skorsing_1_minggu',
        'dikeluarkan',
        'pujian', 'sertifikat', 'hadiah'
    ))
);

-- Indexes
CREATE INDEX idx_disc_tenant_company ON student_disciplines (tenant_id, company_id);
CREATE INDEX idx_disc_student ON student_disciplines (student_id);
CREATE INDEX idx_disc_student_semester ON student_disciplines (student_id, academic_year_id, semester);
CREATE INDEX idx_disc_date ON student_disciplines (incident_date);
CREATE INDEX idx_disc_type ON student_disciplines (record_type);
CREATE INDEX idx_disc_rule ON student_disciplines (rule_id);
CREATE INDEX idx_disc_rels ON student_disciplines USING GIN (_rels);
CREATE INDEX idx_disc_data ON student_disciplines USING GIN (_data);
```

### Field Design Rationale

| Field | Keputusan | Alasan |
|---|---|---|
| `rule_type` | violation/merit | Dua arah: pelanggaran (negatif) dan penghargaan (positif) |
| `category` | 7 kategori | 4 tingkat pelanggaran (ringan → sangat berat) + 3 jenis penghargaan |
| `default_points` | INT di rule | Template poin — bisa di-override per kejadian |
| `points` | INT di discipline | Poin aktual per kejadian — positif untuk merit, negatif untuk violation |
| `action_taken` | VARCHAR(30), CHECK | Sanksi bertingkat sesuai kultur sekolah Indonesia |
| `parent_notified` | BOOLEAN | Tracking apakah orang tua sudah diberitahu |
| `reported_by` | UUID, NOT NULL | Guru/staff yang melaporkan |
| `approved_by` | UUID, nullable | Kepsek/wakasek yang menyetujui sanksi |

### Escalation Thresholds (Application Logic)

```
Pelanggaran ringan   → 5 poin    → teguran_lisan
Pelanggaran sedang   → 15 poin   → teguran_tertulis
Pelanggaran berat    → 30 poin   → panggilan_ortu
Akumulasi 50 poin    → skorsing_1_hari
Akumulasi 75 poin    → skorsing_3_hari
Akumulasi 100 poin   → skorsing_1_minggu
Akumulasi 150 poin   → dikeluarkan (perlu approval kepsek)
```

Threshold ini dikonfigurasi per sekolah, bukan hardcoded.

### Vernon Relationships

**student_disciplines:**

| Relasi | Tipe | Autoload | Alasan |
|---|---|---|---|
| `student` | belongs_to | **Ya** | Pemilik catatan |
| `academic_year` | belongs_to | **Ya** | Konteks tahun ajaran |
| `rule` | belongs_to | **Ya** | Nama dan kategori pelanggaran/penghargaan |

### _rels / _data Structure

```json
{
  "_rels": {
    "student_id": "018f...",
    "academic_year_id": "018f...",
    "rule_id": "018f..."
  },
  "_data": {
    "student": { "id": "018f...", "full_name": "Ahmad", "nis": "12345" },
    "academic_year": { "id": "018f...", "name": "2025/2026" },
    "rule": { "id": "018f...", "name": "Terlambat masuk kelas", "code": "V-001", "category": "ringan", "rule_type": "violation" }
  }
}
```

### API Endpoints

```
# Rules (Master)
GET    /api/v1/discipline-rules                     — List aturan tata tertib
POST   /api/v1/discipline-rules                     — Buat aturan
PUT    /api/v1/discipline-rules/{id}                — Update aturan

# Discipline Records
GET    /api/v1/students/{id}/disciplines             — Catatan disiplin siswa
GET    /api/v1/students/{id}/disciplines/summary     — Rekap poin per semester
POST   /api/v1/student-disciplines                   — Catat kejadian
PUT    /api/v1/student-disciplines/{id}             — Update catatan
POST   /api/v1/student-disciplines/{id}/approve      — Approve sanksi (kepsek)
POST   /api/v1/student-disciplines/{id}/notify-parent — Tandai ortu sudah diberitahu
```

## Consequences

### Positive

- **Dua arah**: Violation + merit memberikan gambaran lengkap perilaku siswa.
- **Eskalasi terstruktur**: Sanksi bertingkat sesuai kultur sekolah Indonesia.
- **Master rule**: Aturan bisa dikustomisasi per sekolah.
- **Audit trail**: Pelapor + approver + notifikasi ortu terdokumentasi.
- **Rapor integration**: Data disiplin bisa di-aggregate untuk catatan perilaku di rapor.

### Negative / Trade-offs

- **Threshold configurable**: Eskalasi threshold perlu disimpan di config terpisah — belum di-define di ADR ini.
- **Poin reset**: Belum ditentukan apakah poin reset per semester atau per tahun ajaran — perlu policy decision.
- **Surat peringatan**: Generate surat formal memerlukan template engine yang belum di-define.
- **Privacy concern**: Data disiplin sensitif — perlu access control yang lebih ketat (future).

## Alternatives Considered

### 1. Poin sebagai kolom di Student Table
- Ditolak: tidak ada riwayat per kejadian, tidak bisa audit trail.

### 2. Single table tanpa master rule
- Ditolak: tanpa master, setiap pencatatan harus input kategori + poin manual — inkonsisten.

### 3. Pisah violation dan merit ke 2 tabel berbeda
- Ditolak: struktur data sama — `record_type` field cukup untuk membedakan.

## Test Cases

### Unit Tests — Descriptor & Validation

| # | Test Case | Input | Expected |
|---|---|---|---|
| U01 | `RuleDescriptor.TableName()` | — | `"discipline_rules"` |
| U02 | `DisciplineDescriptor.TableName()` | — | `"student_disciplines"` |
| U03 | Validate rejects invalid `rule_type` | `"warning"` | Error: must be violation/merit |
| U04 | Validate rejects invalid `category` | `"minor"` | Error: invalid category |
| U05 | Validate rejects invalid `record_type` | `"warning"` | Error: must be violation/merit |
| U06 | Validate rejects invalid `action_taken` | `"detention"` | Error: invalid action |
| U07 | Validate rejects missing `incident_date` | null | Error: required |
| U08 | Validate accepts valid discipline record | All fields valid | No error |

### Integration Tests — Rules

| # | Test Case | Action | Expected |
|---|---|---|---|
| I01 | Create discipline rule | POST with code + name + type + category + points | 201 |
| I02 | Unique code per company | Create 2 rules same code | 409/422 |
| I03 | Rule type CHECK | INSERT with `rule_type = 'warning'` | DB error |
| I04 | Category CHECK | INSERT with `category = 'minor'` | DB error |

### Integration Tests — Discipline Records

| # | Test Case | Action | Expected |
|---|---|---|---|
| I05 | Record violation | POST with valid violation data | 201, record created with `_data` |
| I06 | Record merit | POST with `record_type = 'merit'` | 201, positive points |
| I07 | Get student discipline history | GET /students/{id}/disciplines | 200, ordered by date desc |
| I08 | Get discipline summary | GET /students/{id}/disciplines/summary | 200, total violation points + merit points |
| I09 | Action CHECK enforced | INSERT with `action_taken = 'detention'` | DB error |
| I10 | Semester CHECK enforced | INSERT with `semester = 'midterm'` | DB error |

### Integration Tests — Workflow

| # | Test Case | Action | Expected |
|---|---|---|---|
| I11 | Approve discipline action | POST /student-disciplines/{id}/approve | 200, `approved_by` + `approved_at` set |
| I12 | Mark parent notified | POST /student-disciplines/{id}/notify-parent | 200, `parent_notified = true`, `notified_at` set |
| I13 | Non-kepsek cannot approve severe action | Approve `dikeluarkan` by teacher | 403 |

### Integration Tests — SyncEngine

| # | Test Case | Action | Expected |
|---|---|---|---|
| I14 | StudentUpdated syncs to disciplines | Update student name | `_data.student.full_name` updated |
| I15 | RuleUpdated syncs to disciplines | Update rule name | `_data.rule.name` updated |

### Integration Tests — Tenant Isolation

| # | Test Case | Action | Expected |
|---|---|---|---|
| I16 | Cannot access other tenant's discipline records | GET with wrong tenant | 404 |

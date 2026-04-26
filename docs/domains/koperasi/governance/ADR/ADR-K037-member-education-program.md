# ADR-K037: Member Education Program (Pendidikan Anggota)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

UU Koperasi No. 25/1992 Pasal 7 menyebutkan bahwa koperasi bertugas menyelenggarakan pendidikan perkoperasian bagi anggota. Pendidikan keuangan dan perkoperasian penting karena:

- Meningkatkan literasi keuangan anggota
- Mengurangi NPL (nasabah paham kewajibannya)
- Meningkatkan partisipasi di RAT
- Memenuhi kewajiban regulasi (pelaporan ke Dinas Koperasi)
- Menumbuhkan budaya menabung di kalangan siswa

Saat ini tidak ada ADR yang mengatur program pendidikan anggota.

## Decision

### 1. Education Program Types

```
education_program_type:
├── COOP_BASIC           ← Pendidikan perkoperasian dasar (wajib untuk anggota baru)
├── FINANCIAL_LITERACY   ← Literasi keuangan (budgeting, menabung, utang)
├── PRODUCT_TRAINING     ← Cara menggunakan produk koperasi (tabungan, pinjaman)
├── ISLAMIC_FINANCE      ← Keuangan syariah (hanya BMT mode)
├── ENTREPRENEURSHIP     ← Kewirausahaan untuk anggota
├── DIGITAL_SKILLS       ← Cara menggunakan app/mobile
└── RAT_PREPARATION      ← Persiapan untuk RAT (hak, kewajiban, prosedur)
```

### 2. Data Model

```
education_course:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Course Info ──
├── course_code           VARCHAR
├── title                 VARCHAR
├── description           TEXT
├── program_type          ENUM (coop_basic, financial_literacy, product_training,
│                                islamic_finance, entrepreneurship,
│                                digital_skills, rat_preparation)
│
├── ── Content ──
├── content_type          ENUM (article, video, slide, interactive, offline_class, webinar)
├── content_url           VARCHAR (nullable)
├── content_doc_id        UUID (nullable, FK → document)
├── duration_minutes      INT
│
├── ── Prerequisites ──
├── prerequisite_ids      UUID[] (nullable, FK → education_course)
├── mandatory             BOOLEAN DEFAULT false
│   # mandatory = wajib diikuti (misal: COOP_BASIC untuk anggota baru)
│
├── ── Assessment ──
├── has_quiz              BOOLEAN DEFAULT false
├── passing_score         INT (nullable, 0-100)
├── certificate_template  VARCHAR (nullable)
│
├── ─── Targeting ──
├── target_audience       ENUM (all_members, new_members, staff, students, parents)
├── applicable_mode       ENUM (general_only, islamic_only, both)
│
├── ── Status ──
├── is_active             BOOLEAN DEFAULT true
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

```
education_enrollment:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── course_id             UUID (FK → education_course)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Progress ──
├── status                ENUM (enrolled, in_progress, completed, failed, expired)
├── progress_pct          INT DEFAULT 0    ← 0-100%
├── started_at            TIMESTAMPTZ (nullable)
├── completed_at          TIMESTAMPTZ (nullable)
│
├── ── Quiz Result ──
├── quiz_score            INT (nullable)
├── quiz_attempts         INT DEFAULT 0
├── passed                BOOLEAN (nullable)
│
├── ── Certificate ──
├── certificate_id        UUID (nullable, FK → document)
├── certificate_issued_at TIMESTAMPTZ (nullable)
│
├── ── Feedback ──
├── rating                INT (nullable, 1-5)
├── feedback              TEXT (nullable)
│
├── ── Audit ──
├── enrolled_at           TIMESTAMPTZ
├── deadline_at           TIMESTAMPTZ (nullable)
└── created_at            TIMESTAMPTZ
```

### 3. Mandatory Training Requirements

| Program | Mandatory For | When | Deadline |
|---|---|---|---|
| COOP_BASIC | Anggota baru | Saat approval keanggotaan | 30 hari setelah approval |
| PRODUCT_TRAINING | Anggota dengan pinjaman | Saat disbursement | 14 hari setelah disbursement |
| RAT_PREPARATION | Semua anggota | Sebelum RAT | 7 hari sebelum RAT |
| DIGITAL_SKILLS | Pengguna app | Saat aktivasi app | Optional |
| FINANCIAL_LITERACY | Siswa (e-wallet) | Saat pembukaan rekening | 14 hari |

**Penanganan jika tidak menyelesaikan mandatory training:**
- COOP_BASIC: Reminder otomatis, jika 30 hari belum selesai → flag di profil
- PRODUCT_TRAINING: Tidak blocking, tapi quiz harus lulus sebelum pinjaman berikutnya
- RAT_PREPARATION: Info only, tidak blocking

### 4. Education KPI Tracking

```
education_kpi:
├── Completion rate per course
│   ├── enrolled vs completed within deadline
│   └── Target: ≥ 80% untuk mandatory courses
│
├── Average quiz score
│   └── Target: ≥ 70
│
├── Member satisfaction (rating)
│   └── Target: ≥ 4.0 / 5.0
│
├── Training hours per member per year
│   └── Target: ≥ 4 hours
│
└── Impact metrics
    ├── NPL rate for trained vs untrained members
    ├── Savings rate for trained vs untrained
    └── RAT attendance for trained vs untrained
```

### 5. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| COOP_BASIC | Perkoperasian umum | + prinsip syariah koperasi |
| ISLAMIC_FINANCE | Tidak ada | Wajib untuk anggota baru BMT |
| Product training | Produk konvensional | Produk syariah |
| RAT preparation | Standard | + peran DPS |
| Certificate | Standard | + DPS signature (untuk islamic courses) |

### 6. Vernon _rels dan _data Structure

**Enrollment _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f...",
  "course_id":   "018f..."
}
```

**Enrollment _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "course": {
    "code":     "COOP-BASIC-01",
    "title":    "Pendidikan Perkoperasian Dasar",
    "type":     "coop_basic"
  }
}
```

### 7. Authorization — RBAC

| Permission | Admin | Manager | Nasabah |
|---|---|---|---|
| Create/edit course | v | v | - |
| View course catalog | v | v | v |
| Enroll in course | v | v | v |
| Track own progress | - | - | v |
| View all enrollments | v | v | - |
| Issue certificate | v | v | - |
| View education reports | v | v | - |
| Configure mandatory requirements | v | - | - |

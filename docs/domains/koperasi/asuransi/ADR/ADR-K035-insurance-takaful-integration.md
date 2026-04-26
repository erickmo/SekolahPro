# ADR-K035: Insurance / Takaful Integration

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Pinjaman koperasi memiliki risiko kredit — nasabah bisa saja meninggal, sakit parah, atau kehilangan pekerjaan sehingga tidak bisa membayar. Tanpa proteksi asuransi:

- Koperasi menanggung full loss jika nasabah meninggal dengan pinjaman outstanding
- Ahli waris mewarisi utang yang bisa menjadi beban berat
- Koperasi harus menyisihkan PPAP besar → mengurangi profitabilitas
- Tidak ada safety net untuk nasabah yang mengalami kecelakaan/cacat

Asuransi jiwa/kredit memberikan proteksi bagi kedua belah pihak. Untuk BMT mode, produk takaful (asuransi syariah) wajib digunakan sebagai pengganti asuransi konvensional.

## Decision

### 1. Insurance Products

```
insurance_product_type:
├── CREDIT_LIFE           ← Asuransi jiwa kredit (meninggal/cacat total)
├── CREDIT_DISABILITY     ← Asuransi kecacatan sementara
├── CREDIT_UNEMPLOYMENT   ← Asuransi kehilangan pekerjaan (untuk guru/staff)
├── COLLATERAL_PROTECTION ← Asuransi untuk jaminan (kebakaran, bencana)
└── PERSONAL_ACCIDENT     ← Kecelakaan diri (optional untuk nasabah)
```

### 2. Data Model

```
insurance_product:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Product Info ──
├── product_code          VARCHAR
├── product_name          VARCHAR
├── product_type          ENUM (credit_life, credit_disability,
│                                credit_unemployment, collateral_protection,
│                                personal_accident)
├── provider_id           UUID (nullable, FK → insurance_provider)
│
├── ── Coverage ──
├── coverage_type         ENUM (decreasing_term, level_term, full_outstanding)
│   # decreasing_term: coverage menurun sesuai outstanding
│   # level_term: coverage tetap selama tenor
│   # full_outstanding: coverage = sisa outstanding saat claim
│
├── coverage_percentage   DECIMAL(5,2)    ← % dari plafon yang dicover
├── max_coverage_amount   BIGINT (nullable)
│
├── ── Premium ──
├── premium_type          ENUM (single_upfront, monthly_deducted, annual)
├── premium_rate          DECIMAL(7,5)    ← Rate per mille (%)
├── premium_paid_by       ENUM (nasabah, koperasi, shared)
│
├── ── Eligibility ──
├── min_age               INT (nullable)
├── max_age               INT (nullable)
├── health_check_required BOOLEAN DEFAULT false
├── max_plafon            BIGINT (nullable)
│
├── ── Mode ──
├── applicable_mode       ENUM (general_only, islamic_only, both)
│
├── is_active             BOOLEAN DEFAULT true
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

```
insurance_policy:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── pinjaman_id           UUID (FK → pinjaman)
├── nasabah_id            UUID (FK → nasabah)
├── product_id            UUID (FK → insurance_product)
│
├── ── Policy Info ──
├── policy_number         VARCHAR
├── coverage_amount       BIGINT          ← Jumlah yang dicover
├── premium_amount        BIGINT          ← Total premium
├── premium_type          ENUM (single_upfront, monthly_deducted)
│
├── ── Term ──
├── effective_date        DATE
├── expiry_date           DATE            ← Biasanya = tenor pinjaman
│
├── ── Status ──
├── status                ENUM (active, claimed, expired, cancelled)
├── cancellation_reason   TEXT (nullable)
│
├── ── Audit ──
├── issued_at             TIMESTAMPTZ
├── issued_by             UUID (FK → user)
└── created_at            TIMESTAMPTZ
```

```
insurance_claim:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── policy_id             UUID (FK → insurance_policy)
│
├── ── Claim Info ──
├── claim_number          VARCHAR
├── claim_type            ENUM (death, total_disability, partial_disability,
│                                unemployment, collateral_damage, accident)
├── claim_date            DATE
├── incident_date         DATE
│
├── ── Evidence ──
├── document_ids          UUID[]          ← Surat kematian, surat dokter, dll
├── description           TEXT
│
├── ── Assessment ──
├── assessed_amount       BIGINT (nullable)
├── approved_amount       BIGINT (nullable)
├── assessor_id           UUID (nullable, FK → user)
├── assessment_notes      TEXT (nullable)
│
├── ── Settlement ──
├── status                ENUM (submitted, under_review, approved, rejected, paid)
├── settlement_date       DATE (nullable)
├── settlement_type       ENUM (pay_to_koperasi, pay_to_beneficiary, offset_loan)
│
├── ── Audit ──
├── submitted_at          TIMESTAMPTZ
├── resolved_at           TIMESTAMPTZ (nullable)
└── created_at            TIMESTAMPTZ
```

### 3. Premium Collection

```
Premium Collection Methods:
├── SINGLE_UPFRONT
│   ├── Dibayar sekali saat disbursement
│   ├── Ditambahkan ke plafon atau dipotong dari pencairan
│   └── Jurnal: Debit Kas Nasabah → Credit Premis Asuransi Payable
│
├── MONTHLY_DEDUCTED
│   ├── Dipotong dari angsuran bulanan
│   ├── Komponen angsuran: Pokok + Bunga/Margin + Premi
│   └── Otomatis oleh sistem bersamaan dengan angsuran
│
└── ANNUAL
    ├── Dibayar tahunan oleh koperasi (koperasi tanggung)
    └── Jurnal: Debit Biaya Asuransi → Credit Kas
```

### 4. Claim Workflow

```
Claim Process:
┌──────────────────────────────────────────────────┐
│ Step 1: Claim Submission (Hari ke-1)             │
│ ├── Ahli waris atau koperasi ajukan klaim        │
│ ├── Upload dokumen pendukung                     │
│ └── Status → submitted                           │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 2: Document Verification (H+1 s/d H+7)     │
│ ├── Staff verifikasi kelengkapan dokumen         │
│ ├── Request tambahan dokumen jika perlu          │
│ └── Status → under_review                        │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 3: Assessment (H+7 s/d H+14)               │
│ ├── Hitung jumlah klaim berdasarkan coverage     │
│ ├── Outstanding loan offset vs cash payment      │
│ └── Approval oleh Manager                        │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 4: Settlement (H+14 s/d H+30)              │
│ ├── Offset: klaim digunakan lunasi pinjaman      │
│ ├── Cash: klaim dibayar ke ahli waris            │
│ ├── Jurnal penyelesaian klaim                    │
│ └── Status → paid                                │
└──────────────────────────────────────────────────┘
```

### 5. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Product type | Asuransi konvensional | Takaful (asuransi syariah) |
| Premium | Premi (fixed) | Kontribusi tabarru' (donasi) |
| Claim payment | Dari perusahaan asuransi | Dari pool tabarru' + operator |
| Profit element | Boleh ada | Tidak boleh (tabarru') |
| DPS oversight | Tidak perlu | DPS harus approve provider |
| Provider | Asuransi umum | Perusahaan asuransi syariah |
| Surplus distribution | Tidak ada | Bisa dibagikan ke peserta |

### 6. Vernon _rels dan _data Structure

**Policy _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f...",
  "pinjaman_id": "018f..."
}
```

**Policy _data:**
```json
{
  "nasabah": {
    "id":        "018f...",
    "full_name": "Ahmad Fauzi"
  },
  "pinjaman": {
    "loan_number":  "PJ-2026-JKT-00000001",
    "outstanding":  35000000
  },
  "product": {
    "name":           "Credit Life Insurance",
    "coverage_type":  "decreasing_term",
    "coverage_amount": 35000000
  }
}
```

### 7. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|---|---|---|---|---|
| View policy details | v | v | v | v |
| Issue policy | - | v | v | v |
| Submit claim | v | v | v | v |
| Process claim | - | v | v | v |
| Approve claim | - | - | v | v |
| Configure products | - | - | - | v |
| Manage providers | - | - | - | v |
| View claim reports | - | v | v | v |

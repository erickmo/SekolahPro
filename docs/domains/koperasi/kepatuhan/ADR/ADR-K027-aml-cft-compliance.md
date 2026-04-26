# ADR-K027: AML/CFT Compliance (Anti-Money Laundering)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi yang beroperasi sebagai LKM (Lembaga Keuangan Mikro) dan BMT berada di bawah pengawasan OJK dan **WAJIB** mematuhi regulasi Anti-Money Laundering (AML) dan Counter Financing of Terrorism (CFT):

- **UU No. 8/2010** tentang Pencegahan dan Pemberantasan Tindak Pidana Pencucian Uang (TPPU)
- **POJK 12/POJK.05/2016** tentang Penerapan Anti Pencucian Uang
- **POJK 16/POJK.05/2016** tentang Penerapan Prinsip Mengenal Nasabah
- **PPATK** (Pusat Pelaporan dan Analisis Transaksi Keuangan) sebagai otoritas penerima LTKM/TKM

Saat ini K001 (Nasabah) memiliki KYC dasar (identitas, dokumen), tetapi belum mencakup:
- Customer Due Diligence (CDD) dan Enhanced Due Diligence (EDD)
- Politically Exposed Person (PEP) screening
- Transaction monitoring dan suspicious pattern detection
- Laporan Transaksi Keuangan Mencurigakan (LTKM) ke PPATK
- Laporan Transaksi dalam Nilai Tertentu (TKM)
- Record retention 5 tahun pasca penutupan rekening

## Decision

### 1. Risk-Based Approach Framework

```
aml_risk_framework:
├── CUSTOMER RISK
│   ├── Individual nasabah risk scoring
│   ├── Business entity risk scoring
│   ├── PEP (Politically Exposed Person) classification
│   └── High-risk jurisdiction nexus
│
├── PRODUCT RISK
│   ├── Anonymous product risk (jika ada)
│   ├── Cross-border product risk
│   ├── High-value transaction product risk
│   └── Complex product risk
│
├── GEOGRAPHIC RISK
│   ├── High-risk countries (FATF black/grey list)
│   ├── Border regions
│   └── Conflict zones
│
└── DELIVERY CHANNEL RISK
    ├── Face-to-face (lowest risk)
    ├── Non-face-to-face (digital) — higher risk
    ├── Third-party introduced — highest risk
    └── Correspondent relationship
```

### 2. Customer Due Diligence (CDD) Levels

```
cdd_level:
├── SIMPLIFIED_DUE_DILIGENCE (SDD)
│   ├── Untuk nasabah risiko rendah
│   ├── Verifikasi identitas dasar
│   ├── Pemantauan transaksi standar
│   └── Review period: 3 tahun
│
├── CUSTOMER_DUE_DILIGENCE (CDD)
│   ├── Untuk nasabah risiko menengah
│   ├── Verifikasi identitas + tujuan membuka rekening
│   ├── Pemantauan transaksi reguler
│   └── Review period: 2 tahun
│
└── ENHANCED_DUE_DILIGENCE (EDD)
    ├── Untuk nasabah risiko tinggi (WAJIB)
    ├── Verifikasi mendalam + sumber dana + sumber kekayaan
    ├── Pemantauan transaksi ketat
    ├── Approval senior management untuk relasi bisnis
    └── Review period: 1 tahun
```

**Kapan EDD wajib diterapkan:**
- PEP (Politically Exposed Person) atau kerabat/rekan dekat PEP
- Nasabah dari negara berisiko tinggi (FATF list)
- Transaksi unusual/suspicious pattern terdeteksi
- Correspondent banking relationship
- Nasabah yang menolak menyediakan informasi CDD
- Entity dengan struktur kepemilikan kompleks

### 3. CDD Data Model

```
aml_customer_due_diligence:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Risk Assessment ──
├── risk_level            ENUM (low, medium, high, prohibited)
├── risk_score            INT             ← 1-100 scoring
├── risk_category         JSONB           ← Detail per category
│   ├── customer_risk     ENUM (low, medium, high)
│   ├── product_risk      ENUM (low, medium, high)
│   ├── geographic_risk   ENUM (low, medium, high)
│   └── channel_risk      ENUM (low, medium, high)
│
├── ── CDD Level ──
├── cdd_level             ENUM (sdd, cdd, edd)
├── cdd_purpose           TEXT            ← Tujuan membuka rekening
├── source_of_funds       TEXT            ← Sumber dana (EDD)
├── source_of_wealth      TEXT            ← Sumber kekayaan (EDD)
│
├── ── PEP Screening ──
├── is_pep                BOOLEAN DEFAULT false
├── pep_type              ENUM (domestic, foreign, international_organization, none)
├── pep_position          VARCHAR (nullable)
├── pep_country           VARCHAR (nullable)
├── pep_relationship      ENUM (self, family, close_associate, none)
├── pep_screening_date    DATE
├── pep_screening_source  VARCHAR         ← Nama database/provider
│
├── ── Beneficial Owner ──
├── beneficial_owner_name     VARCHAR (nullable)
├── beneficial_owner_id_no    VARCHAR (nullable)
├── beneficial_ownership_pct  DECIMAL(5,2) (nullable)
├── bo_verified           BOOLEAN DEFAULT false
│
├── ── Ongoing Monitoring ──
├── next_review_date      DATE
├── last_reviewed_at      TIMESTAMPTZ
├── last_reviewed_by      UUID (FK → user)
├── review_count          INT DEFAULT 0
│
├── ── Status ──
├── status                ENUM (pending, active, escalated, restricted, exited)
├── restriction_reason    TEXT (nullable)
├── exit_reason           TEXT (nullable)
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
├── updated_at            TIMESTAMPTZ
├── created_by            UUID (FK → user)
└── updated_by            UUID (FK → user)
```

### 4. Transaction Monitoring

```
aml_transaction_monitoring:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
├── transaksi_id          UUID (FK → transaksi, nullable)
│
├── ── Monitoring Rule ──
├── rule_id               UUID (FK → aml_monitoring_rule)
├── rule_type             ENUM (threshold, pattern, velocity, anomaly, manual)
│
├── ── Alert Details ──
├── alert_type            ENUM (suspicious, threshold_exceeded, unusual_pattern,
│                                structuring, rapid_movement, mismatch)
├── severity              ENUM (low, medium, high, critical)
├── description           TEXT
├── transaction_details   JSONB           ← Snapshot transaksi yang trigger
│
├── ── Investigation ──
├── investigated_by       UUID (FK → user, nullable)
├── investigation_notes   TEXT (nullable)
├── investigation_outcome ENUM (false_positive, suspicious, confirmed, escalated)
├── ltkm_filed            BOOLEAN DEFAULT false   ← Laporan ke PPATK
├── ltkm_reference        VARCHAR (nullable)      ← Nomor referensi PPATK
│
├── ── Status ──
├── status                ENUM (new, investigating, resolved, escalated, reported)
├── resolved_at           TIMESTAMPTZ (nullable)
│
├── ── Audit ──
├── detected_at           TIMESTAMPTZ
└── created_at            TIMESTAMPTZ
```

### 5. Monitoring Rules

```
aml_monitoring_rule:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Rule Definition ──
├── rule_code             VARCHAR         ← "THRESHOLD_001", "VELOCITY_001"
├── rule_name             VARCHAR
├── rule_type             ENUM (threshold, pattern, velocity, anomaly)
├── description           TEXT
│
├── ── Parameters ──
├── parameters            JSONB
│   # THRESHOLD: { "amount": 500000000, "period_days": 1 }
│   # VELOCITY: { "max_count": 5, "period_hours": 24, "amount_min": 10000000 }
│   # PATTERN: { "pattern_type": "structuring", "variance_pct": 10 }
│   # ANOMALY: { "deviation_std": 3 }
│
├── ── Configuration ──
├── is_active             BOOLEAN DEFAULT true
├── applies_to            ENUM (all, high_risk_only, pep_only)
├── auto_alert            BOOLEAN DEFAULT true
├── auto_block            BOOLEAN DEFAULT false   ← Hati-hati: bisa block transaksi legit
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

**Default monitoring rules:**

| Code | Type | Description | Threshold |
|---|---|---|---|
| THRESHOLD_001 | threshold | Single cash transaction ≥ 500 juta | 500,000,000 |
| THRESHOLD_002 | threshold | Single transfer ≥ 1 miliar | 1,000,000,000 |
| VELOCITY_001 | velocity | ≥ 3 transactions > 50 juta dalam 24 jam | 3 tx / 24h |
| VELOCITY_002 | velocity | ≥ 5 deposits in 1 hari by same person | 5 tx / day |
| STRUCTURE_001 | pattern | Multiple deposits just below threshold (structuring) | < 10% variance |
| RAPID_001 | pattern | Fund in → fund out within 24 jam | same day |
| DORMANT_001 | anomaly | Dormant account suddenly active with large amount | > 50 juta |
| MISMATCH_001 | anomaly | Transaction not matching customer profile | per profile |

### 6. Reporting to PPATK

**LTKM (Laporan Transaksi Keuangan Mencurigakan):**
- Wajib dilaporkan ≤ 3 hari kerja setelah indikasi suspicious terdeteksi
- Format sesuai template PPATK
- Identitas pelapor dilindungi (whistleblower protection)

**TKM (Transaksi dalam Nilai Tertentu):**
- Cash transaction ≥ Rp 500.000.000 — wajib dilaporkan bulanan
- Transfer ≥ Rp 1.000.000.000 — wajib dilaporkan bulanan
- Laporan dikirim ke PPATK paling lambat tanggal 10 bulan berikutnya

```
aml_report:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Report Info ──
├── report_type           ENUM (ltkm, tkm_monthly, tkm_annual, internal_sar)
├── reporting_period      VARCHAR         ← "2026-01", "2026"
├── report_status         ENUM (draft, submitted, acknowledged, rejected)
│
├── ── PPATK Reference ──
├── ppatk_reference       VARCHAR (nullable)
├── submitted_at          TIMESTAMPTZ (nullable)
├── acknowledged_at       TIMESTAMPTZ (nullable)
│
├── ── Content ──
├── report_content        JSONB           ← Structured report data
├── related_alert_ids     UUID[]          ← FK → aml_transaction_monitoring
│
├── ── Audit ──
├── prepared_by           UUID (FK → user)
├── approved_by           UUID (FK → user)
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

### 7. Record Retention

| Document Type | Retention Period | Start Date |
|---|---|---|
| CDD records (active nasabah) | Selama relasi bisnis aktif + 5 tahun | Account closure date |
| Transaction records | 5 tahun | Transaction date |
| AML alerts & investigations | 5 tahun | Resolution date |
| LTKM/TKM reports | 5 tahun | Submission date |
| PEP screening results | 5 tahun | Screening date |
| Training records (AML staff) | 5 tahun | Training date |

### 8. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Monitoring scope | Standard financial flows | + zakat/infaq/sadaqah flows |
| Suspicious pattern | Standard indicators | + unusual zakat/sadaqah patterns |
| PEP screening | Standard | Standard + ulama/influencer PEP |
| High-risk products | Cash-intensive products | + hawala-like informal transfer |
| Reporting | LTKM/TKM standard | Same + Dana Sosial reporting |

### 9. Vernon _rels dan _data Structure

**CDD _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f..."
}
```

**CDD _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "risk": {
    "level":     "medium",
    "score":     45,
    "cdd_level": "cdd"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → re-evaluate risk score jika data berubah
- `TransaksiCreatedEvent` → run monitoring rules, generate alerts if triggered
- `AlertEscalatedEvent` → notify compliance officer for investigation

### 10. Authorization — RBAC

| Permission | Teller | Compliance Officer | Manager | Admin |
|---|---|---|---|---|
| View own CDD info | - | - | - | - |
| Perform CDD (basic) | v | v | v | v |
| Perform EDD | - | v | v | v |
| PEP screening | - | v | v | v |
| View monitoring alerts | - | v | v | v |
| Investigate alerts | - | v | v | v |
| Resolve alerts (false positive) | - | v | v | v |
| File LTKM to PPATK | - | - | v | v |
| Generate TKM report | - | v | v | v |
| Configure monitoring rules | - | - | - | v |
| View AML reports | - | v | v | v |
| Risk assessment override | - | - | - | v |
| Manage record retention | - | - | - | v |

### 11. Consequences

**Keuntungan:**
- Compliance terhadap UU TPPU dan POJK AML
- Terhindar dari sanksi administratif dan pidana
- Reputasi koperasi terjaga sebagai lembaga keuangan yang trusted
- Early detection terhadap penyalahgunaan sistem

**Risiko:**
- False positive alerts bisa mengganggu operasional
- Cost implementasi monitoring system
- Training berkelanjutan untuk compliance staff
- Privacy concern dari nasabah (perlu komunikasi yang baik)

**Mitigasi:**
- Tuning monitoring rules secara berkala untuk reduce false positives
- Start dengan rule minimal, expand bertahap
- E-learning module untuk AML training
- Privacy notice di onboarding nasabah

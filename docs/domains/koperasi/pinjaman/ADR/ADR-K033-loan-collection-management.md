# ADR-K033: Loan Collection Management

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

K008 (Angsuran & Jadwal) mengatur pembayaran angsuran reguler dan K009 (Denda & Penalti) mengatur penalti keterlambatan. Tetapi tidak ada ADR yang mengatur **proses penagihan** ketika nasabah benar-benar tidak membayar:

- Tidak ada workflow penagihan bertingkat (reminder → call → visit → legal)
- Tidak ada assignment collector/penagih
- Tidak ada tracking performa collection
- Tidak ada mekanisme negosiasi/restructuring
- Tidak ada escalation ke proses hukum

Tanpa collection management, NPL akan terus meningkat tanpa ada upaya penyelesaian yang terstruktur.

## Decision

### 1. Collection Aging Bucket

```
aging_bucket:
├── CURRENT         ← 0 hari telat — lancar
├── DPW_1_7         ← 1-7 hari telat (DPD 1-7)
├── DPW_8_30        ← 8-30 hari telat (DPD 8-30)
├── DPW_31_60       ← 31-60 hari telat (DPD 31-60)
├── DPW_61_90       ← 61-90 hari telat (DPD 61-90)
├── DPW_91_180      ← 91-180 hari telat (DPD 91-180) — Substandard
├── DPW_181_270     ← 181-270 hari telat — Doubtful
├── DPW_271_360     ← 271-360 hari telat — Loss (provision 100%)
└── DPW_360_PLUS    ← > 360 hari telat — Write-off candidate
```

**NPL Classification (sesuai OJK):**

| DPD Range | Kualitas | PPAP | Kategori |
|---|---|---|---|
| 0 | Lancar | 1% | Performing |
| 1-90 | Dalam Perhatian Khusus (DPK) | 5% | Performing |
| 91-180 | Kurang Lancar (Substandard) | 15% | NPL |
| 181-270 | Diragukan (Doubtful) | 50% | NPL |
| 271-360 | Macet (Loss) | 100% | NPL |

### 2. Collection Actions per Bucket

```
collection_strategy_per_bucket:
├── CURRENT (0 hari)
│   ├── No action needed
│   └── Auto reminder H-3 sebelum jatuh tempo (via K022)
│
├── DPW_1_7 (1-7 hari)
│   ├── Auto SMS/WhatsApp reminder
│   ├── Notifikasi ke nasabah
│   └── Auto-flag di dashboard teller
│
├── DPW_8_30 (8-30 hari)
│   ├── Telepon oleh collector (Level 1)
│   ├── Jadwal kunjungan jika tidak bisa dihubungi
│   ├── Notifikasi ke penjamin (jika ada)
│   └── Penalty auto-calculation (ref K009)
│
├── DPW_31_60 (31-60 hari)
│   ├── Surat Peringatan 1 (SP1)
│   ├── Home visit oleh collector
│   ├── Negosiasi restructuring ditawarkan
│   └── Escalation ke supervisor collection
│
├── DPW_61_90 (61-90 hari)
│   ├── Surat Peringatan 2 (SP2)
│   ├── Multiple home visits
│   ├── Meeting dengan nasabah + penjamin
│   ├── Mandatory restructuring offer
│   └── Escalation ke manager
│
├── DPW_91_180 (91-180 hari) — Substandard
│   ├── Surat Peringatan 3 (SP3) — final
│   ├── Pengenaan denda maksimal
│   ├── Penjamin diminta bertanggung jawab
│   ├── Jaminan/agunan mulai diproses (ref K010)
│   └── Persiapan legal action
│
├── DPW_181_270 (181-270 hari) — Doubtful
│   ├── Legal notice dari notaris
│   ├── Eksekusi jaminan (jika ada)
│   ├── Mediasi melalui pihak ketiga
│   └── Proposal write-off jika tidak memungkinkan
│
├── DPW_271_360 (271-360 hari) — Loss
│   ├── Write-off proposal ke management
│   ├── Laporan ke OJK (NPL reporting)
│   └── Penjualan NPL ke collector pihak ketiga (opsional)
│
└── DPW_360_PLUS (> 360 hari)
    ├── Write-off execution
    ├── Blacklist nasabah
    └── Keep in record for 5 years (regulatory)
```

### 3. Data Model

```
collection_case:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── pinjaman_id           UUID (FK → pinjaman)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Case Info ──
├── case_number           VARCHAR         ← "COL-2026-00001"
├── current_dpd           INT             ← Days Past Due saat ini
├── aging_bucket          ENUM (current, dpw_1_7, dpw_8_30, dpw_31_60,
│                                dpw_61_90, dpw_91_180, dpw_181_270,
│                                dpw_271_360, dpw_360_plus)
├── total_overdue_amount  BIGINT          ← Total tertunggak (pokok + bunga/margin + denda)
├── total_overdue_installments INT        ← Jumlah angsuran tertunggak
│
├── ── Assignment ──
├── assigned_collector_id UUID (nullable, FK → user)
├── assigned_at           TIMESTAMPTZ (nullable)
├── escalation_level      INT DEFAULT 0   ← 0=L1, 1=L2, 2=L3, 3=legal
│
├── ── Status ──
├── status                ENUM (active, restructured, legal, written_off,
│                                settled, closed)
├── resolution_type       ENUM (full_payment, restructuring, collateral_liquidation,
│                                guarantee_claim, write_off, partial_settlement)
│
├── ── Audit ──
├── opened_at             TIMESTAMPTZ
├── closed_at             TIMESTAMPTZ (nullable)
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

```
collection_activity:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── case_id               UUID (FK → collection_case)
│
├── ── Activity Info ──
├── activity_type         ENUM (auto_reminder, phone_call, home_visit,
│                                sp_letter, negotiation, restructuring_offer,
│                                legal_notice, collateral_processing,
│                                guarantee_contact, payment_received)
├── activity_date         DATE
├── performed_by          UUID (FK → user)
│
├── ── Details ──
├── contact_result        ENUM (contacted, no_answer, busy, wrong_number,
│                                message_left, visited_met, visited_not_home,
│                                letter_sent, letter_delivered, letter_returned)
├── notes                 TEXT
├── nasabah_response      ENUM (willing_to_pay, needs_restructuring,
│                                refuses_to_pay, unable_to_pay, dispute,
│                                promise_to_pay, no_response)
├── promise_amount        BIGINT (nullable)
├── promise_date          DATE (nullable)
│
├── ── Documents ──
├── document_ids          UUID[] (nullable)   ← Foto kunjungan, surat, dll
│
├── ── Follow-up ──
├── followup_required     BOOLEAN DEFAULT false
├── followup_date         DATE (nullable)
├── followup_type         VARCHAR (nullable)
│
├── ── Audit ──
└── created_at            TIMESTAMPTZ
```

### 4. Collector Performance Tracking

```
collector_performance:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── collector_id          UUID (FK → user)
│
├── ── Period ──
├── period_month          VARCHAR         ← "2026-01"
│
├── ── Activity Stats ──
├── total_calls           INT DEFAULT 0
├── successful_contacts   INT DEFAULT 0
├── total_visits          INT DEFAULT 0
├── successful_visits     INT DEFAULT 0
│
├── ── Collection Stats ──
├── cases_handled         INT DEFAULT 0
├── cases_resolved        INT DEFAULT 0
├── total_amount_collected BIGINT DEFAULT 0
├── promise_to_pay_count  INT DEFAULT 0
├── promise_kept_count    INT DEFAULT 0
│
├── ── Efficiency ──
├── contact_rate_pct      DECIMAL(5,2)    ← successful_contacts / total_calls * 100
├── resolution_rate_pct   DECIMAL(5,2)    ← cases_resolved / cases_handled * 100
├── collection_rate_pct   DECIMAL(5,2)    ← amount_collected / total_assigned * 100
├── promise_kept_rate_pct DECIMAL(5,2)    ← promises_kept / promises_made * 100
│
├── ── Audit ──
├── calculated_at         TIMESTAMPTZ
└── created_at            TIMESTAMPTZ

UNIQUE(tenant_id, collector_id, period_month)
```

### 5. Restructuring Workflow

```
Restructuring Proposal:
┌──────────────────────────────────────────────────┐
│ Step 1: Eligibility Check                        │
│ ├── DPD ≥ 30 hari                               │
│ ├── Nasabah menunjukkan willingness to pay       │
│ ├── Penyetujuan dari supervisor collection       │
│ └── Tidak dalam proses hukum                     │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 2: Proposal                                 │
│ ├── Opsi restructuring:                          │
│ │   ├── Perpanjangan tenor (max 2x tenor awal)   │
│ │   ├── Pengurangan angsuran (lower amount)      │
│ │   ├── Grace period (3-6 bulan pokok saja)      │
│ │   └── Kombinasi                                │
│ ├── Simulasi ulang jadwal angsuran               │
│ └── Approval: Manager + Bendahara                │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 3: Execution                                │
│ ├── Generate jadwal angsuran baru                │
│ ├── Jurnal adjustment (jika perlu)               │
│ ├── Update pinjaman status → RESTRUCTURED        │
│ └── Monitor ketat selama 3 bulan pertama         │
└──────────────────────────────────────────────────┘
```

### 6. Write-Off Process

```
Write-off Criteria:
├── DPD > 360 hari
├── Semua collection actions sudah dilakukan (documented)
├── Collateral sudah dieksekusi (jika ada)
├── Penjamin sudah diklaim (jika ada)
├── Legal action tidak feasible atau sudah gagal
├── Approval: Manager → Bendahara → Ketua
└── Reporting ke OJK dalam laporan NPL
```

### 7. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Collection approach | Standard | + pendekatan kekeluargaan/nasihat |
| Denda selama collection | Bunga terus berjalan | Ta'zir tetap (nominal tetap) |
| Restructuring | Bisa semua jenis | Harus sesuai prinsip syariah, DPS approve |
| Write-off | Standard | + izin DPS |
| Collateral execution | Standard | Rahn sesuai akad |
| NPL terminology | NPL | NPF (Non-Performing Financing) |
| Collection fee | Dibebankan ke nasabah | Ta'widh (harus ada bukti biaya nyata) |

### 8. Vernon _rels dan _data Structure

**Case _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f...",
  "pinjaman_id": "018f..."
}
```

**Case _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001",
    "phone":         "0812XXXXXXX"
  },
  "pinjaman": {
    "id":            "018f...",
    "loan_number":   "PJ-2026-JKT-00000001",
    "plafon":        50000000,
    "outstanding":   35000000
  },
  "collection": {
    "dpd":                  45,
    "aging_bucket":         "dpw_31_60",
    "total_overdue":        11200000,
    "overdue_installments": 2
  }
}
```

### 9. Authorization — RBAC

| Permission | Collector L1 | Collector L2 | Supervisor | Manager | Admin |
|---|---|---|---|---|---|
| View own cases | v | v | v | v | v |
| View all cases | - | - | v | v | v |
| Log collection activity | v | v | v | v | v |
| Assign cases | - | - | v | v | v |
| Issue SP letter | - | v | v | v | v |
| Offer restructuring | - | - | v | v | v |
| Approve restructuring | - | - | - | v | v |
| Initiate legal action | - | - | - | v | v |
| Propose write-off | - | - | - | v | v |
| Approve write-off | - | - | - | - | v |
| View collector performance | - | - | v | v | v |

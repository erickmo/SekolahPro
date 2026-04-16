# ADR-K039: Multi-Branch Consolidation Reporting

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi besar dengan beberapa cabang memerlukan laporan konsolidasian yang menggabungkan data dari seluruh cabang. Saat ini K014 (Kas & Cash Flow) mengatur inter-branch transfer dasar dan K017 (Laporan Regulasi) mengatur pelaporan, tetapi tidak ada ADR yang mengatur:

- Consolidated financial statements dari multi-branch
- Inter-branch transaction elimination
- Branch performance comparison
- Fund transfer pricing antar cabang
- Branch-level P&L

## Decision

### 1. Branch Financial Statements

```
branch_reporting_levels:
├── LEVEL 1 — Branch Standalone
│   ├── P&L branch (pendapatan - biaya operasional)
│   ├── Balance sheet branch (aset & kewajiban branch)
│   ├── NPL per branch
│   └── Volume transaksi per branch
│
├── LEVEL 2 — Inter-Branch Elimination
│   ├── Eliminasi transaksi antar cabang
│   ├── Eliminasi saldo piutang/hutang antar cabang
│   └── Unrealized profit elimination
│
├── LEVEL 3 — Consolidated
│   ├── Consolidated P&L
│   ├── Consolidated balance sheet
│   ├── Consolidated cash flow
│   └── Consolidated SHU calculation
│
└── LEVEL 4 — Regulatory Reporting
    ├── Single report ke OJK/Dinas untuk seluruh entity
    └── Branch-level data sebagai lampiran
```

### 2. Data Model

```
branch_financial_summary:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (nullable, FK → branch, null = consolidated)
│
├── ── Period ──
├── period_type           ENUM (daily, weekly, monthly, quarterly, yearly)
├── period_start          DATE
├── period_end            DATE
│
├── ── P&L ──
├── pendapatan_operasional    BIGINT DEFAULT 0
├── pendapatan_bunga_margin   BIGINT DEFAULT 0
├── pendapatan_lain           BIGINT DEFAULT 0
├── total_pendapatan          BIGINT DEFAULT 0
├── biaya_operasional         BIGINT DEFAULT 0
├── biaya_personel            BIGINT DEFAULT 0
├── biaya_administrasi        BIGINT DEFAULT 0
├── beban_ppap                BIGINT DEFAULT 0
├── total_biaya               BIGINT DEFAULT 0
├── laba_rugi_bersih          BIGINT DEFAULT 0
│
├── ── Balance Sheet ──
├── total_aset                BIGINT DEFAULT 0
├── kas_dan_bank              BIGINT DEFAULT 0
├── pinjaman_diberikan        BIGINT DEFAULT 0
├── simpanan_diterima         BIGINT DEFAULT 0
├── total_kewajiban           BIGINT DEFAULT 0
├── modal_sendiri             BIGINT DEFAULT 0
│
├── ── KPI ──
├── npl_ratio             DECIMAL(10,4) (nullable)
├── bopo_ratio            DECIMAL(10,4) (nullable)
├── roa                    DECIMAL(10,4) (nullable)
├── car                    DECIMAL(10,4) (nullable)
│
├── ── Audit ──
├── calculated_at         TIMESTAMPTZ
├── approved_by           UUID (nullable, FK → user)
└── created_at            TIMESTAMPTZ

UNIQUE(tenant_id, branch_id, period_type, period_start)
```

### 3. Inter-Branch Elimination Rules

```
elimination_rules:
├── INTER_BRANCH_TRANSFERS
│   ├── Branch A → Branch B transfer: eliminasi (net = 0)
│   ├── Cash transfer: eliminate receivable/payable
│   └── Method: branch_from debit, branch_to credit → net 0
│
├── INTER_BRANCH_LOAN_PARTICIPATION
│   ├── Jika cabang A funding pinjaman cabang B
│   ├── Eliminasi: funding aset di A, funding liability di B
│   └── Net = 0 di konsolidasi
│
└── PROFIT_ELIMINATION
    ├── Jika ada markup antar cabang (transfer pricing)
    ├── Unrealized profit eliminasi
    └── Only realized profit consolidated
```

### 4. Branch Performance Comparison

```
branch_comparison_metrics:
├── Volume
│   ├── Total simpanan per branch
│   ├── Total pinjaman per branch
│   └── Jumlah nasabah per branch
│
├── Quality
│   ├── NPL per branch
│   ├── PPAP adequacy per branch
│   └── Collection efficiency per branch
│
├── Profitability
│   ├── Net income per branch
│   ├── ROA per branch
│   ├── Cost-to-income ratio per branch
│   └── Contribution to total SHU
│
└── Growth
    ├── YoY simpanan growth per branch
    ├── YoY nasabah growth per branch
    └── New accounts vs target
```

### 5. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| P&L line items | Bunga pendapatan/biaya | Bagi hasil pendapatan/biaya |
| Inter-branch pricing | Transfer pricing bunga | Transfer pricing nisbah |
| DPS role | Tidak ada | DPS review consolidated syariah compliance |
| Dana sosial konsolidasi | Tidak ada | Consolidated dana sosial (baitul maal) |

### 6. Vernon _rels dan _data Structure

```json
{
  "tenant_id": "018f...",
  "branch_id": "018f..." // null for consolidated
}
```

### 7. Authorization — RBAC

| Permission | Branch Manager | Bendahara | Ketua | Admin |
|---|---|---|---|---|
| View own branch report | v | v | v | v |
| View all branch reports | - | v | v | v |
| View consolidated report | - | v | v | v |
| Generate consolidated | - | v | v | v |
| Approve consolidated | - | - | v | v |
| Configure elimination rules | - | - | - | v |

# ADR-K031: Cooperative Health Indicators & Risk Management

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Koperasi yang sehat harus dipantau melalui indikator keuangan dan operasional tertentu. OJK dan Dinas Koperasi menggunakan indikator ini untuk menilai kesehatan koperasi:

- **KUK (Keuangan Koperasi) scoring** oleh Dinas Koperasi
- **LKM health assessment** oleh OJK
- **Self-assessment** oleh manajemen untuk early warning

Saat ini K017 (Laporan Regulasi) menyebutkan beberapa rasio keuangan, tetapi tidak ada:
- Dashboard monitoring real-time untuk health indicators
- Early warning system dengan threshold dan alerting
- Trend analysis dan forecasting
- Stress testing scenarios
- Corrective action tracking

## Decision

### 1. Health Indicator Categories

```
health_indicator_categories:
├── MODAL (Capital)
│   ├── Capital Adequacy Ratio (CAR)
│   ├── Modal Sendiri / Total Aset
│   └── Pertumbuhan modal sendiri
│
├── ASSET QUALITY (Kualitas Aset)
│   ├── NPL (Non-Performing Loan) Ratio
│   ├── NPL Net vs Gross
│   ├── PPP (Penyisihan Penghapusan Aset)
│   └── Konsentrasi pembiayaan per nasabah/sector
│
├── MANAJEMEN (Management)
│   ├── Rasio biaya operasional / pendapatan operasional (BOPO)
│   ├── Staff productivity (pinjaman per staff)
│   ├── Training hours per staff per year
│   └── IT investment ratio
│
├── EARNINGS (Rentabilitas)
│   ├── ROA (Return on Assets)
│   ├── ROE (Return on Equity)
│   ├── NOM (Net Operating Margin)
│   └── SHU / Modal Sendiri
│
├── LIKUIDITAS (Liquidity)
│   ├── Cash Ratio
│   ├── Quick Ratio
│   ├── FDR (Financing to Deposit Ratio)
│   └── Loan to Deposit Ratio
│
└── SENSITIVITAS (Sensitivity)
    ├── Interest rate risk
    ├── Concentration risk
    └── Market risk
```

### 2. Key Performance Indicators

```
kpi_definition:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── KPI Info ──
├── kpi_code              VARCHAR         ← "CAR", "NPL_GROSS", "BOPO", "ROA"
├── kpi_name              VARCHAR
├── category              ENUM (capital, asset_quality, management, earnings, liquidity, sensitivity)
├── description           TEXT
│
├── ── Formula ──
├── formula               TEXT            ← SQL-like formula
├── unit                  ENUM (percentage, ratio, amount, count)
├── direction             ENUM (higher_is_better, lower_is_better, target_range)
│
├── ── Thresholds ──
├── healthy_min           DECIMAL(10,4) (nullable)
├── healthy_max           DECIMAL(10,4) (nullable)
├── warning_min           DECIMAL(10,4) (nullable)
├── warning_max           DECIMAL(10,4) (nullable)
├── critical_min          DECIMAL(10,4) (nullable)
├── critical_max          DECIMAL(10,4) (nullable)
│
│   Example NPL:
│   healthy: 0-5%, warning: 5-10%, critical: >10%
│
├── ── Calculation ──
├── calculation_frequency ENUM (daily, weekly, monthly, quarterly, yearly)
├── data_sources          VARCHAR[]       ← Tables/queries yang dibutuhkan
│
├── ── Audit ──
├── is_active             BOOLEAN DEFAULT true
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

### 3. Default KPI Thresholds

| KPI | Formula | Healthy | Warning | Critical | Direction |
|---|---|---|---|---|---|
| CAR | Modal Sendiri / Aset Tertimbang Menurut Risiko | ≥ 15% | 8-15% | < 8% | higher |
| NPL Gross | Pinjaman NPL / Total Pinjaman | 0-5% | 5-10% | > 10% | lower |
| NPL Net | (NPL - PPAP) / Total Pinjaman | 0-3% | 3-6% | > 6% | lower |
| BOPO | Biaya Operasional / Pendapatan Operasional | < 70% | 70-85% | > 85% | lower |
| ROA | Laba Bersih / Total Aset | ≥ 2% | 1-2% | < 1% | higher |
| ROE | Laba Bersih / Modal Sendiri | ≥ 15% | 8-15% | < 8% | higher |
| FDR | Total Pembiayaan / Total Simpanan | 78-92% | 70-78% atau 92-100% | < 70% atau > 100% | range |
| Cash Ratio | Kas + Bank / Simpanan Jangka Pendek | ≥ 10% | 5-10% | < 5% | higher |
| PPAP Adequacy | PPAP / NPL | ≥ 100% | 80-100% | < 80% | higher |
| Concentration | Pembiayaan terbesar / Total Pembiayaan | < 20% | 20-35% | > 35% | lower |
| Growth Modal | Pertumbuhan Modal YoY | ≥ 10% | 0-10% | < 0% | higher |
| SHU Ratio | SHU / Modal Sendiri | ≥ 10% | 5-10% | < 5% | higher |

### 4. KPI Measurement

```
kpi_measurement:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── kpi_id                UUID (FK → kpi_definition)
│
├── ── Measurement ──
├── measurement_date      DATE
├── measurement_period    VARCHAR         ← "2026-Q1", "2026-04", "2026"
├── value                 DECIMAL(20,4)
├── previous_value        DECIMAL(20,4) (nullable)
├── change_pct            DECIMAL(10,4) (nullable)
│
├── ── Status ──
├── health_status         ENUM (healthy, warning, critical)
├── trend                 ENUM (improving, stable, deteriorating)
│
├── ── Alert ──
├── alert_triggered       BOOLEAN DEFAULT false
├── alert_sent_at         TIMESTAMPTZ (nullable)
│
├── ── Audit ──
├── calculated_at         TIMESTAMPTZ
├── calculated_by         VARCHAR         ← "system_auto" atau user ID
└── created_at            TIMESTAMPTZ

UNIQUE(tenant_id, kpi_id, measurement_date)
```

### 5. Early Warning System

```
early_warning_alert:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── kpi_id                UUID (FK → kpi_definition)
├── measurement_id        UUID (FK → kpi_measurement)
│
├── ── Alert Info ──
├── alert_type            ENUM (threshold_breach, trend_deteriorating,
│                                sudden_change, consecutive_decline)
├── severity              ENUM (info, warning, critical)
├── message               TEXT
│
├── ── Response ──
├── assigned_to           UUID (FK → user, nullable)
├── corrective_action     TEXT (nullable)
├── action_deadline       DATE (nullable)
├── action_status         ENUM (pending, in_progress, completed, overdue)
├── resolved_at           TIMESTAMPTZ (nullable)
│
├── ── Escalation ──
├── escalated             BOOLEAN DEFAULT false
├── escalated_to          UUID (FK → user, nullable)
├── escalated_at          TIMESTAMPTZ (nullable)
│
├── ── Audit ──
├── triggered_at          TIMESTAMPTZ
└── created_at            TIMESTAMPTZ
```

### 6. Stress Testing

```
stress_test_scenario:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Scenario Info ──
├── scenario_name         VARCHAR
├── scenario_type         ENUM (baseline, adverse, severely_adverse)
├── description           TEXT
│
├── ── Parameters ──
├── parameters            JSONB
│   ├── npl_increase_pct  DECIMAL        ← Simulasi kenaikan NPL
│   ├── withdrawal_pct    DECIMAL        ← Simulasi penarikan massal
│   ├── interest_rate_change DECIMAL     ← Simulasi perubahan suku bunga
│   └── economic_growth   DECIMAL        ← Simulasi pertumbuhan ekonomi
│
├── ── Results ──
├── projected_car         DECIMAL(10,4) (nullable)
├── projected_npl         DECIMAL(10,4) (nullable)
├── projected_roa         DECIMAL(10,4) (nullable)
├── projected_roe         DECIMAL(10,4) (nullable)
├── capital_adequate      BOOLEAN (nullable)
├── survives_scenario     BOOLEAN (nullable)
│
├── ── Audit ──
├── test_date             DATE
├── created_at            TIMESTAMPTZ
└── created_by            UUID (FK → user)
```

**Default stress scenarios:**

| Scenario | NPL Increase | Withdrawal Shock | Rate Change | Expected Outcome |
|---|---|---|---|---|
| Baseline | +0% | +5% | ±0% | Current position |
| Mild Stress | +3% | +10% | +1% | CAR stays above threshold |
| Moderate Stress | +5% | +20% | +2% | CAR drops but stays positive |
| Severe Stress | +10% | +35% | +3% | Capital adequacy at risk |
| Extreme | +20% | +50% | +5% | Survival threatened |

### 7. Composite Health Score

```
health_score_calculation:
├── Each KPI scored 0-100 based on position vs threshold
├── Weighted by category:
│   ├── Capital:       20%
│   ├── Asset Quality: 25%
│   ├── Management:    15%
│   ├── Earnings:      20%
│   └── Liquidity:     20%
│
├── Composite Score → Rating:
│   ├── 85-100: AA (Sangat Sehat)
│   ├── 70-84:  A  (Sehat)
│   ├── 55-69:  BBB (Cukup Sehat)
│   ├── 40-54:  BB  (Kurang Sehat)
│   ├── 25-39:  B   (Tidak Sehat)
│   └── 0-24:   C   (Sangat Tidak Sehat)
│
└── Reporting frequency:
    ├── Dashboard: real-time (calculated daily)
    ├── Board report: monthly
    ├── Regulatory: quarterly
    └── RAT: annually
```

### 8. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| CAR calculation | Standard risk-weighted | + risk-weighted syariah instruments |
| NPL terminology | NPL (Non-Performing Loan) | NPF (Non-Performing Financing) |
| Profitability | ROA/ROE standard | + indicator kepatuhan syariah |
| Liquidity risk | Standard | + projected profit-sharing rate risk |
| Stress testing | Standard scenarios | + scenario perubahan nisbah |
| Health score | CAMEL framework | CAMELS + Sharia compliance score |

### 9. Vernon _rels dan _data Structure

**KPI Measurement _rels:**
```json
{
  "tenant_id": "018f...",
  "kpi_id":    "018f..."
}
```

**KPI Measurement _data:**
```json
{
  "kpi": {
    "code":    "NPL_GROSS",
    "name":    "NPL Gross Ratio",
    "category": "asset_quality"
  },
  "health": {
    "status": "warning",
    "trend":  "deteriorating"
  }
}
```

**SyncEngine triggers:**
- `EndOfDayBalancesEvent` → recalculate liquidity KPIs
- `MonthlyFinancialCloseEvent` → recalculate all KPIs
- `KpiAlertTriggeredEvent` → send notification via K022

### 10. Authorization — RBAC

| Permission | Manager | Bendahara | Ketua | Pengawas | Admin |
|---|---|---|---|---|---|
| View health dashboard | v | v | v | v | v |
| View KPI details | v | v | v | v | v |
| Configure KPI thresholds | - | - | - | - | v |
| Run stress test | - | v | v | - | v |
| Acknowledge alert | v | v | v | - | v |
| Assign corrective action | - | v | v | - | v |
| View trend analysis | v | v | v | v | v |
| Export health report | v | v | v | v | v |
| Configure composite weights | - | - | - | - | v |

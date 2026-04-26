# ADR-K038: Cooperative Dissolution Process

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Proses pembubaran koperasi diatur dalam UU Koperasi No. 25/1992 Pasal 52-60. Meskipun jarang terjadi, sistem harus mendukung proses ini jika suatu saat diperlukan — baik secara sukarela maupun oleh pemerintah.

## Decision

### 1. Dissolution Types

```
dissolution_type:
├── VOLUNTARY           ← Keputusan RAT (≥ 2/3 suara)
├── GOVERNMENT_ORDER    ← Keputusan pemerintah (pelanggaran berat)
├── MERGER              ← Penggabungan ke koperasi lain
├── SPLIT               ← Pecah menjadi 2+ koperasi
└── BANKRUPTCY          ← Dinyatakan pailit oleh pengadilan
```

### 2. Data Model

```
dissolution_process:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Dissolution Info ──
├── dissolution_type      ENUM (voluntary, government_order, merger, split, bankruptcy)
├── rat_meeting_id        UUID (nullable, FK → rat_meeting, ref K026)
├── reason                TEXT
├── effective_date        DATE
│
├── ── Liquidation Team ──
├── liquidator_ids        UUID[]          ← Tim likuidasi (ditunjuk RAT/pemerintah)
├── supervisor_id         UUID (nullable, FK → user, pengawas)
│
├── ── Stages ──
├── stage                 ENUM (announced, claim_period, asset_liquidation,
│                                debt_settlement, member_distribution, final_report, closed)
├── claim_deadline        DATE (nullable)  ← Batas waktu klaim kreditor
│
├── ── Financial Summary ──
├── total_assets          BIGINT (nullable)
├── total_liabilities     BIGINT (nullable)
├── net_equity            BIGINT (nullable)
├── distribution_per_member BIGINT (nullable)
│
├── ── Status ──
├── status                ENUM (initiated, in_progress, completed, cancelled)
├── final_report_doc_id   UUID (nullable)
├── closed_at             TIMESTAMPTZ (nullable)
│
├── ── Audit ──
├── initiated_at          TIMESTAMPTZ
├── initiated_by          UUID (FK → user)
└── created_at            TIMESTAMPTZ
```

### 3. Dissolution Workflow

```
Voluntary Dissolution:
┌──────────────────────────────────────────────────┐
│ 1. RAT Decision (≥ 2/3 suara hadir)              │
│    ├── Proposal pembubaran                        │
│    ├── Appoint liquidator team                    │
│    └── Set claim period (min 60 hari)             │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ 2. Announcement (Koran + pemberitahuan ke anggota)│
│    ├── Pemberitahuan ke kreditor                  │
│    ├── Pemberitahuan ke Dinas Koperasi            │
│    └── Pemberitahuan ke OJK (jika LKM)            │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ 3. Claim Period (60 hari)                         │
│    ├── Kreditor mengajukan klaim                  │
│    ├── Verifikasi klaim oleh liquidator           │
│    └── Dispute resolution                         │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ 4. Asset Liquidation                               │
│    ├── Jual aset koperasi                          │
│    ├── Collect outstanding loans                   │
│    ├── Settle semua kewajiban                      │
│    └── Pay creditors per priority                  │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ 5. Member Distribution                             │
│    ├── Hitung net equity per anggota               │
│    ├── Distribute simpanan + surplus               │
│    └── Close semua rekening                        │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ 6. Final Report & Closure                         │
│    ├── Laporan ke Dinas Koperasi                   │
│    ├── Laporan ke RAT (final)                      │
│    ├── Archive semua records (5 tahun)             │
│    └── System decommission                         │
└──────────────────────────────────────────────────┘
```

**Priority of settlement (urutan pembayaran):**
1. Biaya likuidasi
2. Gaji/biaya staff yang belum dibayar
3. Utang ke pemerintah (pajak)
4. Utang ke kreditor (bank, supplier)
5. Simpanan anggota
6. SHU/surplus (pro-rata)

### 4. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| DPS role | Tidak ada | DPS verifikasi settlement sesuai syariah |
| Dana sosial | Masuk ke distribution | Disalurkan sesuai ketentuan (tidak dibagi ke anggota) |
| Priority of settlement | Standard | + zakat/infaq harus disalurkan dulu |

### 5. Vernon _rels dan _data Structure

Standard relational model — dissolution process jarang terjadi, tidak perlu Vernon caching.

### 6. Authorization — RBAC

| Permission | Ketua | Pengawas | Liquidator | Admin |
|---|---|---|---|---|
| Initiate dissolution | v (via RAT) | - | - | v |
| Process claims | - | v (verify) | v | v |
| Liquidate assets | - | - | v | v |
| Distribute to members | - | v (witness) | v | v |
| Final report | - | v | v | v |
| Close system | - | - | - | v |

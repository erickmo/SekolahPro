# ADR-K025: Cooperative Governance & Internal Controls

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

UU Koperasi No. 25/1992 Pasal 35-45 mewajibkan setiap koperasi memiliki struktur tata kelola yang terdiri dari Rapat Anggota, Pengurus (Board), dan Pengawas (Supervisory). Untuk koperasi yang beroperasi sebagai LKM/BMT di bawah pengawasan OJK, tata kelola ini harus didukung oleh sistem internal controls yang kuat.

Saat ini tidak ada ADR yang mengatur:
- Struktur organisasi koperasi dan peran masing-masing jabatan
- Pembagian wewenang dan authority limits
- Prosedur internal audit dan oversight
- Kebijakan conflict of interest
- Segregation of duties dalam operasional keuangan
- Approval hierarchy untuk berbagai jenis transaksi

Tanpa governance framework yang jelas, sistem rentan terhadap fraud, abuse of power, dan non-compliance regulasi.

## Decision

### 1. Organizational Structure

```
cooperative_governance_structure:
├── RAPAT ANGGOTA (RAT)         ← K026: Highest authority
│   └── Annual meeting + special meetings
│
├── PENGGURS (Board of Directors)
│   ├── Ketua (Chairman)        ← Strategic decisions, external representation
│   ├── Sekretaris (Secretary)  ← Administration, correspondence, compliance
│   └── Bendahara (Treasurer)   ← Financial oversight, budget approval
│
├── PENGGAWAS (Supervisory Board)
│   ├── Ketua Pengawas          ← Internal audit coordination
│   └── Anggota Pengawas        ← Periodic audit execution
│
├── MANAJER (Operational Manager)
│   └── Day-to-day operations
│
└── DEWAN PENGGAWAS SYARIAH (DPS)  ← Only for BMT mode
    └── Sharia compliance oversight
```

### 2. Position & Role Management

```
governance_position:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah, person holding the position)
│
├── ── Position Info ──
├── position_type         ENUM (ketua, sekretaris, bendahara, pengawas_ketua,
│                                pengawas_anggota, manajer, dps_ketua, dps_anggota,
│                                teller, supervisor, admin)
├── position_level        ENUM (strategic, management, operational)
├── term_start            DATE           ← Mulai menjabat
├── term_end              DATE           ← Akhir masa jabatan
├── term_number           INT            ← Periode ke-berapa (1, 2, 3...)
│
├── ── Status ──
├── status                ENUM (active, resigned, removed, expired)
├── appointed_by          ENUM (rat, board_resolution, manager)
├── appointment_doc_id    UUID (FK → document, notulens/sk pengangkatan)
│
├── ── Authority ──
├── max_approval_amount   BIGINT         ← Batas approval (IDR, 0 = unlimited)
├── can_disburse          BOOLEAN DEFAULT false
├── can_reverse           BOOLEAN DEFAULT false
├── can_waive_penalty     BOOLEAN DEFAULT false
├── can_write_off         BOOLEAN DEFAULT false
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
├── updated_at            TIMESTAMPTZ
├── created_by            UUID (FK → user)
└── updated_by            UUID (FK → user)

UNIQUE(tenant_id, position_type, status) WHERE status = 'active'
-- Satu posisi hanya boleh diisi satu orang aktif
```

### 3. Authority Matrix — Segregation of Duties

**Transaksi Keuangan:**

| Transaction Type | Teller | Supervisor | Manager | Bendahara | Ketua |
|---|---|---|---|---|---|
| Setoran tabungan ≤ 10 juta | v | - | - | - | - |
| Setoran tabungan > 10 juta | v | v (verify) | - | - | - |
| Penarikan tabungan ≤ 5 juta | v | - | - | - | - |
| Penarikan tabungan > 5 juta | v | v (approve) | - | - | - |
| Penarikan tabungan > 25 juta | v | v (verify) | v (approve) | - | - |
| Penarikan tabungan > 100 juta | v | v (verify) | v (review) | v (approve) | - |
| Disbursement pinjaman ≤ 10 juta | v | v | - | - | - |
| Disbursement pinjaman > 10 juta | - | v | v | - | - |
| Disbursement pinjaman > 50 juta | - | - | v | v (approve) | - |
| Reversal transaksi | - | v | v (approve) | - | - |
| Write-off pinjaman | - | - | v (propose) | v (review) | v (approve) |
| Adjust balance | - | - | - | v | v (approve) |
| Cash opname | v (count) | v (witness) | - | - | - |

**Prinsip segregation of duties:**
- **Maker-Checker**: Setiap transaksi di atas threshold harus ada maker + checker berbeda
- **No self-approval**: Tidak boleh menyetujui transaksi sendiri
- **Four-eyes principle**: Transaksi besar membutuhkan minimal 2 approval
- **Rotation**: Teller tidak boleh bertugas di posisi yang sama > 6 bulan berturut-turut

### 4. Internal Audit Framework

```
internal_audit:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Audit Info ──
├── audit_type            ENUM (routine, special, compliance, forensic)
├── audit_period_start    DATE
├── audit_period_end      DATE
├── auditor_ids           UUID[] (FK → governance_position, pengawas)
├── scope                 TEXT            ← Ruang lingkup audit
│
├── ── Findings ──
├── findings              JSONB           ← Structured findings
│   ├── finding_id        STRING
│   ├── severity          ENUM (critical, high, medium, low, info)
│   ├── description       TEXT
│   ├── evidence_refs     STRING[]
│   └── recommendation    TEXT
│
├── ── Status ──
├── status                ENUM (planned, in_progress, completed, follow_up)
├── completed_at          TIMESTAMPTZ
├── presented_to_rat      BOOLEAN DEFAULT false   ← Apakah sudah dipresentasikan ke RAT
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

**Audit schedule:**
- **Routine audit**: Minimal 1x per semester oleh Pengawas
- **Cash audit**: Random/opname kas minimal 1x per bulan
- **Compliance audit**: 1x per tahun (regulatory compliance check)
- **Special audit**: Atas permintaan RAT atau jika ada indikasi irregularitas

### 5. Conflict of Interest Policy

```
conflict_of_interest:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── reporter_id           UUID (FK → governance_position)
│
├── ─── Disclosure ───
├── disclosure_type       ENUM (family_relationship, business_interest,
│                                financial_interest, gift_received, other)
├── related_party_name    VARCHAR
├── related_party_id      UUID (FK → nasabah, nullable)
├── description           TEXT
│
├── ── Assessment ──
├── severity              ENUM (none, low, medium, high)
├── resolution            ENUM (none_needed, recusal, monitoring, escalation)
├── resolved_by_id        UUID (FK → governance_position)
├── resolved_at           TIMESTAMPTZ
│
├── ── Audit ──
├── disclosed_at          TIMESTAMPTZ
└── created_at            TIMESTAMPTZ
```

**Aturan:**
- Pengurus/Pengawas **WAJIB** mengungkapkan konflik kepentingan sebelum memutuskan hal terkait
- Transaksi ke keluarga pengurus harus mendapat approval pihak independen
- Penerimaan hadiah > Rp 500.000 wajib dilaporkan
- Pelanggaran kebijakan CoI → sanksi sesuai AD/ART koperasi

### 6. Operational Controls

**Activity Logging:**
- Semua operasi CRUD pada data keuangan wajib di-log dengan actor, timestamp, old/new value
- Log tidak bisa dihapus (immutable audit trail)
- Log disimpan minimal 5 tahun (sesuai ketentuan arsip keuangan)

**Access Controls:**
- Session timeout: 15 menit idle (teller), 30 menit (back-office)
- Password policy: minimal 8 karakter, harus ada huruf besar, kecil, angka, simbol
- Password rotation: setiap 90 hari
- Failed login lockout: 5x percobaan gagal → lock 30 menit

**Cash Controls:**
- Maksimal kas di tangan teller: sesuai limit per tenant
- Opname kas wajib saat pergantian shift
- Selisih kas > Rp 50.000 wajib dilaporkan ke supervisor
- Selisih kas > Rp 500.000 wajib dilaporkan ke manajer

### 7. Dual-Mode Governance

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Struktur pengurus | Ketua, Sekretaris, Bendahara | Sama + DPS wajib |
| Dewan Pengawas Syariah | Tidak ada | **WAJIB** — minimal 1 anggota |
| Audit syariah | Tidak ada | DPS melakukan audit kepatuhan syariah |
| Approval pembiayaan | Manager + Bendahara | Manager + Bendahara + DPS (di atas threshold) |
| Produk baru | Approval Pengurus | Approval Pengurus + DPS |
| Laporan keuangan | PSAK/SAK ETAP | PSAK Syariah + DPS review |
| Konflik kepentingan | AD/ART koperasi | AD/ART + prinsip syariah |

### 8. Vernon _rels dan _data Structure

**Position _rels:**
```json
{
  "tenant_id":    "018f...",
  "nasabah_id":   "018f..."
}
```

**Position _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "term_info": {
    "term_start":  "2026-01-01",
    "term_end":    "2028-12-31",
    "term_number": 2
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di semua position nasabah tersebut
- `PositionStatusChangedEvent` → trigger access revocation untuk posisi yang expired/removed

### 9. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Bendahara | Ketua | Pengawas |
|---|---|---|---|---|---|---|
| View own position info | v | v | v | v | v | v |
| View all positions | - | - | v | v | v | v |
| Create operational position (teller, supervisor) | - | - | v | - | - | - |
| Create management position | - | - | - | v | v (approve) | - |
| Create strategic position | - | - | - | - | v (RAT) | - |
| Remove position | - | - | - | v | v | - |
| Configure authority limits | - | - | - | v | v | - |
| View audit trail | - | v | v | v | v | v |
| Create internal audit | - | - | - | - | - | v |
| View CoI disclosures | - | - | v | v | v | v |
| Submit CoI disclosure | v | v | v | v | v | v |
| Resolve CoI | - | - | - | v | v | - |

### 10. Consequences

**Keuntungan:**
- Compliance terhadap UU Koperasi dan regulasi OJK
- Segregation of duties mencegah fraud dan abuse
- Audit trail mendukung transparansi dan akuntabilitas
- Clear authority matrix menghindari ambiguitas operasional

**Risiko:**
- Kompleksitas konfigurasi authority matrix untuk koperasi kecil
- Perlu training untuk pengurus/pengawas yang tidak familiar dengan sistem
- Overhead operasional untuk approval multi-level

**Mitigasi:**
- Template authority matrix untuk koperasi kecil/medium/besar
- Simplified mode untuk koperasi dengan < 100 anggota (single approval)
- Guided setup wizard saat first-time configuration

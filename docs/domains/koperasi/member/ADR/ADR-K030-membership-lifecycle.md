# ADR-K030: Membership Lifecycle Management

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

K001 (Nasabah) mengatur proses registrasi dan approval anggota baru, tetapi tidak mencakup **lifecycle lengkap** keanggotaan koperasi:

- **Exit sukarela**: anggota mengundurkan diri dan mencairkan simpanan
- **Ekskumul**: pengeluaran anggota karena pelanggaran AD/ART
- **Transfer**: pindah antar koperasi (jika memungkinkan)
- **Kematian**: pengurusan ahli waris dan pencairan
- **Suspension**: pembekuan sementara keanggotaan
- **Inactive/dormant**: anggota yang tidak aktif dalam periode tertentu

UU Koperasi No. 25/1992 Pasal 19-24 mengatur hak dan kewajiban anggota serta prosedur keluar/dikeluarkan. Tanpa lifecycle management:
- Simpanan anggota keluar tidak terproses dengan benar
- Hak suara di RAT tidak di-update
- Ahli waris tidak bisa mencairkan simpanan almarhum
- Status keanggotaan ambigu (aktif atau tidak)

## Decision

### 1. Membership Status Lifecycle

```
Membership Status Flow:

  APPLIED ──→ APPROVED ──→ ACTIVE ──┬──→ SUSPENDED ──→ ACTIVE (reinstated)
              │                      │        │
              │                      │        └──→ EXPELLED
              │                      │
              │                      ├──→ RESIGNING ──→ RESIGNED
              │                      │        (clearing period)
              │                      │
              │                      ├──→ DECEASED ──→ SETTLING ──→ SETTLED
              │                      │
              │                      └──→ DORMANT ──→ ACTIVE (reactivated)
              │
              └── REJECTED

  SPECIAL:
  ACTIVE ──→ TRANSFERRED_OUT ──→ (moved to another cooperative)
  EXTERNAL ──→ TRANSFERRED_IN (accepted from another cooperative)
```

### 2. Status Definitions

```
membership_status:
├── APPLIED           ← Baru mendaftar, menunggu approval
├── APPROVED          ← Disetujui, belum bayar simpanan pokok
├── ACTIVE            ← Anggota penuh, sudah bayar simpanan pokok
├── SUSPENDED         ← Dibekukan sementara (kasus tertentu)
├── DORMANT           ← Tidak aktif > 12 bulan tanpa transaksi
├── RESIGNING         ← Proses pengunduran diri (clearing period)
├── RESIGNED          ← Selesai keluar, simpanan sudah dicairkan
├── EXPELLED          ← Dikeluarkan karena pelanggaran
├── DECEASED          ← Anggota meninggal dunia
├── SETTLING          ← Proses pencairan ahli waris
├── SETTLED           ← Ahli waris sudah menerima
├── TRANSFERRED_OUT   ← Pindah ke koperasi lain
└── TRANSFERRED_IN    ← Masuk dari koperasi lain
```

### 3. Data Model — Lifecycle Events

```
membership_lifecycle_event:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
│
├── ── Event Info ──
├── event_type            ENUM (approved, activated, suspended, reinstated,
│                                dormant_detected, reactivated, resignation_submitted,
│                                resignation_approved, resigned, expulsion_initiated,
│                                expulsion_approved, expelled, deceased_reported,
│                                settlement_initiated, settlement_completed, settled,
│                                transfer_out, transfer_in)
├── from_status           ENUM (nullable, status sebelumnya)
├── to_status             ENUM (status baru)
│
├── ── Details ──
├── reason                TEXT
├── initiated_by          UUID (FK → user, nullable)
├── approved_by           UUID (FK → user, nullable)
├── supporting_doc_ids    UUID[] (nullable, FK → document)
│
├── ── Financial Impact ──
├── simpanan_pokok_refund BOOLEAN (nullable)
├── simpanan_wajib_refund BOOLEAN (nullable)
├── refund_amount         BIGINT (nullable)
├── refund_status         ENUM (pending, processing, completed, hold)
├── outstanding_loans     BOOLEAN (nullable)   ← Apakah ada pinjaman outstanding
├── loan_settlement_plan  ENUM (none, full_repayment, restructuring, write_off)
│
├── ── Audit ──
├── event_date            DATE
├── effective_date        DATE            ← Tanggal efektif perubahan status
├── created_at            TIMESTAMPTZ
└── created_by            UUID (FK → user)
```

### 4. Resignation Process (Keluar Sukarela)

```
Resignation Workflow:
┌──────────────────────────────────────────────────────┐
│ Step 1: Pengajuan (Hari ke-1)                        │
│ ├── Anggota mengajukan surat pengunduran diri        │
│ ├── Sistem cek: ada pinjaman outstanding?             │
│ │   ├── Ya → Harus lunasi atau restructure dulu      │
│ │   └── Tidak → Lanjut ke clearing                   │
│ └── Status → RESIGNING                               │
└────────────┬─────────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────────┐
│ Step 2: Clearing Period (Hari ke-1 s/d H+30)        │
│ ├── Hitung semua simpanan (pokok + wajib + tabungan) │
│ ├── Hitung semua kewajiban (pinjaman, denda, dll)    │
│ ├── Hitung SHU yang belum dibagikan (pro-rata)       │
│ ├── Hitung biaya administrasi keluar (sesuai AD/ART) │
│ └── Net refund = simpanan - kewajiban - admin fee    │
└────────────┬─────────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────────┐
│ Step 3: Approval (H+7 s/d H+14)                     │
│ ├── Supervisor/Manager review perhitungan            │
│ ├── Jika ada dispute → negosiasi                     │
│ └── Approved → lanjut pencairan                      │
└────────────┬─────────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────────┐
│ Step 4: Settlement (H+14 s/d H+30)                  │
│ ├── Jurnal: Debit simpanan pokok/wajib               │
│ │           Credit kas/bank                           │
│ ├── Tutup semua rekening                              │
│ ├── Revoke akses sistem                              │
│ └── Status → RESIGNED                                │
└──────────────────────────────────────────────────────┘
```

**Perhitungan refund:**
```
Refund = Simpanan Pokok
       + Simpanan Wajib (accumulated)
       + Saldo Tabungan
       + SHU belum dibagikan (pro-rata berdasarkan bulan aktif)
       ────────────────────
       - Pinjaman outstanding (jika ada)
       - Denda tertunggak (jika ada)
       - Biaya administrasi keluar (sesuai AD/ART)
       ────────────────────
       = Net Refund Amount
```

### 5. Expulsion Process (Ekskumul)

```
Expulsion Workflow:
┌──────────────────────────────────────────────────┐
│ Step 1: Pelanggaran Teridentifikasi              │
│ ├── Jenis pelanggaran dicatat                     │
│ ├── Bukti dikumpulkan                             │
│ └── Reporter identity bisa anonim                 │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 2: Due Process (H+1 s/d H+14)              │
│ ├── Anggota diberi tahu secara tertulis           │
│ ├── Anggota berhak membela diri (hearing)         │
│ ├── Pengawas melakukan investigasi                │
│ └── Temuan didokumentasikan                       │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 3: Keputusan                                │
│ ├── Pengurus memutuskan berdasarkan temuan        │
│ ├── Jika ekskumul: approval Ketua + 1 Pengawas   │
│ ├── Anggota berhak banding ke RAT                 │
│ └── Status → EXPELLED                             │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 4: Financial Settlement                     │
│ ├── Simpanan pokok: dikembalikan (UU Koperasi)    │
│ ├── Simpanan wajib: dikembalikan                  │
│ ├── Pinjaman outstanding: harus dilunasi           │
│ ├── SHU: forfeited (tidak dibagikan)              │
│ └── Proses sama dengan resignation settlement     │
└──────────────────────────────────────────────────┘
```

**Alasan ekskumul (sesuai AD/ART umum):**
- Melanggar AD/ART koperasi
- Merugikan koperasi secara finansial (fraud, penggelapan)
- Tidak memenuhi kewajiban keanggotaan setelah peringatan
- Melakukan tindakan yang merusak reputasi koperasi

### 6. Death / Inheritance Settlement

```
Death Settlement Workflow:
┌──────────────────────────────────────────────────┐
│ Step 1: Laporan Kematian                         │
│ ├── Keluarga melaporkan kematian                  │
│ ├── Upload surat kematian / akta kematian         │
│ ├── Status → DECEASED                             │
│ └── Semua akun dibekukan (freeze)                 │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 2: Ahli Waris Verification                  │
│ ├── Ahli waris sesuai K001 (ahli_waris data)     │
│ ├── Jika belum terdaftar: upload surat waris/nota  │
│ ├── Verifikasi identitas ahli waris               │
│ └── Status → SETTLING                             │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 3: Financial Calculation                    │
│ ├── Total simpanan (pokok + wajib + tabungan)    │
│ ├── SHU pro-rata (jika belum dibagikan)           │
│ ├── Pinjaman outstanding (jika ada)               │
│ │   ├── Jika ada asuransi jiwa → klaim           │
│ │   └── Jika tidak → ahli waris tanggung jawab   │
│ └── Net settlement = simpanan - kewajiban         │
└────────────┬─────────────────────────────────────┘
             │
             v
┌──────────────────────────────────────────────────┐
│ Step 4: Disbursement                             │
│ ├── Jurnal pencairan ke ahli waris                │
│ ├── Tutup semua rekening                          │
│ └── Status → SETTLED                              │
└──────────────────────────────────────────────────┘
```

**Prioritas ahli waris (sesuai hukum Indonesia):**
1. Janda/duda + anak-anak
2. Janda/duda (tanpa anak)
3. Anak-anak (tanpa janda/duda)
4. Orang tua
5. Saudara kandung
6. Penerima wasiat (max 1/3 harta, untuk BMT)

### 7. Dormant Management

```
dormant_rules:
├── DORMANT_TRIGGER
│   ├── Tidak ada transaksi selama 12 bulan
│   ├── Tidak hadir di 2 RAT berturut-turut
│   └── Dan tidak ada simpanan wajib yang dibayar selama 6 bulan
│
├── DORMANT_NOTIFICATION
│   ├── H-30: Peringatan via WA/SMS "Rekening akan dormant"
│   ├── H-7: Peringatan terakhir
│   └── H+0: Status → DORMANT
│
├── DORMANT_EFFECTS
│   ├── Tidak berhak voting di RAT
│   ├── Tidak bisa mengajukan pinjaman baru
│   ├── Tabungan tetap aman (tidak hangus)
│   └── Admin fee dormant bisa dikenakan (sesuai AD/ART)
│
└── REACTIVATION
    ├── Anggota datang dan melakukan transaksi
    ├── Bayar simpanan wajib yang tertunggak
    └── Status → ACTIVE
```

### 8. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| SHU pro-rata saat keluar | Dihitung berdasarkan bulan aktif | Dihitung berdasarkan modal + transaksi |
| Ahli waris prioritas | KUHPerdata | Faraidh (hukum Islam) |
| Biaya admin keluar | Sesuai AD/ART | Sesuai AD/ART + DPS approval |
| Ekskumul alasan | Standard | + pelanggaran prinsip syariah |
| Simpanan pokok refund | Full refund | Full refund |
| Wasiat | Sesuai hukum umum | Max 1/3 harta (untuk non-ahli waris) |
| Pinjaman almarhum | Ahli waris tanggung | + cek asuransi takaful |

### 9. Vernon _rels dan _data Structure

**Lifecycle event _rels:**
```json
{
  "tenant_id":   "018f...",
  "nasabah_id":  "018f..."
}
```

**Lifecycle event _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "membership": {
    "status":        "resigning",
    "joined_at":     "2024-03-15",
    "months_active": 24
  },
  "financial_summary": {
    "simpanan_pokok": 1000000,
    "simpanan_wajib": 2400000,
    "tabungan":       3500000,
    "pinjaman_outstanding": 0,
    "net_refund":     6900000
  }
}
```

**SyncEngine triggers:**
- `NasabahStatusChangedEvent` → update `_data.membership.status` di semua cache
- `LifecycleFinancialCalculatedEvent` → update `_data.financial_summary`

### 10. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin | Ketua |
|---|---|---|---|---|---|
| Submit resignation request | - | - | v | v | v |
| Process resignation clearing | v | v | v | v | - |
| Approve resignation | - | - | v | v | - |
| Initiate expulsion | - | - | - | v | v |
| Approve expulsion | - | - | - | v | v |
| Report death | v | v | v | v | v |
| Process death settlement | - | v | v | v | - |
| Approve death settlement | - | - | v | v | - |
| Mark dormant | v (auto) | v | v | v | - |
| Reactivate dormant | - | v | v | v | - |
| View lifecycle history | v | v | v | v | v |

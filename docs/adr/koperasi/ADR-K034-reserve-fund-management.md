# ADR-K034: Reserve Fund Management (Dana Cadangan)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

UU Koperasi No. 25/1992 Pasal 45 dan Anggaran Dasar koperasi mewajibkan pembentukan **dana cadangan** yang dialokasikan minimal dari sisa hasil usaha (SHU). Dana cadangan berfungsi sebagai:

- Buffer untuk menutupi kerugian koperasi
- Jaminan keamanan simpanan anggota
- Modal untuk pengembangan usaha
- Kepatuhan regulasi (Dinas Koperasi, OJK untuk LKM)

Saat ini K016 (SHU) menyebutkan cadangan sebagai bagian dari distribusi SHU, tetapi tidak mengatur:
- Alokasi otomatis dana cadangan
- Investasi dana cadangan yang aman
- Prosedur penarikan dana cadangan
- Pelaporan dana cadangan ke regulator

## Decision

### 1. Reserve Fund Types

```
reserve_fund_type:
├── STATUTORY_RESERVE (Dana Cadangan Wajib)
│   ├── Sumber: minimal 25% dari SHU setiap tahun (sesuai AD/ART)
│   ├── Tujuan: menutupi kerugian, jaminan simpanan
│   ├── Batas: sampai mencapai minimal 25% dari simpanan anggota
│   └── Tidak bisa dibagi ke anggota — hanya untuk menutupi loss
│
├── GENERAL_RESERVE (Dana Cadangan Umum)
│   ├── Sumber: alokasi dari SHU di atas statutory minimum
│   ├── Tujuan: pengembangan usaha, emergency
│   └── Approval penarikan: Pengurus + Pengawas
│
├── INVESTMENT_RESERVE (Dana Cadangan Investasi)
│   ├── Sumber: alokasi dari SHU atau surplus operasional
│   ├── Tujuan: investasi aman untuk generate income
│   └── Hanya instrumen yang diizinkan regulasi
│
└── SPECIAL_RESERVE (Dana Cadangan Khusus)
    ├── Sumber: alokasi khusus dari RAT
    ├── Tujuan: tujuan tertentu (contoh: pembangunan gedung, IT system)
    └── Approval: RAT (pembentukan dan penarikan)
```

### 2. Data Model

```
reserve_fund:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Fund Info ──
├── fund_type             ENUM (statutory, general, investment, special)
├── fund_name             VARCHAR
├── description           TEXT
│
├── ── Balance ──
├── current_balance       BIGINT DEFAULT 0
├── target_balance        BIGINT (nullable)    ← Target yang ingin dicapai
├── min_balance           BIGINT (nullable)    ← Minimum yang harus ada
│
├── ── Configuration ──
├── shu_allocation_pct    DECIMAL(5,2)         ← % dari SHU yang dialokasikan
├── max_balance_pct       DECIMAL(5,2) (nullable) ← % dari simpanan total (target cap)
├── auto_allocate         BOOLEAN DEFAULT true
│
├── ── Investment (untuk investment reserve) ──
├── investment_instrument ENUM (bank_deposit, sbn, mutual_fund, none)
├── investment_maturity   DATE (nullable)
├── investment_rate       DECIMAL(5,4) (nullable) ← Expected return rate
│
├── ── Status ──
├── status                ENUM (active, frozen, closed)
├── frozen_reason         TEXT (nullable)
│
├── ── Audit ──
├── created_at            TIMESTAMPTZ
├── updated_at            TIMESTAMPTZ
├── created_by            UUID (FK → user)
└── updated_by            UUID (FK → user)
```

```
reserve_fund_transaction:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── fund_id               UUID (FK → reserve_fund)
│
├── ── Transaction Info ──
├── transaction_type      ENUM (allocation, investment_return, withdrawal,
│                                loss_coverage, transfer_in, transfer_out,
│                                adjustment)
├── amount                BIGINT
├── balance_after         BIGINT
│
├── ── Reference ──
├── shu_id                UUID (nullable, FK → shu, jika alokasi dari SHU)
├── jurnal_id             UUID (nullable, FK → jurnal_entry)
├── reference_type        VARCHAR (nullable)
├── reference_id          UUID (nullable)
│
├── ── Approval ──
├── requires_approval     BOOLEAN DEFAULT true
├── approved_by           UUID (nullable, FK → user)
├── approved_at           TIMESTAMPTZ (nullable)
├── rejection_reason      TEXT (nullable)
│
├── ── Audit ──
├── transaction_date      DATE
├── created_at            TIMESTAMPTZ
└── created_by            UUID (FK → user)
```

### 3. Automatic SHU Allocation

```
SHU Allocation to Reserve Fund (Annual, saat SHU disetujui di RAT):

SHU Total
├── 1. Dana Cadangan Wajib (statutory)
│   ├── Minimum: 25% dari SHU (sesuai AD/ART, bisa lebih)
│   ├── Sampai target tercapai: 25% dari total simpanan anggota
│   └── Jika target sudah tercapai: bisa dialokasikan ke tujuan lain
│
├── 2. Dana Cadangan Umum (general)
│   ├── Alokasi: sisa setelah statutory (sesuai keputusan RAT)
│   └── Tidak ada batas maksimal
│
├── 3. Dana Cadangan Investasi (investment)
│   ├── Alokasi: jika RAT menyetujui
│   └── Disimpan di instrumen aman (deposito, SBN)
│
├── 4. SHU untuk anggota
│   ├── Jasa Modal: proporsional terhadap simpanan
│   ├── Jasa Usaha: proporsional terhadap transaksi
│   └── Dana Pendidikan: sesuai AD/ART
│
└── 5. Dana Sosial (untuk BMT)
    ├── Ta'zir fund
    └── Zakat/infaq (ref K018)
```

### 4. Withdrawal Rules

```
Withdrawal Rules by Fund Type:
┌──────────────────────────────────────────────────────────┐
│ STATUTORY RESERVE:                                       │
│ ├── Hanya untuk menutupi kerugian                        │
│ ├── Approval: Ketua + Bendahara + Pengawas               │
│ ├── Jika balance < 25% simpanan: WAJIB top-up dari SHU   │
│ └── Tidak boleh untuk operasional atau dibagi ke anggota │
│                                                          │
│ GENERAL RESERVE:                                         │
│ ├── Untuk pengembangan usaha atau emergency               │
│ ├── Approval: Ketua + Bendahara                           │
│ ├── Max withdrawal per event: sesuai keputusan Pengurus   │
│ └── Jika > 25% dari balance: perlu approval Pengawas      │
│                                                          │
│ INVESTMENT RESERVE:                                      │
│ ├── Penarikan sesuai maturity instrumen                   │
│ ├── Premature withdrawal: approval Bendahara + Ketua      │
│ └── Return otomatis di-reinvest                          │
│                                                          │
│ SPECIAL RESERVE:                                         │
│ ├── Penarikan sesuai tujuan yang disetujui RAT            │
│ ├── Approval: Ketua + sekurangnya 1 Pengawas              │
│ └── Jika tujuan berubah: harus approval RAT               │
└──────────────────────────────────────────────────────────┘
```

### 5. Investment Instruments (for Investment Reserve)

```
allowed_instruments:
├── BANK_DEPOSIT (Deposito Bank)
│   ├── Minimum rating: BBB+ (Pefindo) atau setara
│   ├── Max tenor: 12 bulan
│   ├── Max per bank: 25% dari investment reserve
│   └── Return: fixed rate
│
├── SBN (Surat Berharga Negara)
│   ├── Jenis: SUN, SPN, SRI
│   ├── Max tenor: 5 tahun
│   └── Return: fixed/coupon
│
├── MUTUAL_FUND (Reksa Dana)
│   ├── Hanya: money market fund atau fixed income fund
│   ├── Minimum rating: AAA
│   ├── Max: 20% dari investment reserve
│   └── Return: floating
│
└── FORBIDDEN (TIDAK BOLEH)
    ├── Saham (equity) — terlalu risky
    ├── Crypto — tidak diizinkan regulasi
    ├── Real estate (langsung) — illiquid
    └── Foreign currency speculation
```

### 6. Journal Entries

```
Reserve Fund Journal Entries:

1. Alokasi dari SHU:
   Debit:  SHU Belum Distribusi (ekuitas)
   Credit: Dana Cadangan Wajib (ekuitas)
   Credit: Dana Cadangan Umum (ekuitas)

2. Withdrawal untuk tutup kerugian:
   Debit:  Dana Cadangan Wajib (ekuitas)
   Credit: Akumulasi Kerugian (ekuitas)

3. Investment placement:
   Debit:  Investasi - Deposito/SBN (aset)
   Credit: Kas/Bank

4. Investment return:
   Debit:  Kas/Bank
   Credit: Pendapatan Investasi (income)

5. Withdrawal dari investment:
   Debit:  Kas/Bank
   Credit: Investasi - Deposito/SBN (aset)
```

### 7. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Statutory reserve | 25% SHU minimum | 25% SHU minimum |
| Investment instruments | Deposito, SBN | Deposito syariah, Sukuk |
| Return treatment | Pendapatan bunga → income | Bagi hasil → income |
| Loss coverage | Standard | + DPS harus verifikasi |
| Withdrawal approval | Ketua + Bendahara + Pengawas | + DPS untuk withdrawal besar |
| Social fund | Tidak ada | Terpisah dari reserve fund |

### 8. Vernon _rels dan _data Structure

**Fund _rels:**
```json
{
  "tenant_id": "018f..."
}
```

**Fund _data:**
```json
{
  "fund": {
    "type":    "statutory",
    "name":    "Dana Cadangan Wajib"
  },
  "summary": {
    "current_balance":    150000000,
    "target_balance":     250000000,
    "pct_of_simpanan":    18.5,
    "target_pct":         25.0,
    "shortfall":          100000000
  }
}
```

**SyncEngine triggers:**
- `ShuApprovedEvent` (ref K016) → trigger automatic allocation to reserve funds
- `ReserveFundTransactionCreatedEvent` → update `_data.summary`
- `InvestmentMaturityEvent` → notify bendahara untuk reinvest/withdraw

### 9. Authorization — RBAC

| Permission | Bendahara | Ketua | Pengawas | Admin |
|---|---|---|---|---|
| View reserve fund balance | v | v | v | v |
| View fund transaction history | v | v | v | v |
| Configure allocation percentage | - | - | - | v |
| Process SHU allocation (auto) | v | v | - | v |
| Request withdrawal | v | v | - | v |
| Approve withdrawal (statutory) | v | v | v (witness) | v |
| Approve withdrawal (general) | v | v | - | v |
| Investment placement | v | v | - | v |
| View investment performance | v | v | v | v |
| Adjust fund balance | - | - | - | v |

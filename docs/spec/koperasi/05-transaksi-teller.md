# Transaksi & Teller Session

Transaksi adalah **core engine** dari seluruh pergerakan keuangan di koperasi/BMT. Setiap aliran dana harus tercatat sebagai transaksi immutable dan auditable. Teller session mengaitkan transaksi tunai dengan pertanggungjawaban kas fisik per teller per shift.

---

## ADR References

- **ADR-K011** — Transaksi Rekening & Non-Rekening
- **ADR-K012** — Teller Session
- **ADR-K013** — Money Denomination (Pecahan Uang)

---

## Domain Entities

### Tabel `transaksi`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| tenant_id / branch_id | UUID | |
| rekening_id | UUID (FK) | |
| nasabah_id | UUID (FK) | |
| idempotency_key | UUID NOT NULL UNIQUE/tenant | Mencegah duplikat transaksi |
| reference_number | VARCHAR UNIQUE/tenant | Format: `TRX-YYYYMMDD-BRANCH-NNNNNNNN` |
| transaction_type | ENUM | setoran, penarikan, angsuran, pencairan, biaya_admin, denda, koreksi |
| transaction_subtype | VARCHAR | tunai, transfer, bagi_hasil, bunga, pembukaan, penutupan, system |
| description | TEXT | |
| amount | NUMERIC(15,2) | Positif = credit, negatif = debit |
| balance_before | NUMERIC(15,2) | Snapshot saldo sebelum |
| balance_after | NUMERIC(15,2) | Snapshot saldo sesudah |
| is_cash | BOOLEAN NOT NULL | |
| teller_session_id | UUID nullable | FK → teller_session jika cash |
| transfer_ref | VARCHAR nullable | Shared ref untuk pasangan transfer |
| transfer_pair_id | UUID nullable | FK → transaksi pasangan |
| is_reversed | BOOLEAN NOT NULL DEFAULT false | Satu-satunya field yang boleh di-update |
| reversal_of | UUID nullable | FK → original yang di-reverse |
| reversal_reason | TEXT nullable | |
| batch_id | UUID nullable | FK → transaction_batch |
| specimen_verified | BOOLEAN NOT NULL | |
| approval_status | ENUM | none, pending, approved, rejected |
| _rels / _data | JSONB | Vernon pattern |

**Database constraints:**
- `CHECK (balance_after = balance_before + amount)` — integrity check
- `UNIQUE (tenant_id, reference_number)`
- `UNIQUE (tenant_id, idempotency_key)` — idempotency deduplication
- Trigger: reject UPDATE except on `is_reversed` field
- Trigger: reject DELETE

### Tabel `kas_transaksi` (Non-Rekening)

| Field | Tipe | Keterangan |
|-------|------|------------|
| transaction_type | ENUM | kas_masuk, kas_keluar, transfer_kas |
| category | VARCHAR | gaji, atk, listrik, sewa, lain_lain |
| description | TEXT NOT NULL | |
| amount | NUMERIC(15,2) | Selalu positif |
| direction | ENUM | in, out |
| teller_session_id | UUID nullable | |
| source_branch_id / destination_branch_id | UUID nullable | Untuk transfer_kas |

### Tabel `transaction_batch`

| Field | Tipe | Keterangan |
|-------|------|------------|
| batch_type | ENUM | simpanan_wajib, bagi_hasil, admin_fee, denda |
| description | TEXT | |
| period | VARCHAR | `YYYY-MM` |
| total_items | INTEGER | |
| success_count / failed_count | INTEGER | |
| total_amount | NUMERIC(15,2) | |
| status | ENUM | draft, pending, approved, processing, completed, cancelled |

### Tabel `teller_session`

| Field | Tipe | Keterangan |
|-------|------|------------|
| id | UUID v7 (PK) | |
| tenant_id / branch_id | UUID | |
| teller_user_id | UUID (FK) | |
| session_date | DATE | |
| session_number | INTEGER | Sequence per branch per hari |
| opening_amount | NUMERIC(15,2) | Kas awal yang diterima |
| opened_by_supervisor | UUID (FK) | |
| opening_confirmed_at | TIMESTAMPTZ | |
| total_cash_in | NUMERIC(15,2) | Running total kas masuk |
| total_cash_out | NUMERIC(15,2) | Running total kas keluar |
| transaction_count | INTEGER | |
| expected_cash | NUMERIC(15,2) | = opening + in - out |
| actual_cash | NUMERIC nullable | Diisi saat close |
| variance | NUMERIC nullable | = expected - actual |
| variance_explanation | TEXT nullable | Wajib jika variance ≠ 0 |
| variance_approved_by | UUID nullable | |
| status | ENUM | open, suspended, closed |
| opened_at | TIMESTAMPTZ | |
| suspended_at / resumed_at | TIMESTAMPTZ nullable | |
| closed_at | TIMESTAMPTZ nullable | |
| is_force_closed | BOOLEAN | |
| force_close_reason | TEXT nullable | |
| _rels / _data | JSONB | Vernon pattern |

### Tabel `teller_session_denomination`

| Field | Tipe | Keterangan |
|-------|------|------------|
| teller_session_id | UUID (FK) | |
| phase | ENUM | opening, closing |
| denomination_value | INTEGER | Nilai pecahan: 100000, 50000, dll |
| quantity | INTEGER | Jumlah lembar/keping |
| subtotal | NUMERIC(15,2) | = denomination_value × quantity |

### Tabel `denominasi_master`

| Field | Tipe | Keterangan |
|-------|------|------------|
| tenant_id | UUID (FK) | |
| value | INTEGER | 100000, 50000, 20000, 10000, 5000, 2000, 1000, 500, 200, 100 |
| currency | VARCHAR DEFAULT 'IDR' | |
| type | ENUM | banknote, coin |
| label | VARCHAR | Display: "Rp 100.000 Kertas" |
| sort_order | INTEGER | Urutan tampil: besar ke kecil |
| is_active | BOOLEAN | |

### Tabel `vault_count`

| Field | Tipe | Keterangan |
|-------|------|------------|
| branch_id | UUID (FK) | |
| count_date | DATE | |
| counted_by | UUID (FK) | |
| verified_by | UUID nullable | |
| total_counted | NUMERIC(15,2) | |
| total_expected | NUMERIC(15,2) | Calculated vault balance |
| variance | NUMERIC(15,2) | |
| status | ENUM | draft, verified, investigated |

---

## Business Rules

### Transaksi (K011)

1. Transaksi bersifat **immutable** — tidak bisa diedit atau dihapus setelah commit
2. Satu-satunya field yang boleh di-update: `is_reversed`
3. Koreksi dilakukan via **transaksi reversal** — bukan edit
4. Insert transaksi dan update saldo rekening harus dalam **satu database transaction** (atomik)
5. `SELECT ... FOR UPDATE` mencegah race condition pada saldo rekening
6. **Deadlock prevention**: lock rekening dengan UUID lebih kecil terlebih dahulu, baru yang lebih besar
7. `balance_after = balance_before + amount` — di-enforce via database CHECK constraint

### Tipe Transaksi

| Tipe | Direction | Mempengaruhi Saldo |
|------|-----------|-------------------|
| SETORAN | CREDIT | Ya, balance + |
| PENARIKAN | DEBIT | Ya, balance - |
| ANGSURAN | CREDIT ke pinjaman | Ya, outstanding - |
| PENCAIRAN | DEBIT dari pinjaman | Ya, outstanding + |
| BIAYA_ADMIN | DEBIT | Ya, balance - |
| DENDA | DEBIT | Ya, balance - |
| KOREKSI | CREDIT/DEBIT | Ya |
| KAS_KELUAR | - | Tidak |
| KAS_MASUK | - | Tidak |
| TRANSFER_KAS | - | Tidak |

### Validation Pipeline (7 Layer)

8. **V1**: Rekening status — ACTIVE: semua OK; FROZEN: hanya CREDIT; CLOSED: semua ditolak
9. **V2**: Available balance check — `available_balance >= amount` untuk DEBIT
10. **V3**: KYC limit check — SUM transaksi bulanan nasabah ≤ kyc_monthly_limit (basic: default Rp 10 juta)
11. **V4**: Product-level limit — daily limit, monthly limit, single transaction max
12. **V5**: Minimum balance maintenance — `balance - amount >= min_balance`
13. **V6**: Specimen verification — penarikan > threshold wajib verifikasi; > manager_threshold perlu approval Manager
14. **V7**: Teller session check — cash transaction wajib ada session aktif
15. Error response mengembalikan **semua** validasi yang gagal (bukan hanya yang pertama)

### Idempotency

16. Setiap request wajib menyertakan `idempotency_key` (UUID client-generated)
17. Jika key sudah ada: return transaksi existing (HTTP 200, bukan error)
18. Deduplication window: 24 jam

### Transfer Antar Rekening

19. Transfer = dua transaksi atomik (debit sumber + kredit tujuan) dalam satu database transaction
20. Kedua rekening harus dalam **tenant yang sama** — cross-tenant transfer tidak diizinkan
21. Source harus ACTIVE; destination bisa ACTIVE atau FROZEN
22. Kedua transaksi share `transfer_ref` dan dihubungkan via `transfer_pair_id`
23. KYC limit dihitung untuk **kedua** nasabah

### Reversal

24. Reversal **selalu** membutuhkan approval **Manager+** — tidak ada auto-reversal
25. Original transaction ditandai `is_reversed = true`
26. Transaksi reversal baru dibuat dengan `reversal_of = original_id`
27. Satu transaksi hanya bisa di-reverse **sekali**
28. Partial reversal tidak diizinkan

### Batch Transactions

29. Setiap item batch diproses sebagai **transaksi individual** dengan referensi ke batch_id
30. Item yang gagal validasi di-skip — **tidak** rollback seluruh batch
31. Batch summary report: total items, success, failed, total amount
32. Batch membutuhkan approval **Manager+** sebelum eksekusi

### Teller Session (K012)

33. Satu teller hanya bisa punya **satu session aktif** pada satu waktu
34. Transaksi tunai **wajib** dikaitkan dengan teller session aktif
35. Opening session membutuhkan **dua orang**: Supervisor (serahkan kas) + Teller (terima dan konfirmasi)
36. Kas awal **wajib dihitung per denominasi** — bukan hanya total nominal
37. Saat teller inisiasi close, sistem **memblokir transaksi baru** untuk session tersebut
38. Teller **wajib** input denominasi kas fisik saat penutupan
39. Auto-lock: jika teller idle selama 15 menit (configurable), session otomatis suspend
40. Hanya **teller yang sama** yang bisa resume session-nya

### Variance Handling

| Selisih | Tindakan |
|---------|----------|
| = 0 | Auto-close |
| ≤ threshold_minor (default Rp 10.000) | Supervisor approve + explanation wajib |
| ≤ threshold_major (default Rp 100.000) | Manager approve + investigation |
| > threshold_major | Manager + full investigation + flag teller + notifikasi Admin |

> **Catatan: Dua Level Threshold Selisih Kas**
> 
> Terdapat dua set threshold selisih kas yang berbeda dalam sistem:
> 
> **Threshold K012 (Penutupan Sesi Teller Individual)** — berlaku di modul ini:
> - ≤ Rp 10.000: Supervisor approve
> - ≤ Rp 100.000: Manager approve  
> - > Rp 100.000: Manager approve + investigasi + notifikasi Admin
> 
> **Threshold K025 (Internal Audit Kas Cabang)** — berlaku di modul governance/audit:
> - > Rp 50.000: wajib lapor ke Supervisor (untuk rekap selisih seluruh teller)
> - > Rp 500.000: wajib lapor ke Manager (untuk rekap harian kas cabang)
> 
> Kedua threshold ini **tidak saling menggantikan**. K012 adalah kontrol operasional
> per sesi teller. K025 adalah kontrol governance level cabang. Keduanya bisa berlaku
> serentak: seorang teller dengan selisih Rp 80.000 (K012 → Manager approve) sekaligus
> berkontribusi pada selisih cabang > Rp 50.000 (K025 → laporan Supervisor audit).

41. Explanation **wajib** jika selisih ≠ 0
42. Force close hanya oleh **Manager+** — flag `is_force_closed = true` + alasan wajib
43. Konsolidasi semua teller session per branch per hari → input kas harian (K014)

### Money Denomination (K013)

44. Denominasi Rupiah yang didukung: kertas (100rb, 50rb, 20rb, 10rb, 5rb, 2rb, 1rb) + logam (1rb, 500, 200, 100)
45. Rp 1.000 tersedia dalam 2 bentuk (kertas dan logam) — dicatat sebagai entry terpisah
46. Master denominasi **per tenant** — di-seed saat onboarding
47. Admin bisa tambah denominasi baru (future-proof untuk pecahan BI baru) atau nonaktifkan
48. Denominasi tidak bisa dihapus (soft deactivate) — data historis tetap valid
49. `SUM(subtotal)` denominasi opening harus = `opening_amount` — tidak match → recount
50. `SUM(subtotal)` denominasi closing menjadi `actual_cash` — selisih dengan `expected_cash` → variance handling

---

## Transaction Flow

### Setoran Tunai Tabungan

```
Teller input nominal
    │
    v
Validation Pipeline (V1-V7)
    │ semua pass
    v
BEGIN TRANSACTION
  SELECT rekening FOR UPDATE
  INSERT transaksi (type=SETORAN, is_cash=true)
  UPDATE rekening.balance += amount
  UPDATE teller_session.total_cash_in += amount
  Auto-generate jurnal (K015)
COMMIT
    │
    v
Generate receipt
Print/digital bukti transaksi
```

### Pembayaran Angsuran (Multi-Step Allocation)

```
Teller input nominal pembayaran
    │
    v
Identifikasi angsuran target (FIFO: terlama yang belum lunas)
    │
    v
Alokasi berdasarkan prioritas:
  1. Denda outstanding → allocated_penalty
  2. Bunga/margin tertunggak
  3. Pokok tertunggak
  4. Bunga/margin berjalan
  5. Pokok berjalan
  6. Sisa → overpayment handling
    │
    v
UPDATE angsuran.paid_* + payment_status
UPDATE pinjaman.outstanding_principal
UPDATE rekening.balance
INSERT transaksi (type=ANGSURAN)
Cek: semua angsuran paid? → COMPLETED
```

### Teller Session Lifecycle

```
Supervisor serahkan kas awal (hitung denominasi)
    │
    v
┌─────────────┐
│ Status: OPEN │ ← Bisa proses transaksi tunai
└──────┬───────┘
       │ suspend (istirahat/handover)
       v
┌──────────────────┐
│ Status: SUSPENDED │ ← Tidak bisa transaksi, kas di drawer
└──────┬───────────┘
       │ resume (teller kembali)
       v
┌─────────────┐
│ Status: OPEN │
└──────┬───────┘
       │ close (akhir shift)
       v
Hitung denominasi kas fisik
Sistem hitung expected_cash
Bandingkan → selisih
    │
    ├── selisih=0 → auto-close
    └── selisih≠0 → approval + explanation
           │
           v
┌──────────────┐
│Status: CLOSED │ ← Serahkan kas ke vault
└──────────────┘
```

---

## Daily Consolidation

```
Semua teller session CLOSED
    │
    v
┌──────────────────────────────────────┐
│ Branch Daily Summary:                 │
│ Total kas awal + inflow - outflow     │
│ = Expected closing balance            │
│                                       │
│ SUM(actual_cash semua teller)         │
│ + vault end-of-day balance            │
│ = Actual closing balance              │
│                                       │
│ Selisih → investigasi jika ≠ 0        │
└──────────────────────────────────────┘
    │
    v
Feed ke kas_harian (K014)
Generate laporan kas harian
```

---

## RBAC Summary

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Setoran/Penarikan tunai | v | v | v | v |
| Penarikan > specimen threshold (verifikasi) | v* | v | v | v |
| Penarikan > manager threshold | - | - | v | v |
| Transfer antar rekening | v | v | v | v |
| Pencairan pinjaman | - | - | v | v |
| Request reversal | v | v | v | v |
| Approve reversal | - | - | v | v |
| Batch create | - | v | v | v |
| Batch approve & execute | - | - | v | v |
| Kas keluar operasional | - | v | v | v |
| Transfer kas antar cabang | - | v** | v | v |
| Open teller session | v | - | - | - |
| Serahkan kas awal ke teller | - | v | v | v |
| Approve variance minor | - | v | v | v |
| Approve variance major | - | - | v | v |
| Force close session | - | - | v | v |
| Vault count | - | v | v | v |
| Manage denominasi master | - | - | - | v |

`*` Teller verifikasi specimen visual terlebih dahulu
`**` Supervisor membutuhkan approval Manager untuk transfer kas antar cabang

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Immutable transaksi | Integritas audit trail keuangan; regulasi koperasi/OJK melarang manipulasi catatan |
| Atomic balance update (1 DB transaction) | Konsistensi saldo; race condition menciptakan inkonsistensi yang tidak bisa di-recover |
| Idempotency key | Double-click teller, network retry → transaksi tidak ganda |
| 7-layer validation pipeline | Setiap layer mencegah jenis fraud/error berbeda; semua error dikembalikan sekaligus |
| Session teller per orang, bukan per branch | Cash accountability individual; selisih bisa di-trace ke teller spesifik |
| Denominasi wajib saat open/close | Deteksi manipulasi; standar operasional koperasi/bank |
| Force close → kas tetap dihitung per denominasi | Cash accountability tidak bisa di-bypass meskipun force close |
| Batch item failure tidak rollback seluruh batch | Resilience; satu nasabah bermasalah tidak menghambat ratusan lainnya |
| Deadlock prevention via consistent lock ordering | Dua teller proses transfer saling berlawanan → deadlock tanpa aturan ini |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `rekening` (K002) | Transaksi → Rekening | Setiap transaksi rekening memperbarui saldo |
| `nasabah` (K001) | Transaksi ← Nasabah | KYC limit check, specimen verification |
| `produk` (K003) | Transaksi ← Produk | Product-level limits |
| `angsuran` (K008) | Transaksi → Angsuran | Link payment ke jadwal angsuran |
| `denda` (K009) | Transaksi → Denda | Pembayaran denda via transaksi |
| `teller_session` (K012) | Transaksi ↔ Session | Cash transaction linked to session |
| `kas_harian` (K014) | Transaksi → Kas | Posisi kas harian dari aggregasi transaksi |
| `jurnal` (K015) | Transaksi → Jurnal | Auto-generate journal entry per transaction |
| `toko_penjualan` (K019) | Transaksi ← POS | Debit tabungan saat bayar di toko/kantin |

# Toko/Kantin & HR Payroll Deduction

Koperasi sekolah umumnya mengoperasikan toko (ATK, seragam, buku) dan kantin (makanan, minuman) sebagai sumber pendapatan tambahan. Integrasi HR memungkinkan potongan gaji otomatis untuk guru/staf koperasi. Keduanya adalah modul opsional yang bisa diaktifkan/dinonaktifkan per tenant.

---

## ADR References

- **ADR-K019** — Koperasi Toko & Kantin
- **ADR-K020** — Payroll Integration & Auto-Deduction

---

## Domain Entities

### Tabel `toko_lokasi`

| Field | Tipe | Keterangan |
|-------|------|------------|
| branch_id | UUID (FK) | |
| name | VARCHAR | e.g., `Kantin Putra`, `Toko Utama` |
| location_type | ENUM | store, canteen |
| is_active | BOOLEAN | |

### Tabel `toko_kategori`

| Field | Tipe | Keterangan |
|-------|------|------------|
| name | VARCHAR | |
| parent_id | UUID nullable | Self-reference untuk sub-kategori |
| product_type | ENUM | store_item, canteen_item |
| sort_order | INTEGER | |

### Tabel `toko_produk`

| Field | Tipe | Keterangan |
|-------|------|------------|
| branch_id | UUID (FK) | Produk bisa berbeda per lokasi |
| name | VARCHAR | |
| sku | VARCHAR UNIQUE/tenant+branch | |
| barcode | VARCHAR nullable | |
| category_id | UUID (FK) | |
| cost_price | NUMERIC(15,2) | Harga beli/HPP |
| sell_price | NUMERIC(15,2) | Harga jual |
| margin_percentage | NUMERIC(5,2) | Calculated: (sell-cost)/cost × 100 |
| current_stock | INTEGER | Cache stok — source of truth adalah SUM movement |
| minimum_stock | INTEGER | Reorder point alert threshold |
| unit | VARCHAR | pcs, pack, box, kg, liter |
| product_type | ENUM | store_item, canteen_item |
| is_daily_menu | BOOLEAN | Khusus kantin — menu harian |
| image_url | VARCHAR nullable | |
| is_active | BOOLEAN | |

### Tabel `toko_penjualan`

| Field | Tipe | Keterangan |
|-------|------|------------|
| location_id | UUID (FK) | |
| receipt_number | VARCHAR UNIQUE/tenant | |
| sale_date | TIMESTAMPTZ | |
| sale_type | ENUM | store, canteen |
| nasabah_id | UUID nullable | Jika bayar dari tabungan |
| customer_name | VARCHAR nullable | Jika cash/non-nasabah |
| subtotal | NUMERIC(15,2) | |
| discount_amount | NUMERIC(15,2) | |
| total_amount | NUMERIC(15,2) | |
| payment_method | ENUM | cash, tabungan_debit, ewallet_debit, mixed |
| cash_amount / change_amount | NUMERIC(15,2) | |
| tabungan_debit_amount | NUMERIC(15,2) | |
| transaction_id | UUID nullable | FK → transaksi K011 jika debit tabungan |
| status | ENUM | completed, voided |
| voided_at / voided_by / void_reason | TIMESTAMPTZ / UUID / TEXT | |

### Tabel `toko_penjualan_item`

| Field | Tipe | Keterangan |
|-------|------|------------|
| penjualan_id | UUID (FK) | |
| produk_id | UUID (FK) | |
| quantity | INTEGER | |
| unit_price | NUMERIC(15,2) | |
| discount | NUMERIC(15,2) | |
| subtotal | NUMERIC(15,2) | |

### Tabel `toko_stok_movement`

| Field | Tipe | Keterangan |
|-------|------|------------|
| produk_id | UUID (FK) | |
| movement_type | ENUM | stock_in, stock_out, adjustment, opname, transfer |
| quantity | INTEGER | Positif = masuk, negatif = keluar |
| stock_before / stock_after | INTEGER | |
| reference_type | ENUM | purchase, sale, adjustment, opname, transfer |
| reference_id | UUID | FK → tabel referensi sesuai type |
| notes | TEXT nullable | |

### Tabel `toko_supplier`

| Field | Tipe | Keterangan |
|-------|------|------------|
| name | VARCHAR | |
| contact_person | VARCHAR nullable | |
| phone / email | VARCHAR nullable | |
| address | TEXT nullable | |
| is_active | BOOLEAN | |

### Tabel `toko_pembelian` (Purchase Order)

| Field | Tipe | Keterangan |
|-------|------|------------|
| supplier_id | UUID (FK) | |
| po_number | VARCHAR UNIQUE/tenant | |
| po_date | DATE | |
| total_amount | NUMERIC(15,2) | |
| total_items | INTEGER | |
| status | ENUM | draft, pending, approved, received, cancelled |

### Tabel `toko_opname`

| Field | Tipe | Keterangan |
|-------|------|------------|
| opname_number | VARCHAR UNIQUE/tenant | |
| opname_date | DATE | |
| status | ENUM | open, pending, approved, rejected |

### Tabel `toko_opname_item`

| Field | Tipe | Keterangan |
|-------|------|------------|
| opname_id | UUID (FK) | |
| produk_id | UUID (FK) | |
| system_stock | INTEGER | Stok di sistem |
| physical_stock | INTEGER | Stok fisik hasil hitung |
| difference | INTEGER | physical - system |

### Tabel `toko_meal_plan` (Asrama/Boarding)

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | Siswa boarding |
| plan_type | ENUM | full_board, half_board, custom |
| meals_included | JSONB | `["breakfast", "lunch", "dinner"]` |
| monthly_fee | NUMERIC(15,2) | |
| auto_deduct | BOOLEAN | |
| deduct_from_rekening | UUID nullable | Rekening tabungan untuk auto-deduct |
| deduct_day | INTEGER | Tanggal potong bulanan (1-28) |
| is_active | BOOLEAN | |

---

## Domain Entities — HR Payroll

### Tabel `payroll_batch`

| Field | Tipe | Keterangan |
|-------|------|------------|
| batch_number | VARCHAR UNIQUE/tenant | |
| period_month | INTEGER | 1-12 |
| period_year | INTEGER | |
| upload_source | ENUM | csv, xlsx, api |
| total_employees | INTEGER | |
| matched_employees | INTEGER | |
| unmatched_employees | INTEGER | |
| total_deduction_amount | NUMERIC(15,2) | |
| skipped_deductions | INTEGER | |
| status | ENUM | uploaded, calculated, pending_approval, approved, executing, executed, completed, failed |
| raw_data_url | VARCHAR nullable | URL ke file upload asli |

### Tabel `payroll_deduction`

| Field | Tipe | Keterangan |
|-------|------|------------|
| batch_id | UUID (FK) | |
| nasabah_id | UUID (FK) | |
| employee_id | VARCHAR | NIP/ID dari file payroll |
| employee_name | VARCHAR | |
| net_salary | NUMERIC(15,2) | |
| deduction_type | ENUM | simpanan_wajib, angsuran_pinjaman, tabungan_berencana, tabungan_reguler, iuran_anggota, custom |
| authorization_id | UUID (FK) | FK → payroll_authorization |
| rekening_id | UUID (FK) | Target rekening |
| amount | NUMERIC(15,2) | |
| priority_order | INTEGER | |
| status | ENUM | calculated, approved, executed, skipped, failed |
| skip_reason | TEXT nullable | e.g., `INSUFFICIENT_SALARY` |
| transaction_id | UUID nullable | FK → transaksi K011, setelah executed |

### Tabel `payroll_authorization`

| Field | Tipe | Keterangan |
|-------|------|------------|
| nasabah_id | UUID (FK) | |
| deduction_type | ENUM | |
| rekening_id | UUID (FK) | Target rekening |
| amount | NUMERIC(15,2) | |
| effective_date | DATE | |
| end_date | DATE nullable | null = berlaku sampai dicabut |
| status | ENUM | active, revoked, expired |
| authorization_doc_url | VARCHAR nullable | Scan surat kuasa tertulis |

### Tabel `payroll_mapping` (Employee-Nasabah)

| Field | Tipe | Keterangan |
|-------|------|------------|
| employee_id | VARCHAR | NIP dari payroll |
| nasabah_id | UUID (FK) | |
| identity_number | VARCHAR nullable | NIK |
| is_verified | BOOLEAN | |

---

## Business Rules

### Toko & Kantin (K019)

1. Toko/kantin adalah **unit bisnis opsional** — diaktifkan per tenant via feature flag
2. Setiap penjualan mengurangi stok secara **atomik** dalam satu database transaction
3. Jika stok < quantity yang diminta → transaksi ditolak (stok negatif tidak diizinkan)
4. Pembayaran `TABUNGAN_DEBIT` melakukan:
   - Cek available_balance ≥ total
   - Cek **spending limit** harian nasabah
   - Cek category restrictions
   - Cek time restrictions
   - Jika violation → pesan error spesifik
5. Spending limit siswa default: Rp 30.000/hari (configurable per school_relation_type)
6. Orang tua bisa override spending limit untuk anaknya via portal
7. Void penjualan hanya oleh **Supervisor+** — dalam batas waktu yang dikonfigurasi (default: 24 jam)
8. Kasir harus dalam **sesi kasir aktif** (mirip teller session K012)
9. Reconciliation harian: cached stock vs SUM movements — alert jika selisih

### Stok Management

10. `current_stock` di tabel produk adalah **cache** — source of truth adalah SUM `stok_movement`
11. Reorder point alert: jika `current_stock <= minimum_stock` → notifikasi ke Supervisor
12. Stock opname membutuhkan approval **Manager+** — selisih hasil opname menghasilkan adjustment movement
13. Purchase Order: Supervisor buat → Manager approve → terima barang → stok otomatis bertambah (stock_in movement)
14. Transfer stok antar lokasi menggunakan movement type `transfer`

### Margin & Kontribusi SHU

15. Profit toko/kantin masuk ke **pendapatan koperasi** — berkontribusi ke SHU tahunan
16. Margin per item = `sell_price - cost_price`
17. Laporan profit toko/kantin: harian, mingguan, bulanan
18. Pencatatan akuntansi menggunakan jurnal otomatis sesuai mapping K015

### Meal Plan (Boarding/Asrama)

19. Full board = 3 makan (breakfast, lunch, dinner); half board = 2 makan (pilih kombinasi)
20. Auto-deduct bulanan dari tabungan siswa — dicatat sebagai transaksi K011
21. Jika saldo tidak cukup → notifikasi ke orang tua

---

### Payroll Integration (K020)

22. Sistem koperasi adalah **consumer** payroll — tidak membangun payroll engine
23. Sistem **tidak menyimpan data gaji secara permanen** — dihapus setelah retention period (default: 12 bulan)
24. Integrasi via CSV upload (manual) atau API push (otomatis) — configurable per tenant
25. Setiap jenis potongan memerlukan **otorisasi tertulis** (signed authorization) dari nasabah
26. Otorisasi angsuran pinjaman **tidak bisa dicabut** selama pinjaman masih aktif
27. Revoke otorisasi simpanan wajib memerlukan persetujuan **Manager** (mandatory)

### Prioritas Potongan (Default)

| Prioritas | Jenis Potongan |
|-----------|----------------|
| 1 | ANGSURAN_PINJAMAN (kewajiban hukum, prioritas tertinggi) |
| 2 | SIMPANAN_WAJIB (kewajiban keanggotaan) |
| 3 | TABUNGAN_BERENCANA (komitmen sukarela) |
| 4 | TABUNGAN_REGULER (sukarela) |
| 5 | IURAN_ANGGOTA (tahunan) |
| 6 | CUSTOM (terendah) |

28. Urutan prioritas **configurable per tenant**
29. Jika sisa gaji tidak cukup untuk potongan berikutnya → **skip** seluruhnya (bukan partial, by default)
30. Opsi partial deduction bisa diaktifkan per tenant

### Process Flow Payroll Batch

```
STEP 1: Upload data payroll (CSV/API oleh TU)
    │ Status: UPLOADED
    │
STEP 2: Auto-matching & calculation
    ├── Match by identity_number (NIK) — primary
    ├── Fallback: match by employee_id (NIP)
    ├── Unmatched → dilaporkan, tidak di-skip dengan error
    ├── Hitung semua potongan per employee
    ├── Apply priority jika gaji tidak mencukupi
    │ Status: CALCULATED
    │
STEP 3: Generate deduction report
    ├── Per employee: daftar potongan
    ├── Total per jenis potongan
    ├── Skipped deductions + alasan
    ├── Unmatched employees
    │ Status: PENDING_APPROVAL
    │
STEP 4: Manager approve batch
    │ Status: APPROVED
    │
STEP 5: System execute (background job)
    ├── Create transaksi K011 per potongan
    ├── Update saldo rekening tujuan
    ├── Update jadwal angsuran (jika angsuran pinjaman)
    │ Status: EXECUTED / COMPLETED
    │
STEP 6: Notifikasi ke nasabah (detail potongan)
         Konfirmasi ke TU/HR (summary batch)
```

### Idempotency Batch

31. Satu batch per `period_month + period_year + branch_id` — tidak bisa double-execute
32. Jika batch gagal di tengah jalan → retry — sistem cek mana yang sudah executed dan skip
33. Rollback seluruh batch dimungkinkan oleh Admin dalam **24 jam** — reverse semua transaksi

### Employee-Nasabah Matching

34. Matching utama: `identity_number` (NIK) dari payroll vs `nasabah.identity_number`
35. Fallback: `employee_id` (NIP) vs `payroll_mapping.employee_id`
36. Unmatched employees dilaporkan — bisa dibuat manual mapping oleh TU/Supervisor
37. Nasabah tidak di-auto-create dari data payroll (`auto_create = false`)

---

## Laporan

### Toko & Kantin

| Laporan | Frekuensi | Audience |
|---------|-----------|----------|
| Penjualan harian | Harian | Kasir, Supervisor |
| Stok movement | Harian | Supervisor |
| Profit margin per item | Mingguan | Manager |
| Best sellers / Slow moving | Mingguan/Bulanan | Manager |
| Stock opname variance | Per opname | Manager |
| Supplier purchase summary | Bulanan | Manager |
| Spending per siswa | Harian/Bulanan | Parent (via portal K023) |
| Meal plan status | Bulanan | Admin, Parent |

### Payroll Deduction

| Laporan | Audience |
|---------|----------|
| Deduction Report per Batch | Manager, TU |
| Summary per Deduction Type | Manager |
| Unmatched Employees | TU, Supervisor |
| Skipped Deductions | Manager, Nasabah |
| Monthly Deduction History | Nasabah (via portal) |
| Annual Summary | Nasabah, Manager |

---

## RBAC Summary

### Toko & Kantin

| Permission | Kasir | Supervisor | Manager | Admin |
|------------|-------|------------|---------|-------|
| Proses penjualan (POS) | v | v | v | v |
| Void penjualan | - | v | v | v |
| Manage katalog produk | - | v | v | v |
| Set harga jual | - | - | v | v |
| Buat PO | - | v | v | v |
| Approve PO | - | - | v | v |
| Buat session opname | - | v | v | v |
| Approve opname | - | - | v | v |
| Set spending limit | - | - | - | v |
| View laporan penjualan | v | v | v | v |
| View laporan profit | - | - | v | v |

### HR Payroll

| Permission | TU/Operator | Supervisor | Manager | Admin |
|------------|-------------|------------|---------|-------|
| Upload data payroll | v | v | v | v |
| View deduction report | v | v | v | v |
| Approve deduction batch | - | - | v | v |
| Execute batch | - | - | - | System |
| Create employee mapping | v | v | v | v |
| Manage authorization | - | v | v | v |
| Revoke auth simpanan wajib | - | - | v | v |
| Configure priority order | - | - | - | v |
| Retry failed batch | - | - | v | v |

---

## Key Decisions & Rationale

| Keputusan | Alasan |
|-----------|--------|
| Toko/kantin sebagai feature flag opsional | Tidak semua koperasi sekolah punya toko/kantin; overhead tanpa kebutuhan |
| Stok cache + SUM movement reconciliation | Pola yang sama dengan balance rekening (K002); konsistensi arsitektur |
| Atomic stock deduction pada POS | Mencegah overselling/stok negatif dalam kondisi concurrent |
| Spending limit per siswa | Kontrol orang tua; pencegahan jajan berlebihan; keamanan |
| Koperasi sebagai consumer payroll, bukan producer | Payroll adalah domain kompleks (pajak, BPJS); cukup konsumsi net salary |
| Otorisasi tertulis wajib untuk setiap potongan | Perlindungan nasabah; bukti persetujuan yang bisa diaudit |
| FIFO priority dalam batch execution | Kewajiban hukum (angsuran pinjaman) harus diprioritaskan |
| Tidak auto-accumulate skipped deductions | Akumulasi ke bulan berikutnya bisa memotong gaji terlalu besar tanpa persetujuan nasabah |
| Idempotency per period+branch | Mencegah double-potongan gaji untuk bulan yang sama |

---

## Integration Points

| Integrasi | Arah | Keterangan |
|-----------|------|------------|
| `transaksi` (K011) | Toko ↔ Transaksi | Debit tabungan saat POS; referensi ke penjualan |
| `rekening` (K002) | Toko ← Rekening | available_balance check; spending limit enforcement |
| `nasabah` (K001) | Toko ← Nasabah | Identifikasi pembeli, school_relation_type untuk spending limit |
| `jurnal` (K015) | Toko → Jurnal | Penjualan dan pembelian menghasilkan jurnal otomatis |
| `shu_periode` (K016) | Toko → SHU | Profit toko/kantin masuk ke pendapatan koperasi |
| `SekolahPro HR/Payroll` | Payroll ← HR | CSV/API import data gaji dari sistem HR sekolah |
| `angsuran` (K008) | Payroll → Angsuran | Deduction angsuran pinjaman update jadwal angsuran |
| `simpanan_wajib_billing` (K004) | Payroll → Billing | Deduction simpanan wajib mark billing sebagai PAID |
| `Portal Orang Tua` (K023) | Toko → Portal | Laporan spending siswa; set/lihat spending limit |
| `notifikasi` (K022) | Toko/Payroll → Notif | Meal plan gagal debit; skipped deduction → notif ke nasabah |

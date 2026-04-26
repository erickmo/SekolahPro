# 11 — Manajemen Kantin

Modul Kantin mengelola operasional makan di sekolah, khususnya boarding school dan pesantren di mana santri makan 3–4 kali sehari di lingkungan sekolah. Modul terbagi dua: manajemen menu dan jadwal makan (S036) serta transaksi dan penagihan (S037). Transparansi menu ke orang tua melalui Parent Portal (S042) adalah salah satu nilai utama modul ini.

---

## ADR References

| ADR | Judul | Pattern |
|-----|-------|---------|
| ADR-S036 | Canteen Management (Menu & Jadwal) | Vernon |
| ADR-S037 | Canteen Transaction & Billing | Vernon |

---

## Domain Entities

### Canteen Management (S036)

**`canteen_vendors`** — Master vendor/supplier:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(20) | Kode unik per tenant+company |
| `vendor_type` | VARCHAR(20) | `catering` (full-service), `supplier` (bahan baku), `internal` (dapur sendiri) |
| `contract_start` / `contract_end` | DATE | Periode kontrak (nullable) |
| `is_active` | BOOLEAN | Status vendor |

**`canteen_menus`** — Master item menu:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `code` | VARCHAR(20) | Kode unik menu |
| `category` | VARCHAR(20) | `main_course`, `side_dish`, `soup`, `dessert`, `snack`, `beverage`, `fruit` |
| `is_halal` | BOOLEAN | Default `true` — wajib untuk pesantren |
| `calories` | INT | Kalori (nullable — opsional) |
| `protein_gram`, `carbs_gram`, `fat_gram` | DECIMAL | Info nutrisi (nullable) |
| `allergens` | TEXT | Daftar alergen (free-text, nullable) |
| `vendor_id` | UUID | FK → canteen_vendors (nullable) |

**`canteen_meal_schedules`** — Jadwal menu harian:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `academic_year_id` | UUID | FK → academic_years |
| `schedule_date` | DATE | Tanggal jadwal |
| `meal_type` | VARCHAR(20) | `breakfast`, `lunch`, `dinner`, `snack` |
| `serving_time` | TIME | Jam penyajian (nullable) |
| `menu_items` | JSONB | Array menu_id yang disajikan |
| `estimated_portions` | INT | Estimasi porsi (> 0) |
| `actual_portions` | INT | Porsi aktual yang disajikan (nullable) |

### Canteen Transaction & Billing (S037)

**`canteen_student_plans`** — Meal plan / saldo per siswa per tahun ajaran:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK → students |
| `academic_year_id` | UUID | FK |
| `plan_type` | VARCHAR(20) | `flat_monthly`, `prepaid`, `per_meal` |
| `monthly_fee` | BIGINT | Hanya untuk flat_monthly (nullable) |
| `balance` | BIGINT | Saldo prepaid (default 0) |
| `total_topup` | BIGINT | Total top-up sepanjang tahun |
| `total_consumed` | BIGINT | Total yang telah dikonsumsi |
| `includes_breakfast/lunch/dinner/snack` | BOOLEAN | Meal mana yang termasuk dalam plan |

**`canteen_transactions`** — Transaksi makan harian:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `student_id` | UUID | FK |
| `plan_id` | UUID | FK → canteen_student_plans |
| `schedule_id` | UUID | FK → canteen_meal_schedules (nullable) |
| `transaction_date` | DATE | Tanggal transaksi |
| `meal_type` | VARCHAR(20) | `breakfast`, `lunch`, `dinner`, `snack` |
| `amount` | BIGINT | Biaya per meal (0 untuk flat_monthly) |
| `status` | VARCHAR(20) | `consumed`, `skipped`, `absent` |

**`canteen_topups`** — Pengisian saldo prepaid:

| Field | Tipe | Keterangan |
|-------|------|-----------|
| `plan_id` | UUID | FK → canteen_student_plans |
| `receipt_no` | VARCHAR(50) | Nomor kwitansi (unik) |
| `amount` | BIGINT | Nominal top-up (> 0) |
| `payment_method` | VARCHAR(20) | `cash`, `transfer`, `debit`, `qris`, `va` |
| `topup_date` | DATE | Tanggal top-up |

---

## Business Rules

### Menu & Jadwal

1. **Satu jadwal per (tenant, company, tanggal, meal_type).** Duplicate ditolak.
2. **Semua menu wajib halal secara default** (`is_halal = true`). Sekolah umum bisa mengubah, pesantren tidak.
3. **`menu_items` di jadwal menggunakan JSONB array** (denormalisasi) — trade-off simplicity vs normalized query karena query pattern selalu per-tanggal, bukan aggregate lintas jadwal.
4. **`actual_portions` dicatat setelah penyajian** untuk monitoring food waste.
5. **Bulk create jadwal mingguan** didukung — 4 meal × 7 hari = 28 records dalam satu request.

### Meal Plan & Transaksi

6. **Satu plan per siswa per tahun ajaran.** Unique constraint `(student_id, academic_year_id)`.
7. **Tiga model billing didukung:**
   - `flat_monthly`: biaya tetap per bulan, diintegrasikan ke invoice S009.
   - `prepaid`: saldo isi ulang, deducted per transaksi `consumed`.
   - `per_meal`: dihitung per porsi, ditagihkan bulanan via S009.
8. **Status `absent` tidak ditagihkan.** Santri yang izin pulang (S034 permission) diberi status `absent` — balance tidak berkurang.
9. **Status `skipped` tidak ditagihkan** untuk prepaid — santri sengaja tidak makan.
10. **Balance prepaid tidak boleh negatif.** Transaksi `consumed` yang melebihi balance ditolak (422).
11. **`balance`, `total_topup`, `total_consumed` adalah denormalisasi** yang harus dijaga konsisten via application logic setiap transaksi.
12. **`receipt_no` top-up harus unik** per tenant+company untuk audit trail.

### Integrasi Keuangan (S009)

13. **Flat_monthly:** biaya makan dijadikan `fee_type` di S009, invoice di-generate bulanan.
14. **Per_meal:** aggregate transaksi bulan ini menjadi basis invoice S009.
15. **Prepaid:** tidak masuk invoice S009 — dikelola sepenuhnya di domain kantin.

---

## Alur Workflow

### Setup Awal Tahun Ajaran

```
1. Admin buat meal plan untuk semua siswa (bulk)
   → pilih plan_type per siswa atau satu tipe untuk semua
2. Untuk prepaid: orang tua top-up saldo
3. Setup jadwal menu per minggu (bulk schedule)
```

### Pencatatan Harian

```
Staf kantin buka endpoint bulk transaksi per meal_type:
  → Input status per santri: consumed / skipped / absent
  → Untuk prepaid: balance langsung dikurangi per consumed
  → Untuk absent: dari data izin asrama (S034) — tidak ditagih
```

### Top-up Saldo (Prepaid)

```
Orang tua datang ke kantor / transfer → staf input top-up
  → balance ditambah
  → total_topup diupdate
  → kwitansi dicetak dengan receipt_no unik
```

---

## Key Decisions & Rationale

1. **`menu_items` di jadwal menggunakan JSONB** (bukan junction table) karena query selalu per-tanggal dan menu per meal selalu di-read as a whole. Junction table menambah complexity tanpa benefit signifikan untuk volume kecil.
2. **Informasi nutrisi opsional** — tidak semua sekolah memerlukan tracking kalori. Sekolah yang perlu bisa mengisi, yang tidak bisa lewati.
3. **Driver sebagai `internal` vendor** didukung — pesantren yang mengelola dapur sendiri tidak perlu vendor eksternal.
4. **Kantin tidak replace domain keuangan (S009)** — kantin mengelola tracking per-meal, S009 mengelola invoicing dan pembayaran formal.
5. **Tidak ada refund saldo prepaid di MVP** — enhancement masa depan.
6. **Integrasi kartu tap/NFC tidak ada di MVP** — terlalu bergantung pada infrastruktur hardware.

---

## Integration Points

### Internal

| Domain | Arah | Deskripsi |
|--------|------|-----------|
| ADR-S001 Students | → S037 | Data siswa untuk meal plan |
| S009 Student Finance | ← S037 | Flat_monthly dan per_meal di-bridge ke invoice S009 |
| S034 Dormitory Permission | → S037 | Santri yang izin pulang → status `absent` di transaksi |
| S042 Parent Portal | ← S036 | Orang tua lihat menu hari ini / minggu ini |
| S051 Payment Gateway | → S037 | Top-up saldo prepaid bisa melalui payment gateway |
| S044 Notification | ← S037 | Alert saldo rendah ke orang tua |

### Eksternal

| Sistem | Keterangan |
|--------|-----------|
| **Payment Provider** (via S051) | Top-up saldo via VA, QRIS, transfer untuk model prepaid |

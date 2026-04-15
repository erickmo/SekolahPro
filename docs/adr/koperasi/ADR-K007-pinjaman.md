# ADR-K007: Pinjaman / Pembiayaan

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Pinjaman/pembiayaan adalah domain **paling kompleks dan berisiko tinggi** dalam sistem koperasi/BMT. Setiap rupiah yang dicairkan adalah dana anggota yang harus dikelola dengan prudent. Sistem harus mengakomodasi:

- **Multi-level approval**: pencairan dana membutuhkan analisis kredit dan otorisasi bertingkat
- **Dual-mode**: Pinjaman berbasis bunga (konvensional) vs Pembiayaan berbasis akad syariah (BMT) — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)
- **Credit scoring**: kelayakan nasabah harus dievaluasi berdasarkan penghasilan, simpanan, dan riwayat
- **Penjamin (Guarantor)**: beberapa pinjaman memerlukan penjamin sebagai mitigasi risiko
- **NPL management**: klasifikasi kolektibilitas per OJK harus otomatis dan auditable
- **School context**: guru/ustadz dan staf memiliki gaji tetap (lower risk), siswa/santri tidak boleh pinjam langsung (orang tua sebagai penjamin)

Stakeholders: guru/ustadz (salary-based, higher limits), kepala sekolah (oversight), TU/staf (salary-based), siswa/santri (no direct loans, parent as guarantor), orang tua/wali (guarantors, education loans), external, asrama staff, canteen operators.

## Decision

### 1. Application Flow — Multi-Level Approval

Pengajuan pinjaman **wajib** melalui formulir aplikasi dengan approval bertingkat. Ini adalah flow yang lebih ketat dibanding pembukaan rekening tabungan/deposito (lihat [ADR-K002](./ADR-K002-rekening.md)).

```
Nasabah (via Teller) mengisi form pengajuan pinjaman
        |
        v
+-------------------+
|  Status: DRAFT    |  Teller bisa edit sebelum submit
+--------+----------+
         | submit (validasi eligibility, credit check)
         v
+-------------------+
| Status: PENDING   |  Menunggu analisis kredit
+--------+----------+
         |
         v
+----------------------------+
| STEP 1: Credit Analysis    |  Supervisor melakukan analisis:
| - Verifikasi penghasilan   |  - DTI ratio check
| - Cek riwayat pinjaman     |  - Kolektibilitas existing
| - Evaluasi jaminan         |  - Rekomendasi: approve/reject
+--------+-------------------+
         | recommendation
         v
+----------------------------+
| STEP 2: Approval           |  Manager/Admin approve/reject
| - Review analisis kredit   |  berdasarkan approval limit
| - Final decision           |
+--------+-------------------+
         |
    +----+----+
    v         v
+--------+ +----------+
|APPROVED| | REJECTED |  Wajib alasan penolakan
+---+----+ +----------+
    | auto-create rekening pinjaman (via K002 flow)
    v
+-------------------+
| DISBURSEMENT      |  Pencairan ke rekening nasabah
+-------------------+
```

**Approval limits configurable per role per tenant:**

```
approval_limit_config:
  Supervisor:
    max_amount: 5000000         # Bisa approve s/d Rp 5 juta (credit analysis only)
    can_final_approve: false    # Tidak bisa final approve, hanya recommend
  Manager:
    max_amount: 25000000        # Bisa approve s/d Rp 25 juta
    can_final_approve: true
  Admin:
    max_amount: null            # Unlimited
    can_final_approve: true
```

**Aturan:**
- Teller bisa buat pengajuan tapi **tidak bisa approve sendiri** — separation of duties (konsisten dengan K001, K002)
- Supervisor melakukan **credit analysis** dan memberikan rekomendasi — bukan final approval
- Manager approve untuk pinjaman di bawah limit-nya, di atas limit harus Admin
- Approval limits **configurable per tenant** — angka di atas adalah default
- Pengajuan di atas limit Manager **otomatis di-escalate** ke Admin
- Rejection wajib menyertakan `rejection_reason`
- Riwayat setiap step (analysis, approval) tersimpan di audit log

### 2. Loan Types

#### 2a. General Mode — Jenis Pinjaman

| Jenis | Deskripsi | Tenor | Approval | Catatan |
|-------|-----------|-------|----------|---------|
| Pinjaman Reguler | Kebutuhan umum anggota | 3-36 bulan | Standard flow | Produk utama |
| Pinjaman Darurat | Kebutuhan mendesak (sakit, bencana) | 1-12 bulan | Fast-track (Manager langsung) | Plafon lebih kecil, bunga lebih rendah |
| Pinjaman Pendidikan | Biaya sekolah anak/keluarga | 6-48 bulan | Standard flow | Tenor lebih panjang, bunga kompetitif |

#### 2b. Islamic Mode — Jenis Pembiayaan

| Akad | Prinsip | Penggunaan | Mekanisme |
|------|---------|------------|-----------|
| **Murabahah** | Jual beli + margin | Pembelian barang (paling umum) | Koperasi beli barang, jual ke nasabah dengan margin tetap |
| **Musyarakah** | Kemitraan | Modal usaha | Modal bersama, bagi hasil sesuai nisbah |
| **Mudharabah** | Bagi hasil | Modal usaha (koperasi 100% modal) | Koperasi = shahibul maal, nasabah = mudharib |
| **Ijarah** | Sewa | Sewa aset/jasa | Nasabah bayar ujrah (sewa) berkala |
| **Qardh** | Pinjaman kebajikan | Darurat/sosial | Tanpa margin/bunga, hanya biaya admin |

**Aturan per akad:**
- **Murabahah**: margin ditetapkan di awal dan **tidak berubah** selama tenor — angsuran flat. Harga pokok + margin = harga jual, dibagi rata ke installment
- **Musyarakah**: nisbah bagi hasil disepakati di awal, bagi hasil aktual berdasarkan profit riil per periode
- **Mudharabah**: koperasi menanggung seluruh modal, nasabah mengelola, bagi hasil sesuai nisbah
- **Ijarah**: ujrah (sewa) ditetapkan di awal per periode, bisa di-review berkala
- **Qardh**: nasabah hanya mengembalikan pokok + biaya admin (max configurable), tidak ada profit untuk koperasi

### 3. Credit Scoring / Eligibility

Sebelum pengajuan bisa di-submit, sistem melakukan **pre-check kelayakan** otomatis:

```
Credit Pre-check (auto, saat submit):
+-- 1. Nasabah ACTIVE                                     v
+-- 2. KYC level mencukupi (full untuk pinjaman besar)    v
+-- 3. Produk pinjaman eligible untuk relation type        v
+-- 4. Tidak melebihi max concurrent loans                 v
+-- 5. Plafon dalam range produk (min/max loan amount)     v
+-- 6. Tenor dalam range produk (min/max tenor)            v
+-- 7. Kolektibilitas existing loans semua Lancar (Kol-1)  v
    |
    ALL PASS -> Submit allowed
    ANY FAIL -> Submit blocked, return error list
```

**Credit analysis (manual, oleh Supervisor):**

```
Credit Analysis Checklist:
+-- 1. Verifikasi penghasilan
|   +-- Guru/Staf: data gaji dari sistem sekolah (jika terintegrasi)
|   +-- Lainnya: slip gaji, surat keterangan penghasilan
|
+-- 2. Max plafon calculation
|   +-- base_plafon = salary x multiplier (configurable, default: 3x)
|   +-- simpanan_bonus = total_simpanan x bonus_pct (configurable, default: 50%)
|   +-- max_plafon = base_plafon + simpanan_bonus
|   +-- requested <= max_plafon -> PASS
|
+-- 3. Debt-to-Income (DTI) ratio
|   +-- total_angsuran_existing + angsuran_baru
|   +-- DTI = total_angsuran / monthly_income
|   +-- DTI <= max_dti (configurable, default: 40%) -> PASS
|
+-- 4. Concurrent loans check
|   +-- active_loans_count < max_concurrent (configurable, default: 3)
|   +-- PASS
|
+-- 5. Evaluasi jaminan (jika diperlukan)
    +-- Nilai taksasi >= LTV requirement
    +-- Detail di ADR-K010
```

**Aturan khusus per stakeholder:**
- **Guru/Ustadz & Staf**: salary-based, higher plafon multiplier (configurable), eligible payroll deduction
- **Siswa/Santri**: **tidak boleh** mengajukan pinjaman langsung — orang tua/wali harus sebagai peminjam dengan siswa sebagai beneficiary
- **Orang tua/Wali**: eligible untuk pinjaman pendidikan, bisa menjadi penjamin untuk anggota lain
- **External**: plafon lebih kecil (configurable), mungkin wajib jaminan

### 4. Penjamin (Guarantor)

Penjamin diperlukan sebagai mitigasi risiko untuk pinjaman di atas threshold tertentu.

```
guarantor_config:
  required_above_amount: 10000000   # Wajib penjamin untuk pinjaman > Rp 10 juta
  max_guarantors: 3                 # Maksimal penjamin per pinjaman
  min_guarantors_above_threshold: 1 # Minimal 1 penjamin di atas threshold
  student_always_require: true      # Siswa: SELALU wajib penjamin (orang tua)
```

**Jenis penjamin:**

| Jenis | Sumber Data | Validasi |
|-------|-------------|----------|
| Nasabah existing | Link ke `nasabah.id` | Status ACTIVE, tidak sedang menjamin di atas limit |
| Orang tua/wali siswa | Link ke `nasabah.id` (parent) | Harus ber-relation_type = parent |
| Pihak eksternal | Data personal di `penjamin` tabel | KTP, alamat, kontak — tidak harus nasabah |

**Data model penjamin:**

```
pinjaman_penjamin
+-- id                    UUID v7 (PK)
+-- tenant_id             UUID (FK -> tenant)
+-- pinjaman_id           UUID (FK -> pinjaman) NOT NULL
|
+-- -- Identitas Penjamin --
+-- guarantor_type        ENUM (internal, external)
+-- nasabah_id            UUID (nullable, FK -> nasabah, jika internal)
+-- full_name             VARCHAR NOT NULL
+-- identity_number       VARCHAR NOT NULL
+-- phone                 VARCHAR
+-- address               TEXT
+-- relationship          VARCHAR (hubungan dengan peminjam)
|
+-- -- Jaminan Personal --
+-- guarantee_amount      NUMERIC(15,2) NOT NULL (nominal yang dijaminkan)
+-- guarantee_letter_url  VARCHAR (nullable, upload surat pernyataan penjaminan)
|
+-- -- Status --
+-- status                ENUM (active, released) DEFAULT 'active'
+-- released_at           TIMESTAMPTZ (nullable)
+-- released_by           UUID (nullable)
|
+-- -- Vernon Fields --
+-- _rels                 JSONB NOT NULL DEFAULT '{}'
+-- _data                 JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at        TIMESTAMPTZ
    +-- created_by        UUID
    +-- updated_at        TIMESTAMPTZ
    +-- updated_by        UUID
```

**Aturan:**
- Penjamin internal harus nasabah **ACTIVE** di tenant yang sama
- Satu nasabah bisa menjadi penjamin untuk **max N pinjaman** (configurable per tenant, default: 5)
- Penjamin di-release otomatis saat pinjaman COMPLETED
- Penjamin tidak bisa di-deactivate (K001) selama masih menjamin pinjaman aktif — blocking check
- Surat pernyataan penjaminan wajib di-upload untuk pinjaman di atas threshold tertentu

### 5. Disbursement (Pencairan)

Pencairan dilakukan setelah pinjaman di-approve dan rekening pinjaman dibuat.

```
Loan Approved
    |
    v
+-------------------------------------+
| 1. Create rekening pinjaman          |  Via K002 approval flow
|    (balance = plafon, outstanding)   |
| 2. Determine disbursement method     |
+--------+----------------------------+
         |
    +----+----------------+
    v                     v
+------------+   +------------------------+
| GENERAL    |   | ISLAMIC                |
| Transfer   |   | Tergantung akad:       |
| ke rek.    |   | +-- Murabahah:         |
| tabungan   |   | |   bayar ke supplier  |
| nasabah    |   | |   (wakalah)          |
|            |   | +-- Musyarakah:        |
|            |   | |   transfer modal     |
|            |   | +-- Ijarah:            |
|            |   | |   penyerahan aset    |
|            |   | +-- Qardh:             |
|            |   |     transfer ke rek.   |
+------------+   +------------------------+
```

**Aturan pencairan:**
- **General**: dana ditransfer ke rekening tabungan nasabah — dicatat sebagai transaksi debit di rekening pinjaman dan kredit di rekening tabungan
- **Murabahah**: koperasi membeli barang dari supplier, bisa dengan **wakalah** (nasabah membeli atas nama koperasi) — bukti pembelian wajib di-upload
- **Musyarakah/Mudharabah**: modal ditransfer ke rekening nasabah atau langsung ke usaha
- **Ijarah**: penyerahan aset/jasa, dicatat sebagai pencairan
- **Qardh**: transfer ke rekening tabungan nasabah (sama seperti general)
- Pencairan hanya bisa dilakukan **sekali** per pinjaman — tidak ada partial disbursement (simplicity untuk koperasi sekolah)
- Pencairan dicatat di audit log dengan bukti transfer/penyerahan

### 6. Loan Rekening

Saat pinjaman di-approve, sistem membuat **rekening pinjaman** melalui flow K002:

```
Rekening Pinjaman:
+-- category = PINJAMAN
+-- balance = outstanding principal (decreasing)
+-- Initial balance = loan_amount (plafon yang disetujui)
|
|   Setiap pembayaran angsuran:
|   balance -= principal_portion
|
|   Saat lunas:
|   balance = 0, status = CLOSED
```

**Aturan:**
- Balance rekening pinjaman = **sisa pokok outstanding** — berkurang setiap pembayaran angsuran (porsi pokok)
- Initial balance di-set saat pencairan = plafon yang disetujui
- Rekening pinjaman di-CLOSED otomatis saat balance = 0 (lunas)
- Satu pengajuan pinjaman = satu rekening pinjaman — relasi 1:1

### 7. Grace Period (Masa Tenggang)

Masa tenggang sebelum angsuran pertama jatuh tempo:

```
grace_period_config (per produk):
  grace_period_months: 1        # Bulan sebelum angsuran pertama
  grace_period_type:
    +-- NONE                    # Langsung angsuran bulan pertama
    +-- PRINCIPAL_ONLY          # Selama grace, bayar bunga/margin saja
    +-- FULL                    # Selama grace, tidak bayar apapun
```

**Contoh:**
```
Pencairan: 15 Januari 2026
Grace period: 1 bulan, type: FULL
Angsuran pertama jatuh tempo: 15 Maret 2026 (skip Februari)
```

**Aturan:**
- Grace period **configurable per produk** di K003
- Selama grace period type FULL, **tidak ada angsuran** dan **tidak ada denda**
- Selama grace period type PRINCIPAL_ONLY, nasabah bayar bunga/margin saja (pokok belum diangsur)
- Grace period diperhitungkan saat generate jadwal angsuran (K008)
- Pinjaman darurat (emergency) biasanya tanpa grace period

### 8. Restructuring / Rescheduling

Untuk pinjaman bermasalah, koperasi dapat melakukan restrukturisasi:

```
Restructuring Request
        |
        v
+----------------------------+
| Supervisor membuat         |
| proposal restrukturisasi:  |
| - Perpanjangan tenor       |
| - Penurunan angsuran       |
| - Penundaan angsuran       |
| - Konversi akad (BMT)      |
+--------+-------------------+
         |
         v
+----------------------------+
| Manager+ approve           |  Minimal Manager, configurable
| proposal restrukturisasi   |
+--------+-------------------+
         |
         v
+----------------------------+
| Execute:                   |
| 1. Void jadwal lama        |
| 2. Generate jadwal baru    |
| 3. Update loan terms       |
| 4. Record restructuring    |
| 5. Reset NPL counter      |
+----------------------------+
```

**Jenis restrukturisasi:**

| Jenis | Deskripsi | Impact |
|-------|-----------|--------|
| Rescheduling | Perpanjangan tenor, angsuran lebih kecil | Jadwal baru, total bunga/margin bisa berubah |
| Reconditioning | Perubahan syarat (rate, grace period) | Terms berubah, jadwal baru |
| Restructuring | Kombinasi rescheduling + reconditioning + tambahan dana | Paling kompleks, pinjaman baru effective |

**Aturan:**
- Hanya untuk pinjaman dengan kolektibilitas **Kol-2 ke atas** (bermasalah)
- Approval minimal **Manager**, configurable per tenant
- Jadwal lama di-void (status = VOIDED), jadwal baru di-generate
- Riwayat restrukturisasi tersimpan — pinjaman yang pernah di-restructure di-flag
- Maksimal **2x restrukturisasi** per pinjaman (configurable) — lebih dari itu harus write-off
- BMT: restrukturisasi bisa termasuk **konversi akad** (misal: Murabahah ke Qardh) dengan akad baru

### 9. Early Settlement (Pelunasan Dipercepat)

Nasabah bisa melunasi seluruh sisa pinjaman sebelum tenor berakhir:

```
Early Settlement Request
        |
        v
+--------------------------------+
| Hitung total pelunasan:        |
| +-- Sisa pokok outstanding     |
| +-- Bunga/margin berjalan      |
| +-- Denda outstanding (jika ada) |
| +-- [General] +/- early settlement adjustment |
| +-- [Islamic] - Ibra' (diskon margin) |
|     = Total amount to pay      |
+--------+-----------------------+
         | Nasabah bayar
         v
+--------------------------------+
| Execute:                       |
| 1. Record final payment        |
| 2. Close remaining schedule    |
| 3. Set loan status COMPLETED   |
| 4. Close rekening pinjaman     |
| 5. Release penjamin & jaminan  |
+--------------------------------+
```

**Aturan per mode:**

| Aspek | General | Islamic |
|-------|---------|---------|
| Penalti pelunasan dini | Configurable: charge % atau waive | **Tidak ada penalti** — fatwa DSN-MUI |
| Diskon bunga/margin | Opsional, kebijakan koperasi | **Ibra'** (diskon margin belum jatuh tempo) — wajib diberikan |
| Sisa bunga/margin | Bunga berjalan sampai tanggal pelunasan | Margin yang belum jatuh tempo di-diskon (ibra') |

**Ibra' (Islamic mode):**
- Margin Murabahah yang **belum jatuh tempo** wajib didiskon
- Besaran ibra' = sisa margin yang belum earned x discount rate (configurable, default: 100%)
- Ibra' dicatat sebagai pengurang pendapatan margin di jurnal
- Total pelunasan = sisa pokok + margin berjalan - ibra'

### 10. Loan Status Lifecycle

```
+-------------------+
| Status: ACTIVE    |  Pinjaman dicairkan, angsuran berjalan
+--------+----------+
         |
    +----+------------+----------------+
    v                 v                v
+------------+ +--------------+ +------------+
| COMPLETED  | | RESTRUCTURED | | WRITTEN_OFF|
| Lunas      | | Di-restruktur| | Dihapusbuku|
+------------+ +--------------+ +------------+
```

**Status definitions:**

| Status | Deskripsi | Trigger |
|--------|-----------|---------|
| ACTIVE | Pinjaman berjalan, angsuran belum lunas | Saat pencairan |
| COMPLETED | Pinjaman lunas (semua angsuran terbayar atau early settlement) | Balance = 0 |
| RESTRUCTURED | Pinjaman lama diganti jadwal baru | Approval restrukturisasi |
| WRITTEN_OFF | Pinjaman dihapusbukukan (macet, tidak tertagih) | Approval write-off |

**Aturan:**
- `ACTIVE -> COMPLETED`: otomatis saat balance rekening = 0
- `ACTIVE -> RESTRUCTURED`: saat restrukturisasi di-approve — pinjaman ini di-flag, jadwal baru dimulai
- `ACTIVE -> WRITTEN_OFF`: saat write-off di-approve (Section 12)
- Status terminal (COMPLETED, WRITTEN_OFF) **tidak bisa di-revert**
- RESTRUCTURED bukan terminal — pinjaman tetap berjalan dengan jadwal baru, status kembali ke ACTIVE setelah restrukturisasi

### 11. NPL Classification (Kolektibilitas)

Klasifikasi kualitas pinjaman mengikuti standar OJK untuk LKM (Lembaga Keuangan Mikro):

| Kolektibilitas | Nama | DPD (Days Past Due) | Tindakan |
|----------------|------|---------------------|----------|
| Kol-1 | Lancar (Current) | 0 hari | Normal, tidak ada tindakan |
| Kol-2 | Dalam Perhatian Khusus (Special Mention) | 1-90 hari | Surat peringatan 1 |
| Kol-3 | Kurang Lancar (Substandard) | 91-120 hari | Surat peringatan 2, mulai intensif collection |
| Kol-4 | Diragukan (Doubtful) | 121-180 hari | Surat peringatan 3, evaluasi restrukturisasi |
| Kol-5 | Macet (Loss) | > 180 hari | Kandidat write-off, PPAP 100% |

**Auto-classification:**
```
Scheduled Job (daily):
+-- Hitung DPD setiap pinjaman ACTIVE
|   DPD = TODAY - oldest_unpaid_installment.due_date
|
+-- Map DPD -> Kolektibilitas
|   +-- DPD = 0        -> Kol-1
|   +-- DPD 1-90       -> Kol-2
|   +-- DPD 91-120     -> Kol-3
|   +-- DPD 121-180    -> Kol-4
|   +-- DPD > 180      -> Kol-5
|
+-- Update loan.collectibility
|
+-- Trigger notifikasi jika ada perubahan kolektibilitas
    +-- Kol-1 -> Kol-2: notif Supervisor
    +-- Kol-2 -> Kol-3: notif Manager
    +-- Kol-3 -> Kol-4: notif Admin
    +-- Kol-4 -> Kol-5: notif Admin + alert write-off candidate
```

**Aturan:**
- Klasifikasi otomatis berdasarkan **DPD** — Supervisor+ bisa override manual (documented reason)
- DPD range **configurable per tenant** — angka di atas adalah default per OJK untuk LKM
- Perubahan kolektibilitas dicatat di audit log
- Downgrade kolektibilitas (memburuk) otomatis, upgrade (membaik) bisa otomatis setelah pembayaran tepat waktu selama N bulan (configurable, default: 3)
- PPAP (Penyisihan Penghapusan Aktiva Produktif) dihitung berdasarkan kolektibilitas — detail di K015

### 12. Write-Off (Hapus Buku)

Pinjaman macet (Kol-5) yang sudah tidak tertagih bisa dihapusbukukan:

```
Write-off Proposal
        |
        v
+--------------------------------+
| Manager membuat proposal:      |
| - Alasan write-off             |
| - Total outstanding (pokok +   |
|   bunga/margin + denda)        |
| - Riwayat penagihan            |
| - PPAP coverage (harus 100%)  |
+--------+-----------------------+
         |
         v
+--------------------------------+
| Admin approve write-off        |  Hanya Admin
+--------+-----------------------+
         |
         v
+--------------------------------+
| Execute:                       |
| 1. Set loan status WRITTEN_OFF |
| 2. Close rekening pinjaman     |
| 3. Move to off-balance sheet   |
| 4. Release jaminan (jika ada)  |
| 5. Release penjamin            |
| 6. Jurnal: debit PPAP,        |
|    kredit pinjaman             |
+--------------------------------+
```

**Aturan:**
- Hanya pinjaman **Kol-5 (Macet)** yang bisa di-write-off
- PPAP harus sudah **100% tercadangkan** sebelum write-off diizinkan
- Write-off memerlukan approval **Admin** — operasi berisiko tinggi
- Data pinjaman **tidak dihapus** — dipindahkan ke tracking off-balance sheet
- Koperasi masih bisa melakukan penagihan setelah write-off (recovery) — jika ada recovery, dicatat sebagai pendapatan lain-lain
- Riwayat write-off tersimpan permanen di audit log

### 13. Data Model

```
pinjaman
+-- id                      UUID v7 (PK)
+-- tenant_id               UUID (FK -> tenant)
+-- branch_id               UUID (FK -> branch)
+-- nasabah_id              UUID (FK -> nasabah) NOT NULL
+-- rekening_id             UUID (FK -> rekening) NOT NULL (1:1 with rekening pinjaman)
+-- product_id              UUID (FK -> produk) NOT NULL
+-- application_id          UUID (FK -> pinjaman_application) NOT NULL
|
+-- -- Identitas --
+-- loan_number             VARCHAR UNIQUE per tenant (auto-generate)
|
+-- -- Terms --
+-- loan_amount             NUMERIC(15,2) NOT NULL (plafon yang disetujui)
+-- disbursed_amount        NUMERIC(15,2) NOT NULL (nominal yang dicairkan)
+-- outstanding_principal   NUMERIC(15,2) NOT NULL (sisa pokok, = rekening.balance)
+-- tenor_months            INTEGER NOT NULL
+-- installment_amount      NUMERIC(15,2) NOT NULL (angsuran per bulan)
|
+-- -- Rate/Yield (General) --
+-- interest_rate           NUMERIC(7,4) (nullable, % p.a.)
+-- calculation_method      ENUM (flat, declining, annuity) nullable
|
+-- -- Akad/Yield (Islamic) --
+-- akad_type               ENUM (murabahah, musyarakah, mudharabah, ijarah, qardh) nullable
+-- margin_rate             NUMERIC(7,4) (nullable, % margin Murabahah)
+-- margin_amount           NUMERIC(15,2) (nullable, total margin nominal)
+-- nisbah_nasabah          NUMERIC(5,2) (nullable, % bagi hasil nasabah)
+-- nisbah_koperasi         NUMERIC(5,2) (nullable, % bagi hasil koperasi)
+-- ujrah_amount            NUMERIC(15,2) (nullable, nominal ujrah Ijarah)
|
+-- -- Purpose --
+-- loan_purpose            TEXT NOT NULL (tujuan pinjaman)
+-- loan_purpose_category   ENUM (konsumtif, produktif, pendidikan, darurat, lainnya)
|
+-- -- Schedule --
+-- grace_period_months     INTEGER DEFAULT 0
+-- grace_period_type       ENUM (none, principal_only, full) DEFAULT 'none'
+-- first_installment_date  DATE NOT NULL
+-- last_installment_date   DATE NOT NULL
+-- installment_day         INTEGER NOT NULL (tanggal jatuh tempo: 1-28)
|
+-- -- Disbursement --
+-- disbursement_date       DATE (nullable, diisi saat pencairan)
+-- disbursement_method     ENUM (transfer_tabungan, wakalah, direct_supplier)
+-- disbursement_evidence   VARCHAR (nullable, URL bukti pencairan)
+-- disbursement_rekening_id UUID (nullable, FK -> rekening tujuan pencairan)
|
+-- -- NPL --
+-- collectibility          INTEGER NOT NULL DEFAULT 1 (Kol 1-5)
+-- dpd                     INTEGER NOT NULL DEFAULT 0 (Days Past Due)
+-- collectibility_updated_at TIMESTAMPTZ
|
+-- -- Status --
+-- status                  ENUM (active, completed, restructured, written_off)
+-- completed_at            TIMESTAMPTZ (nullable)
+-- written_off_at          TIMESTAMPTZ (nullable)
+-- written_off_by          UUID (nullable)
+-- write_off_reason        TEXT (nullable)
|
+-- -- Restructuring --
+-- is_restructured         BOOLEAN DEFAULT false
+-- restructure_count       INTEGER DEFAULT 0
+-- restructured_from_id    UUID (nullable, FK -> pinjaman, jika ini hasil restrukturisasi)
|
+-- -- Flags --
+-- has_guarantor           BOOLEAN DEFAULT false
+-- has_collateral          BOOLEAN DEFAULT false
|
+-- -- Vernon Fields --
+-- _rels                   JSONB NOT NULL DEFAULT '{}'
+-- _data                   JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at          TIMESTAMPTZ
    +-- created_by          UUID
    +-- updated_at          TIMESTAMPTZ
    +-- updated_by          UUID
```

```
pinjaman_application
+-- id                      UUID v7 (PK)
+-- tenant_id               UUID (FK -> tenant)
+-- branch_id               UUID (FK -> branch)
+-- nasabah_id              UUID (FK -> nasabah) NOT NULL
+-- product_id              UUID (FK -> produk) NOT NULL
|
+-- -- Pengajuan --
+-- requested_amount        NUMERIC(15,2) NOT NULL
+-- requested_tenor_months  INTEGER NOT NULL
+-- loan_purpose            TEXT NOT NULL
+-- loan_purpose_category   ENUM (konsumtif, produktif, pendidikan, darurat, lainnya)
+-- monthly_income          NUMERIC(15,2) (deklarasi penghasilan bulanan)
+-- form_data               JSONB (snapshot seluruh isian form)
|
+-- -- Credit Analysis --
+-- credit_score            NUMERIC(5,2) (nullable, hasil scoring)
+-- dti_ratio               NUMERIC(5,2) (nullable, debt-to-income ratio)
+-- max_eligible_amount     NUMERIC(15,2) (nullable, plafon max sesuai scoring)
+-- analysis_notes          TEXT (nullable, catatan Supervisor)
+-- analyzed_at             TIMESTAMPTZ (nullable)
+-- analyzed_by             UUID (nullable, FK -> user)
|
+-- -- Approval --
+-- approved_amount         NUMERIC(15,2) (nullable, bisa berbeda dari requested)
+-- approved_tenor_months   INTEGER (nullable)
+-- approved_rate           NUMERIC(7,4) (nullable)
|
+-- -- Template --
+-- template_version        INTEGER
+-- terms_accepted          BOOLEAN NOT NULL
+-- terms_accepted_at       TIMESTAMPTZ
|
+-- -- Workflow --
+-- status                  ENUM (draft, pending, analyzing, approved, rejected)
+-- submitted_at            TIMESTAMPTZ (nullable)
+-- reviewed_at             TIMESTAMPTZ (nullable)
+-- reviewed_by             UUID (nullable, FK -> user)
+-- rejection_reason        TEXT (nullable)
|
+-- -- Result --
+-- pinjaman_id             UUID (nullable, FK -> pinjaman, filled on approval)
|
+-- -- Vernon Fields --
+-- _rels                   JSONB NOT NULL DEFAULT '{}'
+-- _data                   JSONB NOT NULL DEFAULT '{}'
|
+-- -- Audit --
    +-- created_at          TIMESTAMPTZ
    +-- created_by          UUID
    +-- updated_at          TIMESTAMPTZ
    +-- updated_by          UUID
```

### 14. Vernon _rels dan _data Structure

Pinjaman menggunakan Vernon pattern karena listing pinjaman membutuhkan data nasabah, produk, branch, dan penjamin (4+ JOIN).

**_rels (pinjaman):**
```json
{
  "tenant_id":      "018f...",
  "branch_id":      "018f...",
  "nasabah_id":     "018f...",
  "rekening_id":    "018f...",
  "product_id":     "018f...",
  "application_id": "018f..."
}
```

**_data (pinjaman):**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "product": {
    "id":   "018f...",
    "name": "Pinjaman Reguler",
    "code": "PJ-REG"
  },
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "PJ-2026-JKT-00000001"
  }
}
```

**Catatan keamanan:**
- **Tidak ada data finansial** di `_data` — no saldo, no rate, no penghasilan
- `_data` hanya berisi informasi yang dibutuhkan untuk listing/display
- Data sensitif harus diambil dari kolom asli via query detail

**SyncEngine triggers:**
- `NasabahUpdatedEvent` -> update `_data.nasabah` di semua pinjaman nasabah tersebut
- `ProductUpdatedEvent` -> update `_data.product` di semua pinjaman dengan produk tersebut
- `BranchUpdatedEvent` -> update `_data.branch` di semua pinjaman cabang tersebut

### 15. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Buat pengajuan pinjaman | v | v | v | v |
| View daftar pengajuan | v | v | v | v |
| Credit analysis (recommend) | - | v | v | v |
| Approve/Reject pinjaman (sesuai limit) | - | - | v | v |
| Approve di atas limit Manager | - | - | - | v |
| Proses pencairan | v | v | v | v |
| View data pinjaman | v | v | v | v |
| View NPL report | - | v | v | v |
| Override kolektibilitas | - | - | v | v |
| Ajukan restrukturisasi | - | v | v | v |
| Approve restrukturisasi | - | - | v | v |
| Ajukan early settlement | v | v | v | v |
| Approve early settlement | - | v | v | v |
| Ajukan write-off | - | - | v | v |
| Approve write-off | - | - | - | v |
| Edit approval limits | - | - | - | v |
| Add/remove penjamin | - | v | v | v |

**Catatan:**
- Teller bisa buat pengajuan tapi **tidak bisa approve** — separation of duties
- Credit analysis adalah tugas **Supervisor** — menghasilkan rekomendasi, bukan keputusan final
- Approval bertingkat berdasarkan **nominal** — Manager sampai limit-nya, di atas itu Admin
- Write-off hanya **Admin** — operasi paling berisiko tinggi
- Override kolektibilitas minimal **Manager** — harus dengan alasan terdokumentasi

### 16. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Pinjaman | Pembiayaan |
| Application title | Pengajuan Pinjaman | Pengajuan Pembiayaan |
| Rate label | Suku Bunga | Margin / Nisbah Bagi Hasil |
| Rate display | Bunga: 12% p.a. flat | Margin: 15% (Murabahah) / Nisbah: 40:60 |
| Payment label | Angsuran Pinjaman | Angsuran Pembiayaan |
| Interest income | Pendapatan Bunga | Pendapatan Margin / Bagi Hasil |
| Penalty | Denda Keterlambatan | Ta'zir |
| Early settlement discount | Diskon Bunga | Ibra' |
| Loan purpose doc | Surat Pengajuan Pinjaman | Surat Pengajuan Pembiayaan + Akad |
| Guarantor | Penjamin | Penjamin / Kafiil |
| Contract | Perjanjian Kredit | Akad Pembiayaan |
| Restructuring | Restrukturisasi Pinjaman | Restrukturisasi Pembiayaan |
| Write-off | Hapus Buku | Hapus Buku |
| NPL | Pinjaman Bermasalah | Pembiayaan Bermasalah |

## Consequences

### Positif

- **Auditable** — setiap pinjaman punya jejak lengkap dari pengajuan, analisis kredit, approval, pencairan, hingga pelunasan
- **Risk-managed** — credit scoring, DTI ratio, kolektibilitas otomatis, dan PPAP mencegah kerugian koperasi
- **Compliant** — NPL classification per OJK, PPAP, dan approval bertingkat memenuhi regulasi LKM
- **Fraud prevention** — separation of duties (Teller input, Supervisor analyze, Manager approve), approval limits
- **Flexible** — dual-mode mendukung konvensional dan BMT, approval limits configurable per tenant
- **School-aware** — guru/staf mendapat plafon berbasis gaji, siswa harus punya penjamin orang tua
- **Syariah compliant** — akad, ibra', ta'zir ke dana sosial memenuhi fatwa DSN-MUI

### Negatif

- **Complexity tinggi** — domain paling kompleks dengan banyak state, flow, dan aturan
- **Multi-step approval lambat** — pencairan membutuhkan analisis + approval + pencairan (3 langkah)
- **Credit scoring sederhana** — formula basic (salary x multiplier + simpanan bonus), bukan ML scoring
- **Single disbursement** — tidak mendukung partial disbursement (revolving credit)
- **Restructuring limit** — max 2x restrukturisasi bisa terlalu kaku untuk beberapa kasus

### Mitigasi

- Fast-track flow untuk pinjaman darurat — Manager bisa approve langsung tanpa Supervisor analysis
- Credit scoring formula configurable per tenant — bisa disesuaikan dengan pengalaman koperasi
- Single disbursement mencukupi untuk koperasi sekolah (bukan bank komersial)
- Restructuring limit configurable — Admin bisa override untuk kasus khusus
- Batch approval untuk Manager yang harus approve banyak pengajuan sekaligus

## Alternatives Considered

### A. Single-Level Approval (Manager langsung)

Semua pengajuan langsung ke Manager tanpa credit analysis Supervisor.

**Ditolak** karena: Manager akan overwhelmed dengan volume pengajuan. Credit analysis oleh Supervisor memfilter pengajuan yang tidak layak sebelum sampai ke Manager. Separation of duties juga penting untuk fraud prevention.

### B. Automated Approval (tanpa manusia)

Sistem otomatis approve berdasarkan credit scoring, tanpa human review.

**Ditolak** karena: risiko terlalu tinggi untuk koperasi sekolah. Banyak faktor kualitatif (karakter nasabah, kondisi keluarga) yang tidak bisa ditangkap oleh scoring formula sederhana. Human judgment tetap diperlukan.

### C. Revolving Credit Line (plafon berjalan)

Nasabah mendapat credit line yang bisa ditarik berulang kali.

**Ditolak** karena: terlalu kompleks untuk koperasi sekolah. Model pinjaman term loan (fixed amount, fixed tenor) lebih sederhana dan lebih mudah dikelola. Revolving credit bisa dipertimbangkan di fase future.

### D. External Credit Bureau Integration

Integrasi dengan SLIK/SID OJK untuk checking BI credit history.

**Ditolak untuk MVP** karena: integrasi teknis kompleks dan butuh izin khusus. Internal credit scoring (DTI, riwayat pinjaman di koperasi, saldo simpanan) sudah cukup untuk koperasi sekolah. Bisa ditambahkan sebagai enhancement di masa depan.

### E. Kolektibilitas Manual Only (tanpa auto-classification)

Supervisor menentukan kolektibilitas secara manual.

**Ditolak** karena: inkonsistensi dan human error. DPD-based auto-classification memastikan standar yang sama untuk semua pinjaman. Manual override tetap tersedia untuk kasus khusus.

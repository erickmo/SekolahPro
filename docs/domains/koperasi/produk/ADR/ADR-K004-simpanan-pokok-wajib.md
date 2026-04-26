# ADR-K004: Simpanan Pokok & Wajib

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Simpanan Pokok dan Simpanan Wajib adalah **dua pilar keanggotaan** koperasi. Sesuai UU Perkoperasian dan AD/ART koperasi, setiap anggota wajib membayar keduanya sebagai syarat keanggotaan aktif. Dana ini menjadi **modal dasar koperasi** dan menentukan hak suara serta pembagian SHU.

Tantangan utama di lingkungan sekolah:
- **Siswa/santri**: nominal kecil, collection bisa mingguan oleh wali kelas/guru
- **Guru/staff**: bisa potong gaji (payroll deduction) atau auto-debit dari tabungan
- **Orang tua/wali**: pembayaran via teller atau transfer, mungkin membayar untuk anaknya
- **Arrears tracking**: siapa yang belum bayar wajib bulan ini, reminder otomatis
- **Refund saat keluar**: kedua simpanan dikembalikan saat nasabah resign/lulus/keluar

Kedua simpanan ini **tidak bisa ditarik** selama nasabah masih aktif — ini yang membedakannya dari tabungan sukarela (K005).

## Decision

### 1. Simpanan Pokok — One-Time Membership Deposit

Simpanan Pokok adalah setoran **satu kali** saat nasabah pertama kali bergabung sebagai anggota koperasi.

```
NasabahApprovedEvent
        │
        v
┌───────────────────────────────────────┐
│ Auto-create rekening SIMPANAN_POKOK   │
│ (K002 Section 2)                      │
│ - produk: default simpanan_pokok      │
│ - saldo: 0 (belum disetor)           │
│ - status: ACTIVE                      │
└───────────────────────────────────────┘
        │
        v
Teller mencatat setoran simpanan pokok
        │
        v
┌───────────────────────────────────────┐
│ Transaksi: SETORAN_POKOK              │
│ - nominal: sesuai tenant config       │
│ - saldo setelah: = nominal pokok      │
│ - sumber: tunai / transfer / potong   │
│   dari tabungan                       │
└───────────────────────────────────────┘
```

**Aturan:**
- Nominal **tetap per tenant** — semua nasabah bayar jumlah yang sama (sesuai AD/ART)
- Nominal disimpan di `tenant_product_config.simpanan_pokok_amount` (K003 Section 10)
- **Satu kali bayar** — tidak ada setoran kedua (kecuali ada perubahan AD/ART yang menaikkan nominal)
- **Tidak bisa ditarik** selama nasabah ACTIVE — penarikan hanya saat keluar (Section 7)
- Rekening otomatis dibuat saat nasabah di-approve, tapi **setoran tetap transaksi terpisah**
- Nasabah dianggap **belum lengkap** jika simpanan pokok belum disetor — ditandai flag di dashboard

**Kenaikan nominal (kasus khusus):**
```
AD/ART berubah: nominal Rp 100.000 → Rp 150.000
├── Nasabah existing: bayar selisih Rp 50.000 (one-time top-up)
├── Nasabah baru: langsung bayar Rp 150.000
└── Implementasi: Manager buat transaksi batch top-up
    dengan approval Admin
```

### 2. Simpanan Wajib — Monthly Mandatory Deposit

Simpanan Wajib adalah setoran **bulanan** yang wajib dibayar setiap anggota selama masa keanggotaan.

```
Setiap bulan (tanggal configurable)
        │
        v
┌───────────────────────────────────────┐
│ Generate tagihan SIMPANAN_WAJIB       │
│ untuk semua nasabah ACTIVE            │
│ - period: YYYY-MM                     │
│ - amount: sesuai tenant config        │
│ - due_date: configurable (1-28)       │
│ - status: UNPAID                      │
└───────────────────────────────────────┘
        │
        v
Collection via berbagai channel (Section 3)
        │
        v
┌───────────────────────────────────────┐
│ Transaksi: SETORAN_WAJIB              │
│ - nominal: sesuai tagihan             │
│ - referensi: wajib_billing.id         │
│ - status tagihan: PAID                │
└───────────────────────────────────────┘
```

**Aturan:**
- Nominal **tetap per bulan** per tenant — configurable di `tenant_product_config.simpanan_wajib_amount`
- Bisa berbeda per `school_relation_type` jika tenant mengonfigurasi:
  ```
  simpanan_wajib_amount_by_type:
    student:  10000     # Rp 10.000/bulan
    teacher:  50000     # Rp 50.000/bulan
    staff:    25000     # Rp 25.000/bulan
    parent:   25000     # Rp 25.000/bulan
    external: 50000     # Rp 50.000/bulan
  ```
- Jika `simpanan_wajib_amount_by_type` tidak dikonfigurasi, semua type menggunakan `simpanan_wajib_amount` yang sama
- **Tidak bisa ditarik** selama nasabah ACTIVE — penarikan hanya saat keluar (Section 7)
- Tagihan di-generate otomatis oleh scheduled job di awal bulan
- Nasabah yang baru join **mid-month** — tagihan pertama prorated atau full bulan berikutnya (configurable)

### 3. Collection Mechanisms

Pembayaran simpanan pokok & wajib bisa melalui beberapa channel:

```
┌─────────────────────────────────────────────────────────────┐
│                   COLLECTION CHANNELS                        │
├──────────────────┬──────────────────────────────────────────┤
│ Channel          │ Keterangan                               │
├──────────────────┼──────────────────────────────────────────┤
│ 1. Teller Cash   │ Nasabah datang ke counter,               │
│                  │ bayar tunai. Teller catat.               │
├──────────────────┼──────────────────────────────────────────┤
│ 2. Auto-debit    │ Potong otomatis dari tabungan            │
│    Tabungan      │ nasabah. Lihat Section 10.               │
├──────────────────┼──────────────────────────────────────────┤
│ 3. Payroll       │ Potong gaji guru/staff via               │
│    Deduction     │ integrasi HR. Batch processing.          │
├──────────────────┼──────────────────────────────────────────┤
│ 4. Wali Kelas    │ Guru mengumpulkan dari siswa,            │
│    Collection    │ setor batch ke koperasi.                 │
├──────────────────┼──────────────────────────────────────────┤
│ 5. Transfer Bank │ Transfer ke rekening koperasi,           │
│                  │ Teller cocokkan manual.                  │
├──────────────────┼──────────────────────────────────────────┤
│ 6. Parent Pay    │ Orang tua bayar untuk anak               │
│                  │ (referensi nasabah_id anak).             │
└──────────────────┴──────────────────────────────────────────┘
```

**Channel detail:**

**3a. Teller Cash:**
- Nasabah atau perwakilan datang ke counter
- Teller memilih nasabah, pilih tagihan yang mau dibayar
- Bisa bayar **multiple bulan** sekaligus (bayar tunggakan + bulan ini)
- Receipt dicetak/dikirim digital

**3b. Auto-debit dari Tabungan:**
- Nasabah bisa opt-in auto-debit untuk simpanan wajib bulanan
- Konfigurasi: `auto_debit_source_rekening_id` (FK → rekening tabungan)
- Execution: scheduled job di tanggal due_date
- Detail aturan di Section 10

**3c. Payroll Deduction:**
- Khusus guru/staff yang gajinya diproses via sistem sekolah
- Koperasi mengirim **file tagihan batch** ke HR, HR memotong gaji
- HR mengirim **file settlement** kembali ke koperasi
- Teller/system memproses settlement batch — mark tagihan sebagai PAID
- Format file: CSV/Excel configurable per tenant

**3d. Wali Kelas Collection:**
- Wali kelas mengumpulkan uang dari siswa di kelas
- Wali kelas menyetor ke koperasi sebagai **batch deposit**
- Batch mencakup: daftar siswa + nominal per siswa
- Teller memvalidasi dan memproses batch — menghasilkan transaksi per siswa
- Fitur: cetak daftar tagihan per kelas untuk wali kelas

**3e. Parent Pay:**
- Orang tua membayar simpanan wajib anak
- Transaksi tercatat di rekening **anak** (bukan orang tua)
- Referensi: `paid_by_nasabah_id` (FK → nasabah orang tua)
- Khusus untuk nasabah dengan `school_relation_type = student`

### 4. Payment Tracking — Billing & Arrears

Setiap tagihan simpanan wajib dilacak per bulan:

```
simpanan_wajib_billing
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
├── rekening_id           UUID (FK → rekening simpanan_wajib)
├── period                VARCHAR (format: "2026-04", unique per nasabah)
├── amount                NUMERIC(15,2) NOT NULL
├── due_date              DATE NOT NULL
│
├── ── Payment ──
├── status                ENUM (unpaid, paid, partial, waived)
├── paid_amount            NUMERIC(15,2) DEFAULT 0
├── paid_at               TIMESTAMPTZ (nullable)
├── paid_via              ENUM (teller, auto_debit, payroll, batch, transfer) nullable
├── transaction_id        UUID (nullable, FK → transaksi, filled on payment)
├── paid_by_nasabah_id    UUID (nullable, FK → nasabah, jika dibayar orang lain)
│
├── ── Arrears ──
├── is_overdue            BOOLEAN DEFAULT false
├── overdue_since         DATE (nullable, = due_date + grace_period)
├── reminder_sent_at      TIMESTAMPTZ (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Status flow:**
```
┌──────────┐
│  UNPAID  │  Generate awal bulan
└────┬─────┘
     │
┌────┴────────────┐
│                 │
v                 v
┌──────┐    ┌─────────┐
│ PAID │    │ PARTIAL │  Bayar sebagian (jika diizinkan tenant)
└──────┘    └────┬────┘
                 │ bayar sisa
                 v
            ┌──────┐
            │ PAID │
            └──────┘

Khusus:
┌──────────┐
│  WAIVED  │  Dibebaskan oleh Manager+ (kasus khusus: sakit, bencana)
└──────────┘
```

**Aturan:**
- Tagihan di-generate otomatis oleh **scheduled job** setiap awal bulan (tanggal 1)
- Status `partial` hanya tersedia jika tenant mengaktifkan `allow_partial_wajib_payment`
- Status `waived` hanya oleh **Manager+** dengan alasan wajib — untuk kasus darurat
- Setiap pembayaran menghasilkan **transaksi** di tabel transaksi (K011) yang mereferensikan billing

### 5. Grace Period & Late Payment

```
grace_period_config:
  grace_days: 7                    # Configurable per tenant (default: 7 hari)
  reminder_schedule:
    - trigger: "due_date"          # Reminder pada tanggal jatuh tempo
    - trigger: "due_date + 3"      # Reminder 3 hari setelah jatuh tempo
    - trigger: "due_date + 7"      # Reminder saat grace period habis
  notification_channels:
    - in_app                       # Notifikasi di aplikasi
    - sms                          # SMS (jika dikonfigurasi)
```

**Flow:**
```
Tanggal 10 (due_date)
        │
        v
┌─────────────────┐
│ Reminder #1     │  "Simpanan wajib bulan April belum dibayar"
└────────┬────────┘
         │ +3 hari
         v
┌─────────────────┐
│ Reminder #2     │  "Simpanan wajib April sudah melewati jatuh tempo"
└────────┬────────┘
         │ +4 hari (= due_date + 7)
         v
┌─────────────────┐
│ Grace Expired   │  is_overdue = true, overdue_since = today
│ Reminder #3     │  "Simpanan wajib April TUNGGAKAN"
└─────────────────┘
```

**Aturan:**
- Simpanan wajib **tidak ada denda/bunga** atas keterlambatan — berbeda dari pinjaman
- Grace period hanya menentukan **kapan masuk kategori tunggakan** (arrears)
- Tunggakan mempengaruhi **laporan compliance** dan potensi sanksi administratif (bukan finansial)
- Tenant bisa mengonfigurasi **sanksi non-finansial** untuk tunggakan berlebih (misal: suspend hak pinjaman)
- Nasabah dengan tunggakan > N bulan (configurable) bisa di-flag untuk review keanggotaan

### 6. Student Considerations

Siswa/santri memiliki perlakuan khusus karena usia, kemampuan finansial, dan cara collection:

```
student_config:
  simpanan_pokok_amount:   25000     # Lebih kecil dari dewasa
  simpanan_wajib_amount:   10000     # Lebih kecil dari dewasa
  collection_mode:         "weekly"  # Bisa weekly, bukan monthly
  weekly_amount:            2500     # wajib bulanan / 4 minggu
  collector:               "wali_kelas"
  parent_can_pay:          true
  auto_debit_from_parent:  false     # Potong dari tabungan orang tua (opt-in)
```

**Weekly collection untuk siswa:**
```
Minggu 1 (wali kelas kumpulkan): Rp 2.500
Minggu 2 (wali kelas kumpulkan): Rp 2.500
Minggu 3 (wali kelas kumpulkan): Rp 2.500
Minggu 4 (wali kelas kumpulkan): Rp 2.500
────────────────────────────────────────
Total bulan ini:                  Rp 10.000
→ Wali kelas setor batch ke koperasi akhir bulan
→ Teller proses sebagai 1 transaksi per siswa per bulan
```

**Aturan:**
- Nominal simpanan pokok & wajib **bisa lebih kecil** untuk student (configurable per school_relation_type)
- Collection bisa **weekly** — wali kelas mengumpulkan, menyetor batch ke koperasi
- Pembayaran oleh orang tua **tercatat di rekening anak**, dengan `paid_by_nasabah_id` = ID orang tua
- Siswa yang **lulus/pindah** → simpanan dikembalikan (Section 7)
- Grace period bisa **lebih longgar** untuk siswa (configurable per school_relation_type)

### 7. Refund on Exit — Pengembalian Saat Keluar

Saat nasabah keluar (deactivation), simpanan pokok dan wajib **wajib dikembalikan**:

```
Nasabah mengajukan keluar
        │
        v
┌─────────────────────────────────┐
│ Deactivation Pre-check (K001)   │
│ - Semua saldo rekening = 0?     │
│ - Simpanan pokok & wajib = 0?   │
└────────┬────────────────────────┘
         │ Ada saldo > 0
         v
┌─────────────────────────────────┐
│ Proses Refund:                  │
│ 1. Hitung total saldo pokok     │
│ 2. Hitung total saldo wajib     │
│ 3. Cek potongan (jika ada):     │
│    - Tunggakan yang belum bayar  │
│    - Biaya admin penutupan       │
│ 4. Net refund = total - potongan │
│ 5. Buat transaksi REFUND        │
│ 6. Saldo → 0                    │
│ 7. Rekening → CLOSED            │
└────────┬────────────────────────┘
         │
         v
┌─────────────────────────────────┐
│ Pembayaran refund:              │
│ - Tunai via teller              │
│ - Transfer ke rekening bank     │
│ - Pindah ke tabungan (jika      │
│   masih aktif sementara)        │
└─────────────────────────────────┘
```

**Aturan:**
- Refund adalah **hak nasabah** — wajib dikembalikan penuh (dikurangi potongan yang sah)
- Potongan yang diperbolehkan:
  - Tunggakan simpanan wajib yang belum lunas
  - Biaya admin penutupan (jika dikonfigurasi di produk)
  - Kewajiban lain yang tercatat (denda pinjaman, dll)
- Refund membutuhkan approval **Manager+**
- Transaksi refund tercatat detail: nominal awal, potongan, net refund
- Setelah refund, rekening simpanan pokok & wajib di-close — saldo = 0
- **Sequence**: refund simpanan pokok & wajib harus dilakukan **setelah** semua rekening lain sudah closed (tabungan = 0, deposito dicairkan, pinjaman lunas)

### 8. Kontribusi ke SHU (Referensi K016)

Simpanan pokok dan wajib menjadi **basis perhitungan SHU** (Sisa Hasil Usaha) tahunan:

```
SHU Distribution Formula:
├── Jasa Modal: proporsional terhadap simpanan pokok + wajib
│   └── share_nasabah = (simpanan_pokok + simpanan_wajib) / total_modal_anggota
│
├── Jasa Usaha: proporsional terhadap volume transaksi
│   └── (Dihitung di K016)
│
└── Total SHU Nasabah = (jasa_modal × porsi_modal) + (jasa_usaha × porsi_usaha)
```

**Aturan:**
- Saldo simpanan pokok & wajib **per akhir tahun buku** yang digunakan untuk kalkulasi
- Nasabah yang join mid-year dihitung **prorated** (bulan aktif / 12)
- Detail formula dan distribusi SHU didokumentasikan di [ADR-K016](./ADR-K016-shu.md)
- Field `shu_eligible` di rekening: true jika rekening sudah ada setoran (saldo > 0)

### 9. Reporting

Laporan yang berkaitan dengan simpanan pokok & wajib:

```
┌─────────────────────────────────────────────────────────────┐
│                      LAPORAN                                 │
├──────────────────┬──────────────────────────────────────────┤
│ Laporan          │ Keterangan                               │
├──────────────────┼──────────────────────────────────────────┤
│ Collection       │ Rekap tagihan bulan ini: paid/unpaid/    │
│ Summary          │ partial per branch. Filter by period.    │
├──────────────────┼──────────────────────────────────────────┤
│ Arrears List     │ Daftar nasabah dengan tunggakan wajib    │
│                  │ > grace period. Sortable by months due.  │
├──────────────────┼──────────────────────────────────────────┤
│ Compliance Rate  │ % nasabah yang bayar tepat waktu         │
│                  │ per branch per bulan.                    │
├──────────────────┼──────────────────────────────────────────┤
│ Per-Class Report │ Rekap per kelas (khusus student) —       │
│                  │ untuk wali kelas.                        │
├──────────────────┼──────────────────────────────────────────┤
│ Total Modal      │ Total simpanan pokok + wajib seluruh     │
│ Anggota          │ nasabah. Untuk neraca koperasi.          │
├──────────────────┼──────────────────────────────────────────┤
│ Refund Report    │ Daftar pengembalian simpanan pokok &     │
│                  │ wajib dalam periode tertentu.            │
└──────────────────┴──────────────────────────────────────────┘
```

**Aturan:**
- Semua laporan bisa di-filter per branch, per period, per school_relation_type
- Export ke PDF dan Excel
- Compliance rate dihitung: `(jumlah_paid / jumlah_total_billing) × 100%`
- Per-class report menjadi alat kerja wali kelas untuk tracking collection

### 10. Auto-Debit Rules

Nasabah yang memiliki rekening tabungan bisa mengaktifkan auto-debit untuk pembayaran simpanan wajib:

```
auto_debit_config (per nasabah):
├── enabled                BOOLEAN
├── source_rekening_id     UUID (FK → rekening tabungan)
├── priority               INTEGER (1 = highest)
├── retry_on_failure       BOOLEAN DEFAULT true
├── max_retry              INTEGER DEFAULT 3
└── retry_interval_days    INTEGER DEFAULT 1
```

**Flow auto-debit:**
```
Tanggal due_date (scheduled job)
        │
        v
┌─────────────────────────────────────┐
│ Untuk setiap billing UNPAID:        │
│ 1. Cek auto_debit enabled?          │
│ 2. Cek source_rekening available?   │
│ 3. Cek available_balance ≥ amount?  │
│    ├── YES → Execute debit          │
│    │   ├── Debit tabungan           │
│    │   ├── Credit simpanan_wajib    │
│    │   ├── Mark billing PAID        │
│    │   └── Kirim notifikasi sukses  │
│    └── NO → Mark attempt failed     │
│        ├── Retry sesuai config      │
│        └── Kirim notifikasi gagal   │
└─────────────────────────────────────┘
```

**Aturan:**
- Auto-debit hanya bisa dari rekening **TABUNGAN** milik nasabah sendiri
- Debit tidak boleh menyebabkan tabungan di bawah `min_balance` produk
- Jika gagal (saldo kurang), retry sesuai config — setelah max retry, tagihan tetap UNPAID
- Notifikasi dikirim baik saat **sukses maupun gagal**
- Auto-debit bisa di-nonaktifkan kapan saja oleh nasabah (via teller)
- Priority: jika ada multiple auto-debit (wajib + lainnya), simpanan wajib diprioritaskan
- Auto-debit **bukan satu-satunya channel** — nasabah tetap bisa bayar manual walau auto-debit aktif

### 11. Data Model

Simpanan pokok dan wajib menggunakan **rekening** dari K002 (tabel `rekening`) dengan `category = simpanan_pokok` atau `simpanan_wajib`. Tidak ada tabel terpisah untuk simpanan itu sendiri — yang spesifik adalah tabel billing dan auto-debit config.

#### 11a. Tabel `simpanan_wajib_billing` (sudah di Section 4)

Lihat data model di Section 4.

#### 11b. Tabel `auto_debit_config`

```
auto_debit_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah)
├── target_rekening_id    UUID (FK → rekening, simpanan_wajib)
├── source_rekening_id    UUID (FK → rekening, tabungan)
│
├── ── Config ──
├── enabled               BOOLEAN DEFAULT true
├── priority              INTEGER DEFAULT 1
├── retry_on_failure      BOOLEAN DEFAULT true
├── max_retry             INTEGER DEFAULT 3
├── retry_interval_days   INTEGER DEFAULT 1
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

#### 11c. Vernon _rels dan _data Structure

**simpanan_wajib_billing _rels:**
```json
{
  "tenant_id":    "018f...",
  "nasabah_id":   "018f...",
  "rekening_id":  "018f..."
}
```

**simpanan_wajib_billing _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001",
    "school_relation_type": "teacher"
  },
  "rekening": {
    "id":             "018f...",
    "account_number": "SW-2026-JKT-00000001"
  }
}
```

**auto_debit_config _rels:**
```json
{
  "tenant_id":          "018f...",
  "nasabah_id":         "018f...",
  "target_rekening_id": "018f...",
  "source_rekening_id": "018f..."
}
```

**auto_debit_config _data:**
```json
{
  "nasabah": {
    "id":            "018f...",
    "full_name":     "Ahmad Fauzi",
    "member_number": "KOP-2026-JKT-000001"
  },
  "target_rekening": {
    "id":             "018f...",
    "account_number": "SW-2026-JKT-00000001",
    "category":       "simpanan_wajib"
  },
  "source_rekening": {
    "id":             "018f...",
    "account_number": "TB-2026-JKT-00000001",
    "category":       "tabungan"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.nasabah` di billing & auto_debit_config
- `RekeningUpdatedEvent` → update `_data.rekening` / `_data.target_rekening` / `_data.source_rekening`

### 12. Authorization — RBAC

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| View tagihan simpanan wajib | v | v | v | v |
| View arrears list | v | v | v | v |
| Catat setoran pokok (teller cash) | v | v | v | v |
| Catat setoran wajib (teller cash) | v | v | v | v |
| Proses batch wali kelas | v | v | v | v |
| Proses batch payroll settlement | - | v | v | v |
| Setup auto-debit nasabah | v | v | v | v |
| Waive tagihan simpanan wajib | - | - | v | v |
| Proses refund simpanan | - | - | v | v |
| Generate laporan collection | v | v | v | v |
| Generate laporan arrears | - | v | v | v |
| Ubah nominal simpanan pokok/wajib | - | - | - | v |
| Batch top-up simpanan pokok | - | - | v | v |

**Catatan:**
- Teller bisa catat setoran dan proses batch wali kelas — operasi day-to-day
- Payroll settlement batch butuh **Supervisor+** karena melibatkan banyak nasabah sekaligus
- Waive tagihan hanya **Manager+** — keputusan kebijakan, bukan operasional
- Refund hanya **Manager+** — operasi sensitif yang mengurangi modal koperasi
- Ubah nominal hanya **Admin** — berdampak ke seluruh nasabah, harus sesuai AD/ART

### 13. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Simpanan Pokok / Simpanan Wajib | Simpanan Pokok / Simpanan Wajib |
| Akad | N/A | Wadiah Yad Dhamanah |
| Receipt title | Bukti Setoran Simpanan | Bukti Setoran Simpanan + Akad Wadiah |
| Refund label | Pengembalian Simpanan | Pengembalian Simpanan |
| Bonus (jika ada) | Bunga (jika dikonfigurasi) | Athaya (bonus sukarela, bukan wajib) |
| SHU contribution | Kontribusi Modal | Kontribusi Modal |
| Report title | Laporan Simpanan Pokok & Wajib | Laporan Simpanan Pokok & Wajib |

**Catatan:**
- Simpanan pokok & wajib **sama di kedua mode** — perbedaan utama hanya di akad (Wadiah untuk BMT)
- BMT mode: bonus atas simpanan bersifat **sukarela** (athaya), bukan bunga
- Pengembalian saat keluar tidak berbeda antar mode — nominal dikembalikan penuh

## Consequences

### Positif

- **Compliant** — memenuhi UU Perkoperasian dan AD/ART tentang simpanan wajib keanggotaan
- **Flexible collection** — multiple channel pembayaran mengakomodasi semua jenis anggota (siswa, guru, orang tua)
- **Trackable** — billing per bulan memungkinkan tracking arrears dan compliance rate
- **Student-friendly** — weekly collection via wali kelas sesuai realitas sekolah
- **Auto-debit** — mengurangi beban manual collection untuk nasabah yang punya tabungan
- **Auditable** — setiap pembayaran, waive, dan refund tercatat lengkap
- **SHU-ready** — saldo simpanan menjadi basis kalkulasi pembagian SHU

### Negatif

- **Billing generation overhead** — scheduled job harus generate billing untuk semua nasabah aktif setiap bulan
- **Batch processing complexity** — wali kelas collection dan payroll membutuhkan validasi batch yang ketat
- **Auto-debit failure handling** — retry mechanism menambah complexity dan potensi edge case
- **Multi-amount configuration** — nominal berbeda per school_relation_type menambah konfigurasi

### Mitigasi

- Billing generation menggunakan **batch insert** — efisien bahkan untuk ribuan nasabah
- Batch processing dilengkapi **preview + validation report** sebelum eksekusi — operator bisa review dulu
- Auto-debit failure di-log detail dan notifikasi dikirim — nasabah/operator bisa tindak lanjut manual
- Default nominal berlaku untuk semua type — konfigurasi per-type hanya jika tenant membutuhkan (opt-in)

## Alternatives Considered

### A. Simpanan Wajib Tanpa Billing (hanya tracking saldo)

Tidak ada tagihan per bulan — hanya melihat saldo kumulatif simpanan wajib.

**Ditolak karena:** tanpa billing, tidak ada mekanisme untuk tracking siapa yang sudah bayar bulan ini dan siapa yang belum. Arrears tracking menjadi tidak mungkin, dan compliance reporting tidak akurat.

### B. Denda Finansial untuk Keterlambatan Simpanan Wajib

Mengenakan bunga/denda atas keterlambatan pembayaran simpanan wajib.

**Ditolak karena:** simpanan wajib bukan pinjaman — tidak seharusnya ada bunga. Denda finansial juga bermasalah di mode BMT (riba). Sanksi non-finansial (suspend hak pinjaman, review keanggotaan) lebih sesuai.

### C. Collection Hanya via Teller (tanpa auto-debit dan batch)

Semua pembayaran harus melalui teller secara individual.

**Ditolak karena:** tidak scalable untuk koperasi sekolah dengan ratusan siswa. Wali kelas collection dan auto-debit sangat mengurangi beban operasional. Tanpa batch processing, teller harus input satu per satu — error-prone dan memakan waktu.

# 02 — Integrasi Sekolah ↔ Koperasi

Dokumen ini menjelaskan semua titik integrasi antara domain Sekolah dan domain Koperasi: alur event, aturan sinkronisasi, portal terpadu, notifikasi terpadu, dan sistem user/role yang dipakai bersama.

---

## 1. Siswa Sekolah → Nasabah Koperasi

### Prinsip Dasar

Sekolah adalah **source of truth** untuk data person. Koperasi **tidak menduplikasi** data siswa — hanya menyimpan referensi `school_entity_id` yang menunjuk ke `students.id`.

```
students (domain Sekolah)          nasabah (domain Koperasi)
─────────────────────────          ─────────────────────────
id (UUID v7) ◀─────────────────── school_entity_id (UUID, nullable)
full_name                          full_name (copy saat pendaftaran)
nis                                member_number (auto-generate KOP-...)
status = 'active'                  status = 'active'
...                                school_relation_type = 'student'
```

### Alur Event: Siswa Aktif → Pendaftaran Nasabah

```
SEKOLAH DOMAIN
──────────────
Siswa di-enroll (PPDB) atau status diubah ke 'active'
  │
  ▼
Command: CreateStudentCommand / EnrollApplicantCommand
  │ (setelah write ke DB berhasil)
  ▼
eventBus.Publish("events.students.student_activated", StudentActivatedEvent)
  │
  │ [ASYNC — tidak menunggu]
  ▼
NATS JetStream Stream: DOMAIN_EVENTS
  │
  ▼
KOPERASI DOMAIN (Consumer)
──────────────────────────
KoperasiStudentEnrollmentHandler.Handle(event)
  │
  ├─ [Mode: AUTO]
  │    CREATE nasabah_application {
  │      school_entity_id: event.StudentID,
  │      school_relation_type: 'student',
  │      full_name: event.FullName,
  │      status: 'pending'  ← WAJIB melalui approval flow
  │    }
  │    Notify Supervisor Koperasi: "Siswa baru menunggu pendaftaran anggota"
  │
  └─ [Mode: MANUAL]
       Notify Supervisor Koperasi: "Siswa baru siap didaftarkan sebagai anggota"
       (tidak otomatis buat application)
```

### Payload Event `StudentActivated`

```json
{
  "id": "018f-event-uuid-v7",
  "type": "student_activated",
  "aggregate_id": "018f-student-uuid",
  "occurred_at": "2026-04-15T08:00:00Z",
  "payload": {
    "student_id": "018f-student-uuid",
    "tenant_id": "tenant-uuid",
    "company_id": "company-uuid",
    "full_name": "Ahmad Fauzi",
    "nis": "2025001",
    "nisn": "0098765432",
    "gender": "L",
    "birth_date": "2012-05-15",
    "class_room_id": "018f-class-uuid",
    "academic_year_id": "018f-ay-uuid",
    "activation_source": "ppdb_enrollment"
  }
}
```

### Aturan Penting

| Aturan | Detail |
|--------|--------|
| Mode enrollment dikonfigurasi per tenant | `auto` atau `manual` (default: `manual`) |
| Mode `auto` tetap melalui approval | Application dibuat dengan status `pending`, tidak langsung aktif |
| Sekolah tidak butuh ACK dari Koperasi | Fire-and-forget — Koperasi consumer boleh down sementara |
| Idempotency | Consumer cek `school_entity_id` — tidak buat duplikat jika event diproses ulang |
| Saat siswa non-aktif | Event `StudentDeactivated` → notifikasi Koperasi untuk review (tidak otomatis tutup rekening) |

### Status Sinkronisasi Siswa ↔ Nasabah

| Event Sekolah | Aksi Koperasi | Otomatis? |
|--------------|---------------|-----------|
| `StudentActivated` | Buat `nasabah_application` | Ya (mode auto) / Notif saja (mode manual) |
| `StudentUpdated` (nama berubah) | Update `nasabah.full_name` | Ya (via SyncEngine jika FK digunakan) |
| `StudentDeactivated` (graduated/transferred) | Notifikasi Koperasi untuk tindak lanjut | Tidak — butuh keputusan manusia |
| `StudentExpelled` | Alert Koperasi — blokir pinjaman baru | Tidak — butuh keputusan manusia |

---

## 2. Guru/Staff → Payroll Deduction Koperasi

### Prinsip

Guru terdaftar di domain Sekolah. Koperasi mengonsumsi data guru untuk keperluan **potongan gaji otomatis** (angsuran pinjaman, simpanan wajib).

```
teachers (domain Sekolah)          nasabah (domain Koperasi)
─────────────────────────          ─────────────────────────
id (UUID v7) ◀─────────────────── school_entity_id (UUID)
full_name                          school_relation_type = 'teacher'
nip / nuptk                        ...
employee_type (pns/honorer/dll)
```

### Alur Payroll Deduction

```
HR Sekolah input data gaji bulan ini
  │
  ▼
POST /api/v1/payroll/batches  (domain Sekolah atau HR system eksternal)
  │
  ▼
Koperasi menerima payroll_batch data via:
  ├─ A) Direct API call ke endpoint Koperasi (inbound, butuh API Key + IP whitelist)
  └─ B) Domain Event: PayrollDataReadyEvent (ASYNC via NATS)

  ▼
Koperasi (K020) proses payroll_batch:
  │
  ├─ Lookup nasabah dengan school_entity_id = teacher.id
  ├─ Cek payroll_authorization (otorisasi tertulis dari nasabah)
  ├─ Hitung total potongan: angsuran + simpanan_wajib + tabungan_autorekt
  ├─ Validasi: potongan tidak melebihi take_home_pay
  │
  ▼
Execute deductions (atomic per nasabah):
  INSERT transaksi (setoran/angsuran) + UPDATE rekening.balance
  UPDATE payroll_deduction.status = 'processed'
  │
  ▼
Kirim notifikasi ke guru: "Potongan gaji Anda bulan ini: Rp X"
```

### Payload Event `PayrollDataReady`

```json
{
  "id": "018f-event-uuid",
  "type": "payroll_data_ready",
  "occurred_at": "2026-04-30T00:00:00Z",
  "payload": {
    "batch_id": "018f-batch-uuid",
    "tenant_id": "tenant-uuid",
    "company_id": "company-uuid",
    "period_month": 4,
    "period_year": 2026,
    "items": [
      {
        "teacher_id": "018f-teacher-uuid",
        "nip": "198501012010012001",
        "full_name": "Ibu Siti Rahayu",
        "gross_salary": 8500000,
        "take_home_pay": 7200000
      }
    ]
  }
}
```

### Aturan Potongan Gaji

| Aturan | Detail |
|--------|--------|
| Wajib ada `payroll_authorization` tertulis | Guru harus menandatangani otorisasi sebelum auto-deduction bisa berjalan |
| Potongan tidak boleh > take_home_pay | Validasi di Koperasi sebelum eksekusi |
| Fallback jika gaji tidak cukup | Catat sebagai tunggakan, kirim notifikasi ke nasabah dan Supervisor |
| Revoke authorization | Guru bisa cabut otorisasi — efektif bulan berikutnya |

---

## 3. Kantin Sekolah (S036-S037) vs Toko Koperasi (K019)

### Perbedaan Fundamental

| Aspek | Kantin Sekolah (S036-S037) | Toko Koperasi (K019) |
|-------|--------------------------|----------------------|
| Pengelola | Bagian keuangan sekolah | Unit usaha Koperasi |
| Pembayaran | Dompet digital siswa (canteen_wallet) | Rekening tabungan nasabah |
| Pencatatan keuangan | SPP module (S009) | Jurnal Koperasi (K015) |
| User yang dilayani | Siswa (via NIS) | Nasabah (via nomor rekening) |
| Spending control | Parent portal Sekolah (S042) | Parent portal Koperasi (K023) |
| Laporan | Reporting Sekolah (S054) | Laporan Koperasi (K017) |
| Regulasi | Manajemen sekolah internal | UU Koperasi No. 25/1992 |

### Overlap: Siswa Belanja di Toko Koperasi

Ketika sekolah menggunakan Toko Koperasi untuk kantin:

```
Siswa → scan QR / tap kartu di kasir Toko Koperasi
  │
  ▼
POS Koperasi (K019) lookup nasabah via school_entity_id = students.id
  │
  ▼
Cek rekening tabungan siswa (via nasabah → rekening)
Cek spending_limit dari parent_portal config (K023)
  │
  ▼
Jika OK: debit rekening tabungan siswa
  INSERT toko_penjualan
  UPDATE rekening.balance (atomic)
  INSERT transaksi (debit tabungan)
  │
  ▼
Notifikasi ke orang tua via K022:
  "Anak Anda membeli [item] Rp X di kantin sekolah"
```

### Rekomendasi Arsitektur

- **Jika sekolah memiliki kantin terpisah dari koperasi:** gunakan S036-S037 (canteen_wallet terpisah)
- **Jika koperasi mengelola kantin:** gunakan K019 (rekening tabungan siswa sebagai dompet)
- **Tidak direkomendasikan menjalankan keduanya bersamaan** untuk satu sekolah — akan membingungkan laporan keuangan

---

## 4. SPP (S009) vs Simpanan/Tabungan Koperasi

### Perbedaan Fundamental

| Aspek | SPP / Iuran Sekolah (S009) | Simpanan Koperasi (K004-K005) |
|-------|--------------------------|-------------------------------|
| Sifat | Biaya pendidikan — wajib dibayar | Simpanan anggota — milik anggota |
| Kepemilikan dana | Sekolah (pendapatan sekolah) | Koperasi (dikelola, bukan milik koperasi) |
| Dapat ditarik | Tidak (biaya, bukan tabungan) | Ya (simpanan sukarela/tabungan) |
| Regulasi | Peraturan sekolah + BOS | UU Koperasi No. 25/1992 |
| Laporan | Laporan keuangan sekolah | Laporan keuangan koperasi |
| Jika tidak dibayar | Invoice overdue → notifikasi orang tua | Simpanan wajib → tunggakan keanggotaan |

### Titik Overlap: Orang Tua Bayar SPP via Koperasi

Beberapa koperasi sekolah memfasilitasi pembayaran SPP siswa dari tabungan koperasi:

```
SKENARIO: Orang tua ingin bayar SPP dari saldo tabungan koperasi anaknya

Orang tua request via parent portal (K023)
  │
  ▼
Koperasi verifikasi: saldo tabungan cukup?
  │
  ▼
Koperasi DEBIT tabungan siswa
  INSERT transaksi (debit) di Koperasi
  UPDATE rekening.balance
  │
  ▼
Koperasi TRANSFER ke rekening bank sekolah (via teller)
  ATAU Koperasi kirim notifikasi ke bendahara sekolah
  │
  ▼
Bendahara Sekolah konfirmasi penerimaan
  UPDATE student_invoices.status = 'paid' (domain Sekolah)
  INSERT payment_transaction (S051) — metode: koperasi_transfer
```

**Catatan Penting:** Ini adalah alur MANUAL yang membutuhkan koordinasi teller/bendahara. Sistem tidak otomatis memotong tabungan koperasi untuk SPP sekolah karena keduanya adalah entitas legal yang terpisah.

---

## 5. Portal Orang Tua Terpadu

### Masalah Tanpa Integrasi

Tanpa integrasi, orang tua harus:
1. Login ke portal sekolah → lihat nilai, absensi, SPP
2. Login ke portal koperasi → lihat saldo tabungan anak, riwayat belanja

### Solusi: Single Login, Context Switching

```
Orang tua login sekali → JWT Phase 1 → Pilih company
  │
  ▼
JWT Phase 2 berisi:
  {
    "sub": "user-uuid",
    "phase": 2,
    "tenant_id": "...",
    "company_id": "...",
    "roles": ["parent"],
    "permissions": [
      "students:read",           ← akses portal sekolah
      "student_invoices:read",
      "koperasi_portal:read",    ← akses portal koperasi
      "koperasi_balance:read"
    ],
    "student_ids": ["018f-student-uuid"],   ← anak yang bisa dilihat
    "nasabah_ids": ["018f-nasabah-uuid"]    ← rekening koperasi anak
  }
  │
  ▼
Frontend menampilkan unified parent dashboard dengan tab:
  ├── Tab Sekolah: Nilai · Absensi · SPP · Pengumuman
  └── Tab Koperasi: Saldo · Belanja Kantin · Tabungan · Spending Limit
```

### Data yang Ditampilkan per Tab

**Tab Sekolah (dari domain Sekolah):**
- Rekap akademik semester (S004)
- Absensi bulan ini (S008)
- Invoice SPP yang belum dibayar (S009)
- Pengumuman sekolah terbaru (S045)
- Nilai terbaru (S011) — jika sudah dipublikasikan

**Tab Koperasi (dari domain Koperasi):**
- Saldo tabungan anak (K002 rekening)
- Riwayat belanja hari ini / minggu ini (K019 toko)
- Spending limit dan sisa limit (K023)
- Status pinjaman jika ada (K007)

### Aturan Keamanan

| Aturan | Detail |
|--------|--------|
| Row-level security per domain | Portal Sekolah hanya baca data `student_ids` yang ter-link; Portal Koperasi hanya baca `nasabah_ids` yang ter-link |
| Link guardian ↔ nasabah via `school_entity_id` | Satu parent user bisa punya multiple children, masing-masing punya rekening koperasi sendiri |
| Koperasi tidak expose data keuangan ke Sekolah | Tab Koperasi hanya visible jika orang tua punya permission `koperasi_portal:read` |

---

## 6. Konsolidasi Notifikasi

### Masalah Tanpa Konsolidasi

Tanpa koordinasi, orang tua bisa menerima notifikasi duplikat atau berlebihan:
- Sekolah kirim: "SPP bulan ini belum dibayar"
- Koperasi kirim: "Saldo tabungan anak Anda Rp 50.000"
- Sekolah kirim: "Anak Anda tidak hadir hari ini"
- Koperasi kirim: "Batas belanja harian hampir tercapai"

### Strategi Konsolidasi

```
PRINSIP: Setiap notifikasi menggunakan idempotency_key yang konsisten

Format idempotency_key:
  "{domain}.{event_type}:{entity_id}:{user_id}:{date}"

Contoh:
  "sekolah.attendance.absent:student-123:parent-456:2026-04-15"
  "koperasi.canteen.daily_summary:nasabah-789:parent-456:2026-04-15"
```

### Quiet Hours Terpadu

Kedua domain menghormati `quiet_hours` yang dikonfigurasi user:
- Default: 22:00 – 06:00 WIB
- CRITICAL event dari kedua domain selalu dikirim (tidak di-queue)
- MEDIUM/LOW event di-queue dan dikirim setelah quiet hours

### Event yang Berpotensi Spam (Perlu Throttle)

| Event | Frekuensi Potensial | Solusi |
|-------|---------------------|--------|
| Setiap transaksi belanja kantin | Per belanja (bisa 3-5x/hari) | Kirim ringkasan harian, bukan per transaksi |
| Setiap update saldo rekening | Per transaksi | Throttle: 1 notif/jam untuk update saldo kecil |
| Angsuran jatuh tempo (H-7, H-3, H-1) | Per hari | Kirim sekali per H, tidak berkali-kali |
| Absensi siswa | Per hari | 1 notifikasi per hari per siswa |

---

## 7. Sistem User/Role Bersama (Core ADR-013)

### Satu User, Dua Domain

User yang sama dapat memiliki role di domain Sekolah DAN Koperasi:

```
Contoh: Bu Siti adalah Guru Sekolah DAN Teller Koperasi

users tabel (Core):
  id: 018f-user-siti
  email: siti@smpalbarakah.sch.id

user_roles tabel (Core):
  user_id: 018f-user-siti, role_id: "homeroom_teacher", company_id: "company-a"
  user_id: 018f-user-siti, role_id: "koperasi_teller",  company_id: "company-a"

JWT Phase 2 saat Bu Siti login:
  {
    "roles": ["homeroom_teacher", "koperasi_teller"],
    "permissions": [
      "students:read", "attendance:write", "grades:write",  ← dari role guru
      "transaksi:write", "nasabah:read", "teller_session:write" ← dari role teller
    ],
    "teacher_id": "018f-teacher-siti",
    "class_room_ids": ["018f-class-viia"],
    "koperasi_branch_id": "018f-branch-main"
  }
```

### Role Default per Domain

**Sekolah (di-seed saat company dibuat):**
```
admin, principal, vice_principal, homeroom_teacher, teacher,
counselor, finance, staff, parent
```

**Koperasi (di-seed saat company dibuat):**
```
koperasi_admin, koperasi_manager, koperasi_supervisor,
koperasi_teller, koperasi_member
```

### Pemisahan Context di Frontend

Meski permissions ada dalam satu JWT, frontend menampilkan UI yang terpisah per context:
- URL `/sekolah/...` → tampilkan menu dan fitur sekolah
- URL `/koperasi/...` → tampilkan menu dan fitur koperasi
- User bisa switch context via navigation

### Implikasi Keamanan

| Aturan | Detail |
|--------|--------|
| Koperasi admin TIDAK otomatis jadi Sekolah admin | Meski satu user, role assignment terpisah |
| Kepala Sekolah mendapat dashboard overview koperasi (read-only) | Untuk oversight, bukan operasional |
| Parent role mendapat akses ke dua portal | Parent portal Sekolah + Parent portal Koperasi (jika anaknya nasabah) |

---

## 8. Katalog Event Lintas Domain (Sekolah ↔ Koperasi)

### Event dari Sekolah → Koperasi

| Event | Subject NATS | Trigger | Consumer Koperasi | Tipe |
|-------|-------------|---------|-------------------|------|
| `StudentActivated` | `events.students.student_activated` | Siswa status → active | Buat nasabah_application | ASYNC |
| `StudentDeactivated` | `events.students.student_deactivated` | Siswa lulus/pindah/keluar | Notifikasi review rekening | ASYNC |
| `StudentUpdated` | `events.students.student_updated` | Perubahan data siswa | Update nasabah.full_name | ASYNC |
| `TeacherPayrollDataReady` | `events.teachers.payroll_data_ready` | Data gaji bulan ini tersedia | Proses payroll deduction | ASYNC |
| `TeacherDeactivated` | `events.teachers.teacher_deactivated` | Guru tidak aktif | Notifikasi review nasabah | ASYNC |

### Event dari Koperasi → Sekolah

| Event | Subject NATS | Trigger | Consumer Sekolah | Tipe |
|-------|-------------|---------|-----------------|------|
| `NasabahApproved` | `events.koperasi.nasabah_approved` | Nasabah baru disetujui | Update student._data (optional enrichment) | ASYNC |
| `CanteenPurchase` | `events.koperasi.canteen_purchase` | Siswa beli di toko koperasi | Update kantin reporting (jika terpadu) | ASYNC |
| `PayrollDeductionProcessed` | `events.koperasi.payroll_processed` | Potongan gaji dieksekusi | Update payroll record Sekolah (opsional) | ASYNC |
| `SaldoRendahAlert` | `events.koperasi.saldo_rendah` | Saldo anak mendekati 0 | Notifikasi ke parent portal Sekolah (opsional) | ASYNC |

### Catatan Critical vs Best-Effort

| Event | Kategori | Alasan |
|-------|----------|--------|
| `StudentActivated` | **CRITICAL** | Kehilangan event = siswa tidak terdaftar di koperasi |
| `TeacherPayrollDataReady` | **CRITICAL** | Kehilangan event = potongan gaji tidak diproses bulan ini |
| `StudentUpdated` | Best-Effort | Data stale di koperasi acceptable (sync dalam <30 detik) |
| `CanteenPurchase` | Best-Effort | Laporan bisa direkonsiliasi ulang |

---

## Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|--------|--------|---------|
| Event `StudentActivated` hilang | Siswa tidak jadi nasabah koperasi | NATS JetStream at-least-once; CRITICAL priority di retry queue |
| Double processing `StudentActivated` | Duplikat nasabah_application | Idempotency check via `school_entity_id` sebelum insert |
| Guru payroll deduction double | Dua kali potong gaji bulan yang sama | Idempotency key per `batch_id + teacher_id + period` |
| Portal orang tua expose data lintas tenant | Data bocor ke orang tua yang salah | Row-level security ketat via `student_ids` dan `nasabah_ids` di JWT |
| Notifikasi spam ke orang tua | Pengalaman buruk | Rate limiting + daily digest untuk non-CRITICAL events |
| Koperasi down saat siswa di-enroll massal | Backlog ribuan aplikasi | NATS JetStream buffer + auto-retry saat Koperasi recovery |

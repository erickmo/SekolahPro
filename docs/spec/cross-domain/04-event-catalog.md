# 04 — Katalog Event Sistem SekolahPro

Dokumen ini mendaftar semua domain event signifikan di seluruh sistem SekolahPro: Core, Sekolah, dan Koperasi. Mencakup naming convention, payload, trigger, consumer, dan klasifikasi kekritisan.

---

## Konvensi Penamaan Event

### Subject NATS

```
{stream_prefix}.{domain}.{event_type}

Domain Events:
  events.{domain}.{entity}_{action}

Vernon Sync Events:
  sync.{target_table}.{source_entity}_updated

Contoh:
  events.students.student_created
  events.students.student_activated
  events.koperasi.nasabah_approved
  sync.class_rooms.teacher_updated
  sync.student_grades.class_room_updated
```

Semua subject harus didefinisikan sebagai **Go constants** (bukan string literal) di `domain/event/subjects.go`:

```go
const (
  SubjectStudentCreated    = "events.students.student_created"
  SubjectStudentActivated  = "events.students.student_activated"
  SubjectNasabahApproved   = "events.koperasi.nasabah_approved"
  // ... dst
)
```

### Nama Event (Type Field)

Format: `{entity}_{past_tense_verb}`

```
student_created, student_activated, student_deactivated
nasabah_approved, nasabah_deactivated
teacher_updated, teacher_payroll_data_ready
invoice_paid, invoice_overdue
transaksi_created, angsuran_overdue
class_placement_created, academic_year_activated
```

---

## Klasifikasi Kekritisan

| Level | Definisi | Strategi Delivery |
|-------|----------|------------------|
| **CRITICAL** | Kehilangan event = data loss atau kegagalan proses bisnis yang tidak bisa dipulihkan otomatis | NATS JetStream durable, DLQ, manual replay, alert monitoring |
| **IMPORTANT** | Kehilangan event = inkonsistensi data yang akan dideteksi, bisa dipulihkan dengan resync | NATS JetStream durable, retry 3x, DLQ |
| **BEST-EFFORT** | Kehilangan event = stale cache atau delayed notification — tidak menyebabkan data corruption | InMemory OK, NATS non-durable untuk scale-down |

---

## A. Core / Shared Events

### A1. Academic Year Events

| Event | Subject | Trigger | Kekritisan |
|-------|---------|---------|------------|
| `AcademicYearCreated` | `events.academic_years.academic_year_created` | Tahun ajaran baru dibuat | Important |
| `AcademicYearActivated` | `events.academic_years.academic_year_activated` | Tahun ajaran di-set active | **CRITICAL** |
| `AcademicYearClosed` | `events.academic_years.academic_year_closed` | Tahun ajaran ditutup | **CRITICAL** |
| `AcademicYearUpdated` | `events.academic_years.academic_year_updated` | Data tahun ajaran diupdate | Important |

**Payload `AcademicYearActivated`:**
```json
{
  "id": "018f-event-uuid",
  "type": "academic_year_activated",
  "aggregate_id": "018f-ay-uuid",
  "occurred_at": "2026-07-01T00:00:00Z",
  "payload": {
    "academic_year_id": "018f-ay-uuid",
    "tenant_id": "tenant-uuid",
    "company_id": "company-uuid",
    "name": "2026/2027",
    "code": "2627",
    "start_date": "2026-07-14",
    "end_date": "2027-06-21",
    "previous_year_id": "018f-prev-ay-uuid"
  }
}
```

**Consumer SyncEngine:**
- `sync.class_rooms.academic_year_updated` → update `_data.academic_year` di class_rooms
- `sync.students.academic_year_updated` → update `_data.academic_year` di students
- `sync.student_grades.academic_year_updated` → update invoices, grades

### A2. Class Room Events

| Event | Subject | Trigger | Kekritisan |
|-------|---------|---------|------------|
| `ClassRoomCreated` | `events.class_rooms.class_room_created` | Kelas baru dibuat | Best-Effort |
| `ClassRoomUpdated` | `events.class_rooms.class_room_updated` | Data kelas diupdate | Important |
| `ClassRoomTeacherAssigned` | `events.class_rooms.teacher_assigned` | Wali kelas diassign | Important |

### A3. Teacher Events

| Event | Subject | Trigger | Kekritisan |
|-------|---------|---------|------------|
| `TeacherCreated` | `events.teachers.teacher_created` | Guru baru didaftarkan | Best-Effort |
| `TeacherUpdated` | `events.teachers.teacher_updated` | Data guru diupdate | Important |
| `TeacherSignatureUploaded` | `events.teachers.teacher_signature_uploaded` | TTD digital diupload | Important |
| `TeacherDeactivated` | `events.teachers.teacher_deactivated` | Guru tidak aktif | Important |
| `TeacherPayrollDataReady` | `events.teachers.payroll_data_ready` | Data gaji tersedia untuk Koperasi | **CRITICAL** |

**Payload `TeacherPayrollDataReady`:**
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
        "full_name": "Ibu Siti Rahayu",
        "nip": "198501012010012001",
        "gross_salary": 8500000,
        "take_home_pay": 7200000
      }
    ]
  }
}
```

### A4. User Events

| Event | Subject | Trigger | Kekritisan |
|-------|---------|---------|------------|
| `UserCreated` | `events.users.user_created` | Akun baru dibuat | Important |
| `UserRoleAssigned` | `events.users.role_assigned` | Role diassign ke user | Important |
| `UserDeactivated` | `events.users.user_deactivated` | Akun dinonaktifkan | Important |

---

## B. Domain Sekolah Events

### B1. Student Core Events

| Event | Subject | Trigger | Producer | Consumer | Kekritisan |
|-------|---------|---------|----------|---------|------------|
| `StudentCreated` | `events.students.student_created` | Siswa baru dibuat | S001 | SyncEngine, Koperasi | Important |
| `StudentActivated` | `events.students.student_activated` | Status → active | S001 | **Koperasi (K001)** | **CRITICAL** |
| `StudentUpdated` | `events.students.student_updated` | Data siswa diupdate | S001 | SyncEngine | Important |
| `StudentDeactivated` | `events.students.student_deactivated` | Status → graduated/transferred/expelled | S001 | Koperasi (K001), Notifikasi | Important |
| `StudentPromoted` | `events.students.student_promoted` | Kenaikan kelas | S014 | SyncEngine, Notifikasi | Important |
| `StudentTransferred` | `events.students.student_transferred` | Mutasi sekolah | S001 | Notifikasi | Important |

**Payload `StudentActivated`:**
```json
{
  "id": "018f-event-uuid",
  "type": "student_activated",
  "aggregate_id": "018f-student-uuid",
  "occurred_at": "2026-07-15T08:00:00Z",
  "payload": {
    "student_id": "018f-student-uuid",
    "tenant_id": "tenant-uuid",
    "company_id": "company-uuid",
    "full_name": "Ahmad Fauzi",
    "nis": "2026001",
    "nisn": "0098765432",
    "gender": "L",
    "birth_date": "2012-05-15",
    "religion": "islam",
    "class_room_id": "018f-class-uuid",
    "academic_year_id": "018f-ay-uuid",
    "guardian_id": "018f-guardian-uuid",
    "guardian_phone": "08123456789",
    "activation_source": "ppdb_enrollment"
  }
}
```

### B2. Attendance Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `StudentAbsent` | `events.attendance.student_absent` | Siswa absen hari ini | Notifikasi (S044) | Important |
| `AttendanceValidated` | `events.attendance.attendance_validated` | Absensi dikunci kepsek | Academic record | Important |
| `AttendanceAggregated` | `events.attendance.attendance_aggregated` | Akhir semester | S004 academic record | Important |

**Payload `StudentAbsent`:**
```json
{
  "type": "student_absent",
  "payload": {
    "student_id": "018f-student-uuid",
    "attendance_id": "018f-attendance-uuid",
    "class_room_id": "018f-class-uuid",
    "attendance_date": "2026-04-15",
    "semester": "ganjil",
    "guardian_ids": ["018f-guardian-uuid"],
    "recorded_by": "018f-teacher-uuid"
  }
}
```

### B3. Finance / SPP Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `InvoiceCreated` | `events.finance.invoice_created` | Invoice SPP digenerate | Notifikasi | Best-Effort |
| `InvoiceDue` | `events.finance.invoice_due` | H-3 sebelum jatuh tempo | Notifikasi (S044) | Important |
| `InvoiceOverdue` | `events.finance.invoice_overdue` | Melewati jatuh tempo | Notifikasi CRITICAL | **CRITICAL** |
| `InvoicePaid` | `events.finance.invoice_paid` | Pembayaran dikonfirmasi | S009, Notifikasi | **CRITICAL** |
| `PaymentConfirmed` | `events.payment.payment_confirmed` | Callback provider berhasil | S009, S037 (kantin) | **CRITICAL** |
| `PaymentExpired` | `events.payment.payment_expired` | VA/QRIS kadaluarsa | Notifikasi | Important |

**Payload `InvoicePaid`:**
```json
{
  "type": "invoice_paid",
  "payload": {
    "invoice_id": "018f-invoice-uuid",
    "student_id": "018f-student-uuid",
    "payment_transaction_id": "018f-txn-uuid",
    "amount_paid": 500000,
    "total_amount": 500000,
    "payment_method": "virtual_account",
    "bank_code": "BNI",
    "paid_at": "2026-04-10T09:30:00Z"
  }
}
```

### B4. Academic Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `GradesPublished` | `events.grades.grades_published` | Guru publish nilai | Notifikasi (S044) | Important |
| `RaporGenerated` | `events.rapor.rapor_generated` | Rapor semester digenerate | Notifikasi, Parent portal | Important |
| `ExamScheduled` | `events.exam.exam_scheduled` | Jadwal ujian dibuat | Notifikasi | Best-Effort |

### B5. PPDB Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `ApplicantRegistered` | `events.ppdb.applicant_registered` | Calon siswa mendaftar | Notifikasi (konfirmasi) | Important |
| `ApplicantAccepted` | `events.ppdb.applicant_accepted` | Siswa diterima | Notifikasi | Important |
| `ApplicantEnrolled` | `events.ppdb.applicant_enrolled` | Daftar ulang selesai → jadi siswa | `StudentActivated` downstream | **CRITICAL** |

### B6. Notification Events (Sekolah)

| Event | Subject | Trigger | Kekritisan |
|-------|---------|---------|------------|
| `DisciplineRecorded` | `events.discipline.discipline_recorded` | Pelanggaran dicatat | Important |
| `AchievementRecorded` | `events.achievement.achievement_recorded` | Prestasi dicatat | Best-Effort |
| `AnnouncementPublished` | `events.announcements.announcement_published` | Pengumuman dipublish | Important |
| `DapodikSyncCompleted` | `events.dapodik.sync_completed` | Sinkronisasi Dapodik selesai | Important |

---

## C. Domain Koperasi Events

### C1. Nasabah (Member) Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `NasabahApplicationCreated` | `events.koperasi.application_created` | Formulir pendaftaran dibuat | Notifikasi Supervisor | Important |
| `NasabahApproved` | `events.koperasi.nasabah_approved` | Nasabah disetujui → aktif | Buka rekening otomatis (K002) | **CRITICAL** |
| `NasabahRejected` | `events.koperasi.nasabah_rejected` | Pendaftaran ditolak | Notifikasi pemohon | Important |
| `NasabahUpdated` | `events.koperasi.nasabah_updated` | Data nasabah diupdate | SyncEngine (_data rekening) | Important |
| `NasabahDeactivated` | `events.koperasi.nasabah_deactivated` | Nasabah keluar | Notifikasi, tutup rekening | **CRITICAL** |
| `KYCUpgraded` | `events.koperasi.kyc_upgraded` | KYC naik ke full | Notifikasi | Best-Effort |

**Payload `NasabahApproved`:**
```json
{
  "type": "nasabah_approved",
  "payload": {
    "nasabah_id": "018f-nasabah-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "018f-branch-uuid",
    "member_number": "KOP-2026-JKT-000001",
    "full_name": "Ahmad Fauzi",
    "school_entity_id": "018f-student-uuid",
    "school_relation_type": "student",
    "approved_by": "018f-supervisor-uuid",
    "approved_at": "2026-07-20T10:00:00Z"
  }
}
```

### C2. Rekening Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `RekeningOpened` | `events.koperasi.rekening_opened` | Rekening baru dibuka | Notifikasi nasabah | Important |
| `RekeningFrozen` | `events.koperasi.rekening_frozen` | Rekening dibekukan | Notifikasi CRITICAL | **CRITICAL** |
| `RekeningUnfrozen` | `events.koperasi.rekening_unfrozen` | Pembekuan dicabut | Notifikasi | Important |
| `RekeningClosed` | `events.koperasi.rekening_closed` | Rekening ditutup | SyncEngine, Notifikasi | **CRITICAL** |
| `DormantFlagged` | `events.koperasi.dormant_flagged` | Rekening dormant | Notifikasi reminder | Best-Effort |

### C3. Transaksi Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `TransaksiCreated` | `events.koperasi.transaksi_created` | Setiap transaksi rekening | Notifikasi receipt, Jurnal (K015) | **CRITICAL** |
| `TransaksiReversed` | `events.koperasi.transaksi_reversed` | Reversal transaksi | Notifikasi, Jurnal reversal | **CRITICAL** |
| `BatchTransactionCompleted` | `events.koperasi.batch_completed` | Batch simpanan wajib selesai | Notifikasi distribusi | Important |

**Payload `TransaksiCreated`:**
```json
{
  "type": "transaksi_created",
  "payload": {
    "transaksi_id": "018f-txn-uuid",
    "tenant_id": "tenant-uuid",
    "branch_id": "018f-branch-uuid",
    "nasabah_id": "018f-nasabah-uuid",
    "rekening_id": "018f-rekening-uuid",
    "transaction_type": "setoran",
    "amount": 500000,
    "balance_after": 1500000,
    "created_by": "018f-teller-uuid",
    "teller_session_id": "018f-session-uuid",
    "created_at": "2026-04-15T10:00:00Z"
  }
}
```

### C4. Pinjaman Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `PinjamanApplicationCreated` | `events.koperasi.pinjaman_applied` | Permohonan pinjaman masuk | Notifikasi Supervisor | Important |
| `PinjamanApproved` | `events.koperasi.pinjaman_approved` | Pinjaman disetujui | Notifikasi, Jadwal angsuran | **CRITICAL** |
| `PinjamanDisbursed` | `events.koperasi.pinjaman_disbursed` | Dana pinjaman dicairkan | Notifikasi, Jurnal | **CRITICAL** |
| `AngsuranDue` | `events.koperasi.angsuran_due` | H-7/H-3/H-1 | Notifikasi reminder | Important |
| `AngsuranOverdue` | `events.koperasi.angsuran_overdue` | DPD > 0 | Notifikasi CRITICAL, NPL update | **CRITICAL** |
| `PinjamanLunas` | `events.koperasi.pinjaman_lunas` | Pinjaman lunas | Notifikasi, Lepas jaminan | Important |

### C5. SHU Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `SHUPeriodeCreated` | `events.koperasi.shu_created` | Periode SHU dibuat | — | Important |
| `SHUDistributed` | `events.koperasi.shu_distributed` | SHU dibagikan ke anggota | Notifikasi, Jurnal | **CRITICAL** |

**Payload `SHUDistributed`:**
```json
{
  "type": "shu_distributed",
  "payload": {
    "shu_periode_id": "018f-shu-uuid",
    "fiscal_year": 2025,
    "total_shu": 125000000,
    "distributed_to_members": 93750000,
    "reserved_amount": 31250000,
    "member_count": 150,
    "distributed_at": "2026-03-15T00:00:00Z"
  }
}
```

### C6. Payroll Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `PayrollDeductionProcessed` | `events.koperasi.payroll_processed` | Potongan gaji berhasil | Notifikasi guru, Sekolah HR | **CRITICAL** |
| `PayrollDeductionFailed` | `events.koperasi.payroll_failed` | Potongan gagal (saldo tidak cukup) | Notifikasi Supervisor | **CRITICAL** |

### C7. Toko / Kantin Events

| Event | Subject | Trigger | Consumer | Kekritisan |
|-------|---------|---------|---------|------------|
| `CanteenPurchase` | `events.koperasi.canteen_purchase` | Transaksi POS kantin | Notifikasi orang tua, spending limit check | Important |
| `SpendingLimitReached` | `events.koperasi.spending_limit_reached` | Limit harian tercapai | Notifikasi CRITICAL ke orang tua | **CRITICAL** |

---

## D. Vernon Sync Events (Internal)

Vernon sync events menggunakan prefix `sync.` dan dihandle oleh SyncEngine Worker.

| Event | Subject | Trigger | Tabel yang Di-sync |
|-------|---------|---------|-------------------|
| Teacher data updated | `sync.class_rooms.teacher_updated` | `TeacherUpdated` | `class_rooms._data.homeroom_teacher` |
| Teacher data updated | `sync.student_grades.teacher_updated` | `TeacherUpdated` | `student_grades._data.teacher` |
| AcademicYear updated | `sync.class_rooms.academic_year_updated` | `AcademicYearUpdated` | `class_rooms._data.academic_year` |
| AcademicYear updated | `sync.students.academic_year_updated` | `AcademicYearUpdated` | `students._data.academic_year` |
| ClassRoom updated | `sync.students.class_room_updated` | `ClassRoomUpdated` | `students._data.class_room` |
| ClassRoom updated | `sync.student_grades.class_room_updated` | `ClassRoomUpdated` | `student_grades._data.class_room` |
| Student updated | `sync.student_academics.student_updated` | `StudentUpdated` | `student_academics._data.student` |
| Student updated | `sync.student_finance_invoices.student_updated` | `StudentUpdated` (CRITICAL) | `student_invoices._data.student` |
| Nasabah updated | `sync.rekening.nasabah_updated` | `NasabahUpdated` | `rekening._data.nasabah` |
| Nasabah updated | `sync.transaksi.nasabah_updated` | `NasabahUpdated` | `transaksi._data.nasabah` |
| Branch updated | `sync.nasabah.branch_updated` | `BranchUpdated` | `nasabah._data.branch` |
| Branch updated | `sync.rekening.branch_updated` | `BranchUpdated` | `rekening._data.branch` |
| Product updated | `sync.rekening.product_updated` | `ProductUpdated` | `rekening._data.product` |

---

## E. Strategi Versioning Event

### Prinsip

Setiap `Event.Type` adalah versi implisit. Perubahan breaking memerlukan event baru:

```
student_activated          ← versi saat ini
student_activated_v2       ← jika payload berubah secara breaking
```

### Aturan Kompatibilitas

| Perubahan | Butuh Versi Baru? |
|-----------|------------------|
| Tambah optional field ke payload | Tidak — consumer harus toleransi field tidak dikenal |
| Ubah tipe field yang ada | **Ya** — breaking change |
| Hapus field dari payload | **Ya** — breaking change |
| Ubah semantik field (nilai berubah) | **Ya** — breaking change |

### Migrasi Versi

```
1. Deploy versi baru: publisher mulai kirim event_v2
2. Consumer lama masih berjalan: subscribe ke event (v1) DAN event_v2
3. Setelah semua consumer upgrade: hapus subscription v1
4. Setelah retensi stream habis (7 hari): v1 events tidak ada lagi
```

---

## Ringkasan: CRITICAL Events

Events yang TIDAK BOLEH hilang — harus menggunakan NATS JetStream dengan:
- Durable consumer (survive restart)
- Ack explicit dari consumer
- Max delivery attempts = 5
- DLQ jika semua attempts gagal
- Alert jika DLQ depth > 0

| Domain | Event | Alasan |
|--------|-------|--------|
| Core | `AcademicYearActivated` | Seluruh sistem bergantung pada tahun ajaran aktif |
| Core | `AcademicYearClosed` | Trigger proses akhir tahun + kenaikan kelas |
| Sekolah | `StudentActivated` | Trigger onboarding Koperasi |
| Sekolah | `ApplicantEnrolled` | Konversi applicant → student (transaksi atomik) |
| Sekolah | `InvoicePaid` | Konfirmasi pembayaran SPP |
| Sekolah | `PaymentConfirmed` | Update status invoice + dompet kantin |
| Sekolah | `InvoiceOverdue` | Trigger collection dan notifikasi |
| Koperasi | `NasabahApproved` | Trigger buka rekening otomatis |
| Koperasi | `NasabahDeactivated` | Trigger penutupan rekening |
| Koperasi | `TransaksiCreated` | Audit trail keuangan immutable |
| Koperasi | `PinjamanDisbursed` | Pencairan dana — tidak boleh terproses dua kali |
| Koperasi | `AngsuranOverdue` | Trigger NPL update dan notifikasi penagihan |
| Koperasi | `SHUDistributed` | Distribusi dana ke rekening anggota |
| Koperasi | `PayrollDeductionProcessed` | Potongan gaji guru harus tercatat |
| Koperasi | `PayrollDeductionFailed` | Kegagalan potongan harus ditangani manual |
| Koperasi | `SpendingLimitReached` | Keamanan finansial anak — orang tua harus tahu |

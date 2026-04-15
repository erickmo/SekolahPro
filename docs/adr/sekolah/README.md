# ADR — Management Sekolah

Keputusan arsitektur yang spesifik untuk subproject **Management Sekolah**. Mencakup domain seperti siswa, guru, kelas, jadwal pelajaran, absensi, nilai, dan lain-lain.

> ADR di sini hanya berlaku untuk subproject ini. Untuk keputusan yang berlaku lintas subproject, lihat [core/](../core/).

## Index

### Siswa (Student Domain)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S001](./ADR-S001-student-core-data-model.md) | Student Core Data Model | Proposed | 2026-04-15 |
| [ADR-S002](./ADR-S002-student-relationships-autoload.md) | Student Relationships & Autoload Strategy | Proposed | 2026-04-15 |
| [ADR-S003](./ADR-S003-student-guardian.md) | Student Guardian (Orang Tua/Wali) | Proposed | 2026-04-15 |
| [ADR-S004](./ADR-S004-student-academic-record.md) | Student Academic Record | Proposed | 2026-04-15 |
| [ADR-S005](./ADR-S005-student-health-record.md) | Student Health Record | Proposed | 2026-04-15 |
| [ADR-S006](./ADR-S006-student-dashboard-menu.md) | Student Dashboard Menu Structure | Proposed | 2026-04-15 |
| [ADR-S007](./ADR-S007-student-address-previous-school.md) | Student Address & Previous School | Proposed | 2026-04-15 |
| [ADR-S008](./ADR-S008-daily-attendance.md) | Daily Attendance Transaction | Approved | 2026-04-15 |
| [ADR-S009](./ADR-S009-student-finance-spp.md) | Student Finance / SPP | Proposed | 2026-04-15 |
| [ADR-S010](./ADR-S010-student-document-management.md) | Student Document Management | Proposed | 2026-04-15 |
| [ADR-S011](./ADR-S011-subject-grade-detail.md) | Subject Grade Detail | Proposed | 2026-04-15 |
| [ADR-S012](./ADR-S012-student-discipline.md) | Student Discipline | Proposed | 2026-04-15 |
| [ADR-S013](./ADR-S013-student-achievement.md) | Student Achievement | Proposed | 2026-04-15 |
| [ADR-S014](./ADR-S014-student-class-placement.md) | Student Class Placement | Proposed | 2026-04-15 |
| [ADR-S015](./ADR-S015-student-extracurricular.md) | Student Extracurricular | Proposed | 2026-04-15 |
| [ADR-S016](./ADR-S016-student-admission-ppdb.md) | Student Admission / PPDB | Proposed | 2026-04-15 |
| [ADR-S017](./ADR-S017-student-counseling-bk.md) | Student Counseling / BK | Proposed | 2026-04-15 |
| [ADR-S018](./ADR-S018-rapor-generation.md) | Rapor Generation | Proposed | 2026-04-15 |

### Kurikulum & Akademik (Curriculum & Academic)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S019](./ADR-S019-curriculum-management.md) | Curriculum Management (Manajemen Kurikulum) | Proposed | 2026-04-15 |
| [ADR-S020](./ADR-S020-subject-management.md) | Subject Management (Mata Pelajaran) | Proposed | 2026-04-15 |
| [ADR-S021](./ADR-S021-teaching-schedule.md) | Teaching Schedule / Timetable (Jadwal Pelajaran) | Proposed | 2026-04-15 |
| [ADR-S022](./ADR-S022-exam-assessment.md) | Exam & Assessment Management (Ujian & Penilaian) | Proposed | 2026-04-15 |
| [ADR-S023](./ADR-S023-academic-calendar.md) | Academic Calendar (Kalender Akademik) | Proposed | 2026-04-15 |
| [ADR-S024](./ADR-S024-lesson-plan-rpp.md) | Lesson Plan / RPP (Rencana Pelaksanaan Pembelajaran) | Proposed | 2026-04-15 |
| [ADR-S025](./ADR-S025-teaching-journal.md) | Teaching Journal (Jurnal Mengajar) | Proposed | 2026-04-15 |

### Guru & Staff (Teacher & Staff HR)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S026](./ADR-S026-teacher-attendance.md) | Teacher Attendance (Absensi Guru & Staff) | Proposed | 2026-04-15 |
| [ADR-S027](./ADR-S027-teacher-workload.md) | Teacher Workload (Beban Mengajar) | Proposed | 2026-04-15 |
| [ADR-S028](./ADR-S028-teacher-performance-evaluation.md) | Teacher Performance Evaluation (PKG) | Proposed | 2026-04-15 |
| [ADR-S029](./ADR-S029-professional-development.md) | Professional Development (PKB) | Proposed | 2026-04-15 |
| [ADR-S030](./ADR-S030-leave-management.md) | Leave Management (Cuti & Izin Guru/Staff) | Proposed | 2026-04-15 |
| [ADR-S031](./ADR-S031-staff-payroll.md) | Staff Payroll (Penggajian) | Proposed | 2026-04-15 |
| [ADR-S032](./ADR-S032-teacher-substitution.md) | Teacher Schedule & Substitution (Piket & Penggantian) | Proposed | 2026-04-15 |

### Asrama / Dormitory (Boarding School)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S033](./ADR-S033-dormitory-management.md) | Dormitory Management (Kamar & Penghuni) | Proposed | 2026-04-15 |
| [ADR-S034](./ADR-S034-dormitory-activity-attendance.md) | Dormitory Activity & Attendance | Proposed | 2026-04-15 |
| [ADR-S035](./ADR-S035-dormitory-discipline-health.md) | Dormitory Discipline & Health | Proposed | 2026-04-15 |

### Kantin / Canteen

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S036](./ADR-S036-canteen-management.md) | Canteen Management (Menu & Vendor) | Proposed | 2026-04-15 |
| [ADR-S037](./ADR-S037-canteen-transaction-billing.md) | Canteen Transaction & Billing | Proposed | 2026-04-15 |

### Fasilitas & Sarana (Facilities & Infrastructure)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S038](./ADR-S038-library-management.md) | Library Management (Perpustakaan) | Proposed | 2026-04-15 |
| [ADR-S039](./ADR-S039-laboratory-management.md) | Laboratory Management (Laboratorium) | Proposed | 2026-04-15 |
| [ADR-S040](./ADR-S040-asset-inventory.md) | Asset & Inventory Management (Aset & Inventaris) | Proposed | 2026-04-15 |
| [ADR-S041](./ADR-S041-room-facility-booking.md) | Room & Facility Booking (Peminjaman Ruangan) | Proposed | 2026-04-15 |

### Komunikasi & Portal (Communication & Portal)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S042](./ADR-S042-parent-portal-dashboard.md) | Parent Portal & Dashboard | Proposed | 2026-04-15 |
| [ADR-S043](./ADR-S043-communication-messaging.md) | Communication & Messaging | Proposed | 2026-04-15 |
| [ADR-S044](./ADR-S044-notification-system.md) | Notification System (Multi-Channel) | Proposed | 2026-04-15 |
| [ADR-S045](./ADR-S045-announcement-news.md) | Announcement & News (Pengumuman) | Proposed | 2026-04-15 |

### Administrasi & Tata Usaha (Administration)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S046](./ADR-S046-correspondence-management.md) | Correspondence Management (Surat Menyurat TU) | Proposed | 2026-04-15 |
| [ADR-S047](./ADR-S047-approval-workflow.md) | Approval Workflow Engine (Mesin Persetujuan) | Proposed | 2026-04-15 |
| [ADR-S048](./ADR-S048-school-profile-accreditation.md) | School Profile & Accreditation (Profil & Akreditasi) | Proposed | 2026-04-15 |
| [ADR-S049](./ADR-S049-school-committee.md) | School Committee (Komite Sekolah) | Proposed | 2026-04-15 |

### Keuangan Sekolah (School Finance)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S050](./ADR-S050-school-budget-rkas.md) | School Budget / RKAS | Proposed | 2026-04-15 |
| [ADR-S051](./ADR-S051-payment-gateway.md) | Payment Gateway (Gateway Pembayaran) | Proposed | 2026-04-15 |

### Transportasi & Layanan (Transportation & Services)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S052](./ADR-S052-transportation-management.md) | Transportation Management (Antar Jemput) | Proposed | 2026-04-15 |

### Integrasi & Pelaporan (Integration & Reporting)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S053](./ADR-S053-elearning-integration.md) | E-Learning Integration | Proposed | 2026-04-15 |
| [ADR-S054](./ADR-S054-reporting-analytics.md) | Reporting & Analytics Dashboard | Proposed | 2026-04-15 |
| [ADR-S055](./ADR-S055-dapodik-integration.md) | Dapodik Integration | Proposed | 2026-04-15 |

### Alumni & Beasiswa (Alumni & Scholarship)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S056](./ADR-S056-alumni-management.md) | Alumni Management | Proposed | 2026-04-15 |
| [ADR-S057](./ADR-S057-scholarship-management.md) | Scholarship Management (Beasiswa) | Proposed | 2026-04-15 |

### Kegiatan Sekolah (School Events)

| No. | Judul | Status | Tanggal |
|-----|-------|--------|---------|
| [ADR-S058](./ADR-S058-school-event-management.md) | School Event Management (Kegiatan Sekolah) | Proposed | 2026-04-15 |

---

## Stakeholder Coverage Matrix

| Stakeholder | ADRs yang Melayani |
|-------------|-------------------|
| **Siswa** | S001-S018 (core student data, academic, finance, health, extracurricular) |
| **Guru** | S019-S025 (kurikulum, jadwal, lesson plan, journal), S026-S032 (HR, payroll, kinerja) |
| **Kepala Sekolah** | S028 (PKG), S047 (approval), S048 (profil & akreditasi), S054 (analytics) |
| **Wakasek Kurikulum** | S019-S025 (kurikulum, jadwal, RPP approval) |
| **Wakasek Kesiswaan** | S012 (disiplin), S015 (ekskul), S017 (BK), S058 (kegiatan) |
| **Wakasek Sarana** | S038-S041 (perpustakaan, lab, aset, booking fasilitas) |
| **Tata Usaha (TU)** | S046 (surat menyurat), S010 (dokumen siswa), S048 (profil sekolah) |
| **Bendahara** | S009 (SPP), S050 (RKAS), S051 (payment gateway), S031 (payroll) |
| **Orang Tua / Wali** | S003 (guardian), S042 (parent portal), S043 (messaging), S044 (notifikasi) |
| **Asrama / Pesantren** | S033-S035 (dormitory management, activity, discipline) |
| **Kantin** | S036-S037 (canteen management, transaction & billing) |
| **Pustakawan** | S038 (library management) |
| **Laboran** | S039 (laboratory management) |
| **Komite Sekolah** | S049 (committee management), S050 (budget approval) |
| **Alumni** | S056 (alumni management) |
| **Dinas Pendidikan** | S055 (Dapodik integration), S054 (reporting) |

## Domain Count

- **Total ADR**: 58 (S001 — S058)
- **Tabel Database**: ~100+ tabel (beberapa ADR memiliki 2-4 tabel)
- **Stakeholder Tercakup**: 16 peran

# 04 — School Platform Spec

**Status**: Draft  
**Perspective**: System Analyst + System Engineer  
**Scope**: Management Sekolah

## Purpose

Mendefinisikan capability, actor, dependency, rollout, dan acceptance baseline untuk seluruh domain sekolah berdasarkan ADR-S001 sampai ADR-S058.

## Primary Actors

- admin sekolah
- operator akademik
- guru
- wali kelas
- kepala sekolah
- wakasek
- tata usaha
- bendahara
- orang tua/wali
- siswa
- pustakawan/laboran/pengelola fasilitas
- dinas / sistem eksternal

## Capability Domains

### 1. Student Lifecycle
**ADRs**: S001, S002, S003, S007, S010, S014, S016  
**Goal**: satu sumber data siswa dari intake sampai penempatan kelas.

Key processes:
- registrasi dan intake siswa
- pengelolaan biodata, wali, alamat, dokumen
- placement ke kelas/rombel
- PPDB dan transisi ke siswa aktif

### 2. Academic Operations
**ADRs**: S019, S020, S021, S023, S024, S025  
**Goal**: perencanaan dan pelaksanaan kegiatan belajar mengajar.

Key processes:
- setup kurikulum
- setup mata pelajaran
- kalender akademik
- penyusunan jadwal
- lesson plan / RPP
- jurnal mengajar

### 3. Attendance & Assessment
**ADRs**: S004, S008, S011, S018, S022  
**Goal**: menangkap aktivitas belajar harian sampai keluaran rapor.

Key processes:
- absensi harian siswa
- pencatatan performa akademik
- penilaian per mapel
- ujian dan assessment
- rapor generation

### 4. Student Services
**ADRs**: S005, S009, S012, S013, S015, S017  
**Goal**: mengelola layanan pendukung siswa dan intervensi sekolah.

Key processes:
- kesehatan siswa
- pelanggaran/disiplin
- prestasi
- ekstrakurikuler
- BK
- SPP dan status pembayaran

### 5. HR & School Operations
**ADRs**: S026-S032  
**Goal**: mengelola operasional guru dan staff.

Key processes:
- absensi guru/staff
- beban mengajar
- cuti/izin
- pengganti jadwal
- PKG/PKB
- payroll

### 6. Facilities & Services
**ADRs**: S033-S041, S052  
**Goal**: mengelola fasilitas, layanan sekolah, dan boarding operations.

### 7. Communication & Administration
**ADRs**: S042-S049  
**Goal**: membangun control plane komunikasi dan administrasi sekolah.

### 8. Finance, Integration & Growth
**ADRs**: S050-S058  
**Goal**: memperluas kapabilitas sekolah ke budgeting, pembayaran, integrasi, dan analytics.

## Rollout by Wave

### Wave A — School Core MVP
- auth + tenant + RBAC
- academic years, teachers/staff, class rooms
- student lifecycle
- curriculum and schedules
- attendance and grades
- rapor baseline
- SPP baseline

### Wave B — School Operations Expansion
- HR operations
- communication/admin
- facilities dasar
- approval workflow

### Wave C — School Premium / Integration
- Dapodik
- e-learning
- analytics
- scholarship, alumni, events
- advanced payments / RKAS
- optional boarding/canteen/service modules

## Dependency Rules

- Student lifecycle bergantung pada core foundation shared
- Academic operations bergantung pada academic year, teacher, class, dan subject setup
- Attendance & assessment bergantung pada class placement dan schedule
- Rapor adalah orchestration capability di atas attendance, grades, discipline, dan academic record
- Integrations hanya boleh aktif setelah master data dan governance stabil

## Acceptance Baseline

- `AC-FUNC`: sekolah dapat menjalankan proses akademik utama end-to-end
- `AC-AUTH`: actor hanya melihat data sesuai tenant, peran, dan scope
- `AC-DATA`: status siswa, kelas, nilai, dan pembayaran konsisten
- `AC-AUDIT`: approval dan perubahan data penting tercatat
- `AC-INT`: integrasi eksternal punya retry, reconciliation, dan failure handling
- `AC-NFR`: modul inti sekolah memenuhi target respons, availability, dan operability

# Sprint Ordering — Outside-In Dependency Map

> **Prinsip**: Setiap fase hanya membutuhkan ADR dari fase sebelumnya.
> Ikuti urutan ini untuk sprint planning — fase atas = fondasi, fase bawah = fitur lanjutan.
> Dalam satu fase, ADR bisa dikerjakan **paralel**.

---

## Legend

| Symbol | Arti |
|--------|------|
| `[C]` | Core (shared foundation) |
| `[S]` | Sekolah domain |
| `[K]` | Koperasi domain |
| `→` | Depends on |
| `★` | Critical path — harus selesai sebelum fase berikutnya |
| ✅ | Done |
| ❌ | Not done |
| N/A | Not applicable (infra/architecture) |

---

## FASE 0 — Platform Foundation (Sprint 0)

> Arsitektur, pattern, dan infrastruktur dasar. **Tidak ada fitur bisnis.**
> Semua ADR di fase ini TIDAK punya dependency ke ADR lain.

| # | ADR | Judul | Paralel? | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|----------|----------|-----------|------------|------------------|
| 1 | `[C] ADR-001` | Go Clean Architecture + CQRS ★ | Ya | ✅ | N/A | N/A | N/A |
| 2 | `[C] ADR-002` | Vernon Denormalized Read-Cache ★ | Ya | ✅ | N/A | N/A | N/A |
| 3 | `[C] ADR-003` | Hybrid CQRS + Vernon Decision Criteria ★ | Ya | ✅ | N/A | N/A | N/A |
| 4 | `[C] ADR-005` | Event Bus (InMemory / NATS JetStream) ★ | Ya | ✅ | N/A | N/A | N/A |
| 5 | `[C] ADR-006` | Uber FX Dependency Injection ★ | Ya | ✅ | N/A | N/A | N/A |
| 6 | `[C] ADR-007` | UUID v7 Primary Key ★ | Ya | ✅ | N/A | N/A | N/A |
| 7 | `[C] ADR-008` | React 18 + Vite + CSS Modules | Ya | ✅ | N/A | N/A | N/A |
| 8 | `[C] ADR-009` | Dual-Mode Institution Type (General / Islamic) ★ | Ya | ✅ | N/A | N/A | N/A |

**Deliverable**: Boilerplate project skeleton, CI pipeline, dev environment.

---

## FASE 1 — Auth & Multi-Tenant (Sprint 1)

> Tanpa auth dan tenant isolation, tidak ada domain yang bisa berjalan.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 9 | `[C] ADR-004` | Multi-Tenant 4-Level Hierarchy + JWT ★ | Fase 0 | ✅ | ✅ | ❌ | ✅ |
| 10 | `[C] ADR-013` | Users & Roles (Pengguna & Hak Akses) ★ | ADR-004 | ✅ | ✅ | ❌ | ❌ |

**Deliverable**: Login, tenant switching, RBAC middleware, user CRUD.

---

## FASE 2 — Master Data Sekolah (Sprint 2)

> Foundation domains yang dibutuhkan oleh SEMUA domain Sekolah.
> ADR-010 dan ADR-012 paralel, ADR-011 menunggu keduanya.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 11 | `[C] ADR-010` | Academic Years (Tahun Ajaran) ★ | Fase 1 | ✅ | ✅ | ✅ | ✅ |
| 12 | `[C] ADR-012` | Teachers & Staff (Guru & Tenaga Kependidikan) ★ | Fase 1 | ✅ | ✅ | ✅ | ✅ |
| 13 | `[C] ADR-011` | Class Rooms (Kelas & Rombel) ★ | ADR-010, ADR-012 | ✅ | ✅ | ✅ | ✅ |

**Deliverable**: CRUD tahun ajaran, guru/staff, kelas. Semester active toggle.

---

## FASE 3 — Core Koperasi (Sprint 2, paralel dengan Fase 2)

> Foundation domains Koperasi — bisa dikerjakan **paralel** dengan Fase 2.
> K003 harus selesai sebelum K004-K007.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 14 | `[K] ADR-K003` | Produk & Akad ★ | Fase 1 (ADR-009) | ✅ | ✅ | ❌ | ❌ |
| 15 | `[K] ADR-K001` | Nasabah (Anggota/Member) ★ | Fase 1 (ADR-004) | ✅ | ✅ | ❌ | ❌ |
| 16 | `[K] ADR-K002` | Rekening ★ | K001, K003 | ✅ | ✅ | ❌ | ❌ |

**Deliverable**: Katalog produk, registrasi nasabah, pembukaan rekening.

---

## FASE 4A — Student Core (Sprint 3)

> Data model siswa — fondasi seluruh domain Sekolah.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 17 | `[S] ADR-S001` | Student Core Data Model ★ | Fase 2 (ADR-010, 011) | ✅ | ✅ | ❌ | ❌ |
| 18 | `[S] ADR-S007` | Student Address & Previous School | S001 | ✅ | ✅ | ❌ | ❌ |
| 19 | `[S] ADR-S003` | Student Guardian (Orang Tua/Wali) ★ | S001 | ✅ | ✅ | ❌ | ❌ |
| 20 | `[S] ADR-S002` | Student Relationships & Autoload | S001 | ✅ | ✅ | ❌ | ❌ |
| 21 | `[S] ADR-S010` | Student Document Management | S001 | ✅ | ✅ | ❌ | ❌ |
| 22 | `[S] ADR-S014` | Student Class Placement ★ | S001, ADR-011 | ✅ | ✅ | ❌ | ❌ |
| 23 | `[S] ADR-S016` | Student Admission / PPDB | S001, S003 | ✅ | ✅ | ❌ | ❌ |

**Deliverable**: CRUD siswa, wali, placement ke kelas, PPDB workflow.

---

## FASE 4B — Simpanan Koperasi (Sprint 3, paralel dengan Fase 4A)

> Produk simpanan — membutuhkan Fase 3 selesai.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 24 | `[K] ADR-K004` | Simpanan Pokok & Wajib ★ | K002, K003 | ✅ | ❌ | ❌ | ❌ |
| 25 | `[K] ADR-K005` | Tabungan | K002, K003 | ✅ | ❌ | ❌ | ❌ |
| 26 | `[K] ADR-K006` | Deposito / Simpanan Berjangka | K002, K003 | ✅ | ❌ | ❌ | ❌ |

**Deliverable**: Setoran pokok/wajib, tabungan multi-produk, deposito + rollover.

---

## FASE 5A — Kurikulum & Akademik (Sprint 4)

> Inti kegiatan belajar mengajar. S019 adalah fondasi akademik.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 27 | `[S] ADR-S019` | Curriculum Management ★ | ADR-010 | ✅ | ❌ | ❌ | ❌ |
| 28 | `[S] ADR-S020` | Subject Management (Mata Pelajaran) ★ | S019 | ✅ | ❌ | ❌ | ❌ |
| 29 | `[S] ADR-S023` | Academic Calendar | ADR-010 | ✅ | ❌ | ❌ | ❌ |
| 30 | `[S] ADR-S021` | Teaching Schedule / Timetable ★ | S020, ADR-011, ADR-012 | ✅ | ❌ | ❌ | ❌ |
| 31 | `[S] ADR-S024` | Lesson Plan / RPP | S020, ADR-012 | ✅ | ❌ | ❌ | ❌ |
| 32 | `[S] ADR-S025` | Teaching Journal | S021, ADR-012 | ✅ | ❌ | ❌ | ❌ |

**Deliverable**: Kurikulum, mata pelajaran, jadwal, RPP, jurnal mengajar.

---

## FASE 5B — Pembiayaan Koperasi (Sprint 4, paralel dengan Fase 5A)

> Pinjaman dan pembiayaan — membutuhkan Fase 3 selesai.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 33 | `[K] ADR-K007` | Pinjaman / Pembiayaan ★ | K002, K003 | ✅ | ❌ | ❌ | ❌ |
| 34 | `[K] ADR-K008` | Angsuran & Jadwal | K007 | ✅ | ❌ | ❌ | ❌ |
| 35 | `[K] ADR-K009` | Denda & Penalti | K007, K008 | ❌ | ❌ | ❌ | ❌ |
| 36 | `[K] ADR-K010` | Jaminan / Agunan | K007 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Pengajuan pinjaman, jadwal angsuran, denda, jaminan.

---

## FASE 6A — Kehadiran & Penilaian (Sprint 5)

> Daily operations sekolah — absensi, nilai, rapor.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 37 | `[S] ADR-S008` | Daily Attendance (Absensi Harian) ★ | S001, S014, ADR-011 | ❌ | ❌ | ❌ | ❌ |
| 38 | `[S] ADR-S004` | Student Academic Record ★ | S001 | ❌ | ❌ | ❌ | ❌ |
| 39 | `[S] ADR-S011` | Subject Grade Detail ★ | S001, S020, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 40 | `[S] ADR-S022` | Exam & Assessment Management | S011, S020, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 41 | `[S] ADR-S005` | Student Health Record | S001 | ❌ | ❌ | ❌ | ❌ |
| 42 | `[S] ADR-S012` | Student Discipline | S001, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 43 | `[S] ADR-S013` | Student Achievement | S001 | ❌ | ❌ | ❌ | ❌ |
| 44 | `[S] ADR-S015` | Student Extracurricular | S001, ADR-012 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Absensi harian, input nilai, ujian, catatan akademik, kesehatan, disiplin, ekskul.

---

## FASE 6B — Transaksi & Operasional Koperasi (Sprint 5, paralel dengan Fase 6A)

> Engine transaksi keuangan koperasi.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 45 | `[K] ADR-K011` | Transaksi Rekening & Non-Rekening ★ | K002, K007 | ❌ | ❌ | ❌ | ❌ |
| 46 | `[K] ADR-K012` | Teller Session | K011 | ❌ | ❌ | ❌ | ❌ |
| 47 | `[K] ADR-K013` | Money Denomination | K012 | ❌ | ❌ | ❌ | ❌ |
| 48 | `[K] ADR-K014` | Kas & Cash Flow | K011, K012 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Transaction engine, teller workflow, kas harian, denominasi.

---

## FASE 7A — Rapor & HR Guru (Sprint 6)

> Rapor = orchestration layer. HR guru = fitur pendukung.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 49 | `[S] ADR-S018` | Rapor Generation ★ | S004, S008, S011, S012, S014, S015 | ❌ | ❌ | ❌ | ❌ |
| 50 | `[S] ADR-S009` | Student Finance / SPP | S001, ADR-010 | ❌ | ❌ | ❌ | ❌ |
| 51 | `[S] ADR-S026` | Teacher Attendance | ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 52 | `[S] ADR-S027` | Teacher Workload | S021, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 53 | `[S] ADR-S030` | Leave Management (Cuti Guru/Staff) | ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 54 | `[S] ADR-S032` | Teacher Substitution (Piket) | S021, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 55 | `[S] ADR-S017` | Student Counseling / BK | S001, S012, ADR-012 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Cetak rapor, SPP, absensi guru, beban mengajar, cuti, piket, BK.

---

## FASE 7B — Akuntansi Koperasi (Sprint 6, paralel dengan Fase 7A)

> Double-entry accounting dan pelaporan keuangan.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 56 | `[K] ADR-K015` | Jurnal & Akuntansi (COA) ★ | K011 | ❌ | ❌ | ❌ | ❌ |
| 57 | `[K] ADR-K016` | SHU (Sisa Hasil Usaha) | K015 | ❌ | ❌ | ❌ | ❌ |
| 58 | `[K] ADR-K017` | Laporan Regulasi (OJK, Dinas) | K015, K016 | ❌ | ❌ | ❌ | ❌ |
| 59 | `[K] ADR-K018` | Zakat & Infaq (BMT mode) | K015, ADR-009 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: COA, jurnal otomatis, SHU, laporan OJK/Dinas, zakat.

---

## FASE 8 — Fasilitas & Sarana (Sprint 7)

> Perpustakaan, lab, aset, booking — tidak blocking fitur lain.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 60 | `[S] ADR-S038` | Library Management (Perpustakaan) | S001 | ❌ | ❌ | ❌ | ❌ |
| 61 | `[S] ADR-S039` | Laboratory Management | ADR-011 | ❌ | ❌ | ❌ | ❌ |
| 62 | `[S] ADR-S040` | Asset & Inventory Management | Fase 1 | ❌ | ❌ | ❌ | ❌ |
| 63 | `[S] ADR-S041` | Room & Facility Booking | ADR-011 | ❌ | ❌ | ❌ | ❌ |
| 64 | `[S] ADR-S028` | Teacher Performance Evaluation (PKG) | ADR-012, S027 | ❌ | ❌ | ❌ | ❌ |
| 65 | `[S] ADR-S029` | Professional Development (PKB) | ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 66 | `[S] ADR-S031` | Staff Payroll (Penggajian) | ADR-012, S030 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Perpustakaan, lab, inventaris, booking ruangan, PKG, PKB, payroll guru.

---

## FASE 9 — Asrama & Kantin (Sprint 8)

> Boarding school features — opsional, hanya untuk pesantren/asrama.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 67 | `[S] ADR-S033` | Dormitory Management (Kamar & Penghuni) | S001, ADR-011 | ❌ | ❌ | ❌ | ❌ |
| 68 | `[S] ADR-S034` | Dormitory Activity & Attendance | S033 | ❌ | ❌ | ❌ | ❌ |
| 69 | `[S] ADR-S035` | Dormitory Discipline & Health | S033, S012 | ❌ | ❌ | ❌ | ❌ |
| 70 | `[S] ADR-S036` | Canteen Management (Menu & Vendor) | Fase 1 | ❌ | ❌ | ❌ | ❌ |
| 71 | `[S] ADR-S037` | Canteen Transaction & Billing | S036, S001 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Manajemen asrama, absensi asrama, kantin, POS kantin.

---

## FASE 10 — Stakeholder Extensions Koperasi (Sprint 8, paralel dengan Fase 9)

> Fitur pendukung koperasi: toko, e-wallet, payroll deduction.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 72 | `[K] ADR-K019` | Koperasi Toko & Kantin | K011 | ❌ | ❌ | ❌ | ❌ |
| 73 | `[K] ADR-K020` | Payroll Integration & Auto-Deduction | K004, K008, ADR-012 | ❌ | ❌ | ❌ | ❌ |
| 74 | `[K] ADR-K021` | E-Wallet / Uang Saku Digital | K002, K011 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: POS toko koperasi, auto-deduction gaji, e-wallet siswa.

---

## FASE 11 — Komunikasi & Portal (Sprint 9)

> Portal orang tua, notifikasi, dan komunikasi.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 75 | `[S] ADR-S044` | Notification System (Multi-Channel) ★ | Fase 1 (infrastruktur) | ❌ | ❌ | ❌ | ❌ |
| 76 | `[S] ADR-S043` | Communication & Messaging | S044, ADR-013 | ❌ | ❌ | ❌ | ❌ |
| 77 | `[S] ADR-S045` | Announcement & News | S044 | ❌ | ❌ | ❌ | ❌ |
| 78 | `[S] ADR-S042` | Parent Portal & Dashboard | S001, S003, S008, S011, S018 | ❌ | ❌ | ❌ | ❌ |
| 79 | `[S] ADR-S006` | Student Dashboard Menu Structure | S001 | ❌ | ❌ | ❌ | ❌ |
| 80 | `[K] ADR-K022` | Notifikasi & Komunikasi Koperasi | K011, S044 | ❌ | ❌ | ❌ | ❌ |
| 81 | `[K] ADR-K023` | Dashboard & Self-Service Portal | K001, K002, K011, K015 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Notifikasi multi-channel, messaging, portal orang tua, dashboard koperasi.

---

## FASE 12 — Administrasi & Keuangan Sekolah (Sprint 10)

> Tata usaha, surat menyurat, keuangan sekolah.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 82 | `[S] ADR-S046` | Correspondence Management (Surat TU) | ADR-013 | ❌ | ❌ | ❌ | ❌ |
| 83 | `[S] ADR-S047` | Approval Workflow Engine | ADR-013 | ❌ | ❌ | ❌ | ❌ |
| 84 | `[S] ADR-S048` | School Profile & Accreditation | Fase 1 | ❌ | ❌ | ❌ | ❌ |
| 85 | `[S] ADR-S049` | School Committee (Komite Sekolah) | S048 | ❌ | ❌ | ❌ | ❌ |
| 86 | `[S] ADR-S050` | School Budget / RKAS | ADR-010, S047 | ❌ | ❌ | ❌ | ❌ |
| 87 | `[S] ADR-S051` | Payment Gateway | S009 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Surat-menyurat, approval flow, profil sekolah, RKAS, payment gateway.

---

## FASE 13 — Integrasi & Pelaporan (Sprint 11)

> Integrasi ke sistem eksternal — membutuhkan data lengkap.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 88 | `[S] ADR-S054` | Reporting & Analytics Dashboard | Semua domain sekolah | ❌ | ❌ | ❌ | ❌ |
| 89 | `[S] ADR-S055` | Dapodik Integration | S001, ADR-012, S048 | ❌ | ❌ | ❌ | ❌ |
| 90 | `[S] ADR-S053` | E-Learning Integration | S019, S020, S021 | ❌ | ❌ | ❌ | ❌ |
| 91 | `[S] ADR-S052` | Transportation Management | S001 | ❌ | ❌ | ❌ | ❌ |
| 92 | `[K] ADR-K024` | Integrasi & API Koperasi | K011, K015, K017 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Dashboard analytics, Dapodik sync, e-learning, transportasi, API koperasi.

---

## FASE 14 — Nice-to-Have & Future (Sprint 12+)

> Fitur tambahan — bisa ditunda tanpa menghambat MVP.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 93 | `[S] ADR-S056` | Alumni Management | S001 | ❌ | ❌ | ❌ | ❌ |
| 94 | `[S] ADR-S057` | Scholarship Management (Beasiswa) | S001, S009 | ❌ | ❌ | ❌ | ❌ |
| 95 | `[S] ADR-S058` | School Event Management | ADR-013 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Alumni tracking, beasiswa, manajemen kegiatan.

---

## FASE 15 — Koperasi Governance & Compliance (Sprint 6-7, paralel)

> Tata kelola, kepatuhan regulasi, dan perlindungan data — CRITICAL untuk koperasi yang beroperasi sebagai LKM/BMT.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 96 | `[K] ADR-K025` | Cooperative Governance & Internal Controls ★ | K001, ADR-013 | ❌ | ❌ | ❌ | ❌ |
| 97 | `[K] ADR-K026` | RAT (Rapat Anggota Tahunan) Management ★ | K025, K016 | ❌ | ❌ | ❌ | ❌ |
| 98 | `[K] ADR-K027` | AML/CFT Compliance ★ | K001, K011 | ✅ | ❌ | ❌ | ❌ |
| 99 | `[K] ADR-K028` | Data Privacy (UU PDP) ★ | K001 | ✅ | ❌ | ❌ | ❌ |
| 100 | `[K] ADR-K029` | Business Continuity & Disaster Recovery ★ | K011, K014 | ✅ | ❌ | ❌ | ❌ |

**Deliverable**: Struktur pengurus, RAT system, AML monitoring, consent management, BCP/DRP.

---

## FASE 16 — Koperasi Operational Excellence (Sprint 8-9, paralel)

> Penyempurnaan operasional koperasi — lifecycle, risiko, keamanan, collection.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 101 | `[K] ADR-K030` | Membership Lifecycle Management | K001, K004, K007 | ❌ | ❌ | ❌ | ❌ |
| 102 | `[K] ADR-K031` | Cooperative Health Indicators & Risk Management | K015, K017 | ❌ | ❌ | ❌ | ❌ |
| 103 | `[K] ADR-K032` | Biometric Authentication | K028, ADR-004 | ✅ | ❌ | ❌ | ❌ |
| 104 | `[K] ADR-K033` | Loan Collection Management | K007, K008, K009 | ✅ | ❌ | ❌ | ❌ |
| 105 | `[K] ADR-K034` | Reserve Fund Management | K016 | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Lifecycle anggota, health dashboard, biometric auth, collection system, reserve fund.

---

## FASE 17 — Koperasi Growth & Innovation (Sprint 10-11, paralel)

> Fitu pertumbuhan — asuransi, mobile app, pendidikan anggota.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 106 | `[K] ADR-K035` | Insurance / Takaful Integration | K007 | ❌ | ❌ | ❌ | ❌ |
| 107 | `[K] ADR-K036` | Mobile App Strategy | K021, K022 | ❌ | ❌ | ❌ | ❌ |
| 108 | `[K] ADR-K037` | Member Education Program | K001 | ✅ | ❌ | ❌ | ❌ |

**Deliverable**: Asuransi kredit/takaful, Flutter mobile app, program pendidikan anggota.

---

## FASE 18 — Koperasi Strategic (Sprint 12+)

> Fitur strategis jangka panjang — dissolution, konsolidasi, roadmap digital.

| # | ADR | Judul | Depends On | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|------------|----------|-----------|------------|------------------|
| 109 | `[K] ADR-K038` | Cooperative Dissolution Process | K025, K026 | ✅ | ❌ | ❌ | ❌ |
| 110 | `[K] ADR-K039` | Multi-Branch Consolidation Reporting | K015, K017 | ✅ | ❌ | ❌ | ❌ |
| 111 | `[K] ADR-K040` | Digital Transformation Roadmap | All K-ADRs | ❌ | ❌ | ❌ | ❌ |

**Deliverable**: Proses pembubaran, laporan konsolidasian, roadmap transformasi digital.

---

## FASE INF — Infrastructure & Ops (Ongoing, mulai Sprint 0)

> ADR infrastruktur — dikerjakan bertahap sepanjang project.

| # | ADR | Judul | Kapan | Is Coded | Test Code | API Tested | Dashboard Tested |
|---|-----|-------|-------|----------|-----------|------------|------------------|
| — | `[C] ADR-014` | Vernon Sync Engine Strategy | Sprint 0-2 | ❌ | N/A | N/A | N/A |
| — | `[C] ADR-015` | Frontend Architecture Multi-App | Sprint 0-1 | ✅ | N/A | N/A | N/A |
| — | `[C] ADR-016` | Deployment & CI/CD Pipeline | Sprint 0 | ❌ | N/A | N/A | N/A |
| — | `[C] ADR-017` | Data Migration & Import Strategy | Sprint 8+ | ❌ | N/A | N/A | N/A |
| — | `[C] ADR-018` | Regulatory Compliance | Ongoing | ❌ | N/A | N/A | N/A |
| — | `[C] ADR-BIZ-001` | Business Model & Pricing | Pre-Sprint 0 | ✅ | N/A | N/A | N/A |

---

## Visual: Critical Path

```
Sprint 0    Sprint 1      Sprint 2         Sprint 3           Sprint 4
────────    ────────      ────────         ────────           ────────
ADR-001 ──→ ADR-004 ──→ ADR-010 ──┐      S001 ──→ S014     S019 ──→ S020 ──→ S021
ADR-002     ADR-013      ADR-012 ──┼──→   S003               S023
ADR-003                  ADR-011 ──┘      S016
ADR-005
ADR-006     ║             ║ (paralel)
ADR-007     ║            K003 ──→ K002    K004               K007 ──→ K008
ADR-008     ║            K001 ──┘         K005                        K009
ADR-009                                   K006                        K010

Sprint 5           Sprint 6           Sprint 7        Sprint 8
────────           ────────           ────────        ────────
S008, S011 ──→     S018 (RAPOR) ★     S038-S041      S033-S035
S022, S004         S009, S026          S028, S029     S036-S037
S012, S015         S027, S030          S031           K019, K020
                   S017, S032                          K021
K011 ──→ K012      K015 ──→ K016
          K013              K017
          K014              K018

Sprint 9           Sprint 10          Sprint 11       Sprint 12+
────────           ──────────         ────────        ──────────
S044 ──→ S043      S046, S047         S054            S056, S057
         S045      S048 ──→ S049      S055            S058
S042               S050, S051         S053, S052
S006               K024 (API)
K022, K023
```

---

## Ringkasan Per Sprint

| Sprint | Fase | ADR Count | Focus | Coded | Tested |
|--------|------|-----------|-------|-------|--------|
| 0 | Fase 0 + INF | 8 + 6 | Platform foundation, CI/CD, boilerplate | 8/8 + 2/6 | N/A |
| 1 | Fase 1 | 2 | Auth, multi-tenant, RBAC | 2/2 | 1/2 |
| 2 | Fase 2 + 3 | 3 + 3 = 6 | Master data sekolah + core koperasi | 6/6 | 6/6 |
| 3 | Fase 4A + 4B | 7 + 3 = 10 | Student core + simpanan koperasi | 10/10 | 7/10 |
| 4 | Fase 5A + 5B | 6 + 4 = 10 | Kurikulum + pembiayaan koperasi | 4/10 | 0/10 |
| 5 | Fase 6A + 6B | 8 + 4 = 12 | Kehadiran, nilai + transaksi koperasi | 0/12 | 0/12 |
| 6 | Fase 7A + 7B + 15 | 7 + 4 + 5 = 16 | Rapor, HR guru + akuntansi + governance/compliance | 3/16 | 0/16 |
| 7 | Fase 8 + 15 cont. | 7 | Fasilitas, sarana, PKG, payroll | 0/7 | 0/7 |
| 8 | Fase 9 + 10 + 16 | 5 + 3 + 5 = 13 | Asrama, kantin + ext. koperasi + operational excellence | 2/13 | 0/13 |
| 9 | Fase 11 + 16 cont. | 7 | Portal, notifikasi, dashboard | 0/7 | 0/7 |
| 10 | Fase 12 + 17 | 6 + 3 = 9 | Administrasi, keuangan + insurance, mobile, education | 1/9 | 0/9 |
| 11 | Fase 13 | 5 | Integrasi, reporting, analytics | 0/5 | 0/5 |
| 12+ | Fase 14 + 18 | 3 + 3 = 6 | Alumni, beasiswa, event + dissolution, consolidation, roadmap | 2/6 | 0/6 |
| — | Fase INF | 6 | Infrastructure (ongoing) | 2/6 | N/A |

---

## Progress Summary

| Metric | Count | Percentage |
|--------|-------|------------|
| **Total ADR** | 111 | — |
| **Is Coded** | 40/111 | 36% |
| **Test Code** | 17/105 | 16% |
| **API Tested** | 3/105 | 3% |
| **Dashboard Tested** | 4/105 | 4% |

> Catatan: 6 ADR infrastruktur (Fase INF) dikecualikan dari Test/API/Dashboard karena N/A.
> Test Code: 13 domain punya unit test (descriptor_test.go), 1 punya integration test (example).

---

*Generated: 2026-04-15 — Berdasarkan dependency analysis dari semua ADR.*
*Updated: 2026-04-16 — Added implementation tracking columns (Is Coded, Test Code, API Tested, Dashboard Tested).*
*Prinsip: Outside-In = fondasi dulu, fitur user-facing bertahap, integrasi terakhir.*

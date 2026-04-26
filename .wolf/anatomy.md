# anatomy.md

> Auto-maintained by OpenWolf. Last scanned: 2026-04-26T11:00:00.659Z
> Files: 508 tracked | Anatomy hits: 0 | Misses: 0

## ./

- `.gitignore` — Git ignore rules (~103 tok)
- `CLAUDE.md` — OpenWolf (~1246 tok)
- `README.md` — Project documentation (~1562 tok)

## .claude/

- `settings.json` (~492 tok)
- `settings.local.json` (~458 tok)

## .claude/rules/

- `openwolf.md` (~313 tok)

## .claude/worktrees/agent-a04f6200/

- `.gitignore` — Git ignore rules (~103 tok)
- `README.md` — Project documentation (~1562 tok)

## .claude/worktrees/agent-a04f6200/.claude/

- `settings.local.json` (~243 tok)

## .claude/worktrees/agent-a04f6200/api/

- `.air.toml` (~258 tok)
- `.gitignore` — Git ignore rules (~11 tok)
- `CLAUDE.md` — Go Boilerplate API (~2230 tok)
- `docker-compose.yml` — Docker Compose services (~365 tok)
- `Dockerfile` — Docker container definition (~155 tok)
- `go.mod` — Go module definition (~1472 tok)
- `go.sum` — Go dependency checksums (~7288 tok)
- `Makefile` — Make build targets (~654 tok)
- `prometheus.yml` (~51 tok)
- `sqlc.yaml` (~92 tok)

## .claude/worktrees/agent-a04f6200/api/cmd/api/

- `main.go` — @title           Boilerplate API (~3184 tok)
- `server.go` (~1156 tok)

## .claude/worktrees/agent-a04f6200/api/docs/

- `.gitkeep` (~0 tok)

## .claude/worktrees/agent-a04f6200/api/infrastructure/cache/

- `cache.go` — menyediakan abstraksi caching dan implementasi Redis. (~227 tok)
- `redis_cache.go` — RedisCache (10 fields); methods: Get, Set, Delete (~432 tok)

## .claude/worktrees/agent-a04f6200/api/infrastructure/config/

- `config.go` — Config (77 fields) (~2310 tok)

## .claude/worktrees/agent-a04f6200/api/infrastructure/database/

- `db.go` — NewDB (~160 tok)
- `dissolution_repository.go` — DissolutionRepository (63 fields); methods: Save, Update, Delete, GetByID (~2626 tok)
- `example_repository.go` — ExampleRepository (59 fields); methods: Save, Update, Delete, GetByID (~1611 tok)

## .claude/worktrees/agent-a04f6200/api/infrastructure/telemetry/

- `telemetry.go` — menginisialisasi OpenTelemetry SDK dengan: (~912 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/cancel_dissolution/

- `handler.go` — menangani command untuk membatalkan proses pembubaran. (~504 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/complete_dissolution/

- `handler.go` — menangani command untuk menyelesaikan proses pembubaran. (~778 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/create_dissolution/

- `handler.go` — menangani command untuk membuat proses pembubaran koperasi baru. (~953 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/create_example/

- `handler.go` — menangani command untuk membuat Example baru. (~609 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/delete_dissolution/

- `handler.go` — menangani command untuk soft-delete proses pembubaran. (~313 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/delete_example/

- `handler.go` — menangani command untuk soft-delete Example. (~300 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/update_dissolution_stage/

- `handler.go` — menangani command untuk mengubah tahapan proses pembubaran. (~723 tok)

## .claude/worktrees/agent-a04f6200/api/internal/command/update_example/

- `handler.go` — menangani command untuk mengupdate Example yang sudah ada. (~576 tok)

## .claude/worktrees/agent-a04f6200/api/internal/delivery/http/

- `dissolution_handler.go` — DissolutionHandler (57 fields); methods: RegisterRoutes, List, GetByID, Create (~3217 tok)
- `example_handler.go` — ExampleHandler (18 fields); methods: RegisterRoutes, List, GetByID, Create (~1507 tok)
- `response.go` — HTTP handlers: respondJSON, respondError (~536 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/academic_year/

- `descriptor_test.go` — TestDescriptor_TableName, TestDescriptor_DefaultRels_Empty, TestValidate_RejectsEmptyName, TestValidate_RejectsEmptyCode + 4 more (~772 tok)
- `descriptor.go` — adalah domain Vernon untuk tahun ajaran (ADR-010). (~1116 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/class_room/

- `descriptor_test.go` — TestDescriptor_TableName, TestDescriptor_DefaultRels_Returns2, TestDescriptor_Rels_AcademicYearAutoload, TestValidate_RejectsEmptyName + 6 more (~1032 tok)
- `descriptor.go` — adalah domain Vernon untuk kelas (ADR-011). (~1174 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/dissolution/

- `entity.go` — mendefinisikan domain untuk proses pembubaran koperasi. (~1363 tok)
- `errors.go` (~352 tok)
- `events.go` — DissolutionInitiatedEvent dipublikasikan saat proses pembubaran baru berhasil dibuat. (~522 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/example/

- `entity.go` — adalah contoh domain module. (~478 tok)
- `errors.go` (~65 tok)
- `events.go` — ExampleCreatedEvent dipublikasikan saat Example baru berhasil dibuat. (~278 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/product/

- `descriptor.go` — adalah contoh domain Vernon — tipe: belongs_to dengan autoload. (~758 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/product_category/

- `descriptor.go` — adalah contoh domain Vernon — tipe: standalone (tanpa autoload). (~616 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/role/

- `entity_test.go` — TestHasPermission_Exact, TestHasPermission_Missing, TestHasPermission_Wildcard, TestHasPermission_Empty (~271 tok)
- `entity.go` — adalah domain CQRS untuk roles dan permissions (ADR-013). (~1144 tok)
- `errors.go` — Declares tidak (~184 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/teacher/

- `descriptor_test.go` — TestDescriptor_TableName, TestDescriptor_DefaultRels_Empty, TestValidate_RejectsEmptyFullName, TestValidate_RejectsInvalidRole + 4 more (~751 tok)
- `descriptor.go` — adalah domain Vernon untuk guru dan tenaga kependidikan (ADR-012). (~1162 tok)

## .claude/worktrees/agent-a04f6200/api/internal/domain/user/

- `entity.go` — adalah domain CQRS untuk users dan authentication (ADR-013). (~634 tok)
- `errors.go` (~159 tok)

## .claude/worktrees/agent-a04f6200/api/internal/eventhandler/

- `example_handler.go` — berisi event handler lintas domain. (~729 tok)

## .claude/worktrees/agent-a04f6200/api/internal/query/get_dissolution_by_id/

- `handler.go` — menangani query untuk mengambil proses pembubaran berdasarkan ID. (~1084 tok)

## .claude/worktrees/agent-a04f6200/api/internal/query/get_example_by_id/

- `handler.go` — menangani query untuk mengambil Example berdasarkan ID. (~490 tok)

## .claude/worktrees/agent-a04f6200/api/internal/query/list_dissolutions/

- `handler.go` — menangani query untuk mengambil daftar proses pembubaran. (~712 tok)

## .claude/worktrees/agent-a04f6200/api/internal/query/list_examples/

- `handler.go` — menangani query untuk mengambil daftar Example. (~554 tok)

## .claude/worktrees/agent-a04f6200/api/migrations/

- `001_create_examples.sql` — +migrate Up (~384 tok)
- `002_vernon_base.sql` — +migrate Up (~861 tok)
- `003_vernon_product_categories.sql` — +migrate Up (~462 tok)
- `004_vernon_products.sql` — +migrate Up (~414 tok)
- `005_academic_years.sql` — +migrate Up (~580 tok)
- `006_teachers.sql` — +migrate Up (~902 tok)
- `007_class_rooms.sql` — +migrate Up (~573 tok)
- `008_users_roles.sql` — +migrate Up (~962 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/commandbus/

- `commandbus.go` — Interface: Command (3 methods) (~918 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/eventbus/

- `eventbus.go` — Interface: DomainEvent (11 methods) (~936 tok)
- `nats_eventbus.go` — NATSConfig (26 fields); methods: Publish, Subscribe, StartRouter, StopRouter (~852 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/jwt/

- `jwt.go` — menyediakan utility untuk generate dan validasi JWT token. (~1459 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/middleware/

- `auth_middleware.go` — menyediakan HTTP middleware untuk API. (~801 tok)
- `middleware.go` — Tracer, Logger, Recoverer, RealIP (~904 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/pagination/

- `pagination.go` — ListParams (9 fields) (~247 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/querybus/

- `querybus.go` — Interface: Query (3 methods) (~618 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/scope/

- `scope.go` — menyediakan tipe dan helper untuk hierarki organisasi 4 level: (~2434 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/tenant/

- `tenant.go` — menyediakan abstraksi untuk multi-tenant dan single-tenant. (~1553 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/vernon/

- `domain.go` — menyediakan Vernon denormalized read-cache architecture. (~1320 tok)
- `errors.go` (~150 tok)
- `handler.go` — HTTP handlers: respondJSON, respondError, respondList, scopeOrError, parseUUID (~1936 tok)
- `query_parser.go` — ParseQueryParams, BuildWhereClause, BuildOrderClause (~933 tok)
- `registry.go` — ConsumerRef (33 fields); methods: Register, GetDomain, GetConsumers, MountRoutes (~1228 tok)
- `repository.go` — Interface: scanner (3 methods) (~2357 tok)
- `service.go` — BaseService (68 fields); methods: Create, Update, Patch, Delete (~2022 tok)

## .claude/worktrees/agent-a04f6200/api/pkg/vernonsync/

- `engine.go` — menyediakan SyncEngine untuk Vernon denormalized read-cache. (~2223 tok)

## .claude/worktrees/agent-a04f6200/api/scripts/

- `migrate.sh` — migrate.sh — wrapper untuk golang-migrate (~196 tok)

## .claude/worktrees/agent-a04f6200/api/seeds/

- `001_example.sql` — Seed data untuk domain Example. (~430 tok)

## .claude/worktrees/agent-a04f6200/api/sqlc/queries/

- `examples.sql` — tenant_id dan company_id selalu menjadi filter pertama di semua query (~344 tok)

## .claude/worktrees/agent-a04f6200/api/sqlc/schema/

- `examples.sql` — Schema untuk sqlc code generation. (~175 tok)

## .claude/worktrees/agent-a04f6200/api/tests/integration/

- `example_test.go` — berisi integration tests yang membutuhkan PostgreSQL container. (~1924 tok)

## .claude/worktrees/agent-a04f6200/docs/.obsidian/

- `app.json` (~1 tok)
- `appearance.json` (~1 tok)
- `core-plugins.json` (~199 tok)

## .claude/worktrees/agent-a04f6200/docs/adr/

- `README.md` — Project documentation (~795 tok)

## .claude/worktrees/agent-a04f6200/docs/adr/core/

- `ADR-001-go-clean-architecture-cqrs.md` — ADR-001: Go Clean Architecture + CQRS (~1609 tok)
- `ADR-002-vernon-denormalized-read-cache.md` — ADR-002: Vernon Denormalized Read-Cache Pattern (_rels/_data JSONB) (~2152 tok)
- `ADR-003-hybrid-cqrs-vernon.md` — ADR-003: Hybrid CQRS + Vernon — Decision Criteria (~1682 tok)
- `ADR-004-multi-tenant-4level-hierarchy.md` — ADR-004: Multi-Tenant 4-Level Hierarchy + Two-Phase JWT (~2222 tok)
- `ADR-005-event-bus-inmemory-nats.md` — ADR-005: Event Bus Abstraction — InMemory (Dev) / NATS JetStream (Prod) (~2461 tok)
- `ADR-006-uber-fx-dependency-injection.md` — ADR-006: Uber FX sebagai Dependency Injection Container (~2515 tok)
- `ADR-007-uuid-v7-primary-key.md` — ADR-007: UUID v7 sebagai Primary Key (~2417 tok)
- `ADR-008-react-vite-css-modules.md` — ADR-008: React 18 + Vite + CSS Modules (Tanpa Tailwind / UI Library) (~3011 tok)
- `ADR-009-dual-mode-institution-type.md` — ADR-009: Dual-Mode Institution Type (General / Islamic) (~1685 tok)
- `ADR-010-academic-years.md` — ADR-010: Academic Years (Tahun Ajaran) (~2185 tok)
- `ADR-011-class-rooms.md` — ADR-011: Class Rooms (Kelas) (~2359 tok)
- `ADR-012-teachers-staff.md` — ADR-012: Teachers & Staff (Guru dan Tenaga Kependidikan) (~2453 tok)
- `ADR-013-users-roles.md` — ADR-013: Users & Roles (Authentication & Authorization) (~3099 tok)
- `ADR-014-vernon-sync-engine-strategy.md` — ADR-014: Vernon Sync Engine Strategy (~1852 tok)
- `ADR-015-frontend-architecture-multi-app.md` — ADR-015: Frontend Architecture — Multi-App Strategy (~1871 tok)
- `ADR-016-deployment-cicd-pipeline.md` — ADR-016: Deployment & CI/CD Pipeline (~1593 tok)
- `ADR-017-data-migration-import-strategy.md` — ADR-017: Strategi Migrasi Data & Import (~2146 tok)
- `ADR-018-regulatory-compliance.md` — ADR-018: Regulatory Compliance Matrix (~1797 tok)
- `ADR-BIZ-001-business-model-pricing.md` — ADR-BIZ-001: Business Model & Pricing Strategy (~1972 tok)
- `README.md` — Project documentation (~753 tok)

## .claude/worktrees/agent-a04f6200/docs/adr/koperasi/

- `ADR-K001-nasabah.md` — ADR-K001: Nasabah (Anggota / Member) (~5652 tok)
- `ADR-K002-rekening.md` — ADR-K002: Rekening (Akun Nasabah) (~6575 tok)
- `ADR-K003-produk-akad.md` — ADR-K003: Produk & Akad (~6580 tok)
- `ADR-K004-simpanan-pokok-wajib.md` — ADR-K004: Simpanan Pokok & Wajib (~6435 tok)
- `ADR-K005-tabungan.md` — ADR-K005: Tabungan (Simpanan Sukarela) (~7229 tok)
- `ADR-K006-deposito.md` — ADR-K006: Deposito / Simpanan Berjangka (~8020 tok)
- `ADR-K007-pinjaman.md` — ADR-K007: Pinjaman / Pembiayaan (~8745 tok)
- `ADR-K008-angsuran-jadwal.md` — ADR-K008: Angsuran & Jadwal (~7728 tok)
- `ADR-K009-denda-penalti.md` — ADR-K009: Denda & Penalti (~5556 tok)
- `ADR-K010-jaminan-agunan.md` — ADR-K010: Jaminan / Agunan (Collateral Management) (~8767 tok)
- `ADR-K011-transaksi.md` — ADR-K011: Transaksi Rekening & Non-Rekening (~7728 tok)
- `ADR-K012-teller-session.md` — ADR-K012: Teller Session (~5535 tok)
- `ADR-K013-money-denomination.md` — ADR-K013: Money Denomination (Pecahan Uang) (~4078 tok)
- `ADR-K014-kas-cashflow.md` — ADR-K014: Kas & Cash Flow (~5854 tok)
- `ADR-K015-jurnal-coa.md` — ADR-K015: Jurnal & Akuntansi (COA) (~8506 tok)
- `ADR-K016-shu.md` — ADR-K016: SHU (Sisa Hasil Usaha) (~6679 tok)
- `ADR-K017-laporan-regulasi.md` — ADR-K017: Laporan Regulasi (Regulatory Reporting) (~6496 tok)
- `ADR-K018-zakat-infaq.md` — ADR-K018: Zakat & Infaq (~9326 tok)
- `ADR-K019-toko-kantin.md` — ADR-K019: Koperasi Toko & Kantin (~6408 tok)
- `ADR-K020-payroll-deduction.md` — ADR-K020: Payroll Integration & Auto-Deduction (~4638 tok)
- `ADR-K021-ewallet.md` — ADR-K021: Uang Saku Digital (Kartu Belanja Siswa) (~5418 tok)
- `ADR-K022-notifikasi.md` — ADR-K022: Notifikasi & Komunikasi (~4668 tok)
- `ADR-K023-dashboard-portal.md` — ADR-K023: Dashboard & Self-Service Portal (~6741 tok)
- `ADR-K024-integrasi-api.md` — ADR-K024: Integrasi & API (Integration & API Gateway) (~8718 tok)
- `README.md` — Project documentation (~1941 tok)

## .claude/worktrees/agent-a04f6200/docs/adr/koperasi/reviews/

- `review-K003-K006.md` — C-Suite Advisory Review: ADR K003-K006 (~7096 tok)
- `review-K007-K010.md` — C-Suite Advisory Review: ADR K007-K010 (~7322 tok)
- `review-K011-K014.md` — C-Suite Advisory Review: ADR K011-K014 (~6972 tok)
- `review-K015-K018.md` — C-Suite Advisory Review: ADR K015-K018 (~7020 tok)
- `review-K019-K024.md` — C-Suite Advisory Review: ADR K019-K024 (Extension Modules) (~8284 tok)

## .claude/worktrees/agent-a04f6200/docs/adr/sekolah/

- `ADR-S001-student-core-data-model.md` — ADR-S001: Student Core Data Model (~2366 tok)
- `ADR-S002-student-relationships-autoload.md` — ADR-S002: Student Relationships & Autoload Strategy (~1915 tok)
- `ADR-S003-student-guardian.md` — ADR-S003: Student Guardian (Orang Tua/Wali) (~2308 tok)
- `ADR-S004-student-academic-record.md` — ADR-S004: Student Academic Record (~2149 tok)
- `ADR-S005-student-health-record.md` — ADR-S005: Student Health Record (~2099 tok)
- `ADR-S006-student-dashboard-menu.md` — ADR-S006: Student Dashboard Menu Structure (~2528 tok)
- `ADR-S007-student-address-previous-school.md` — ADR-S007: Student Address & Previous School (~2497 tok)
- `ADR-S008-daily-attendance.md` — ADR-S008: Daily Attendance Transaction (~2919 tok)
- `ADR-S009-student-finance-spp.md` — ADR-S009: Student Finance / SPP (~3939 tok)
- `ADR-S010-student-document-management.md` — ADR-S010: Student Document Management (~2038 tok)
- `ADR-S011-subject-grade-detail.md` — ADR-S011: Subject Grade Detail (~3064 tok)
- `ADR-S012-student-discipline.md` — ADR-S012: Student Discipline / Tata Tertib (~2919 tok)
- `ADR-S013-student-achievement.md` — ADR-S013: Student Achievement / Prestasi (~1563 tok)
- `ADR-S014-student-class-placement.md` — ADR-S014: Student Class Placement / Mutasi Kelas (~2334 tok)
- `ADR-S015-student-extracurricular.md` — ADR-S015: Student Extracurricular (~3242 tok)
- `ADR-S016-student-admission-ppdb.md` — ADR-S016: PPDB / Student Admission (~3084 tok)
- `ADR-S017-student-counseling-bk.md` — ADR-S017: BK / Student Counseling Record (~2425 tok)
- `ADR-S018-rapor-generation.md` — ADR-S018: Rapor Generation (~4098 tok)
- `ADR-S019-curriculum-management.md` — ADR-S019: Curriculum Management (Manajemen Kurikulum) (~4403 tok)
- `ADR-S020-subject-management.md` — ADR-S020: Subject Management (Manajemen Mata Pelajaran) (~3582 tok)
- `ADR-S021-teaching-schedule.md` — ADR-S021: Teaching Schedule / Timetable (Jadwal Pelajaran) (~4234 tok)
- `ADR-S022-exam-assessment.md` — ADR-S022: Exam & Assessment Management (Ujian & Penilaian) (~4533 tok)
- `ADR-S023-academic-calendar.md` — ADR-S023: Academic Calendar (Kalender Akademik) (~3892 tok)
- `ADR-S024-lesson-plan-rpp.md` — ADR-S024: Lesson Plan / RPP (Rencana Pelaksanaan Pembelajaran) (~5334 tok)
- `ADR-S025-teaching-journal.md` — ADR-S025: Teaching Journal (Jurnal Mengajar) (~5593 tok)
- `ADR-S026-teacher-attendance.md` — ADR-S026: Teacher Attendance (Absensi Guru & Staff) (~3553 tok)
- `ADR-S027-teacher-workload.md` — ADR-S027: Teacher Workload (Beban Mengajar) (~3618 tok)
- `ADR-S028-teacher-performance-evaluation.md` — ADR-S028: Teacher Performance Evaluation (Penilaian Kinerja Guru - PKG) (~4300 tok)
- `ADR-S029-professional-development.md` — ADR-S029: Professional Development (Pengembangan Profesi Berkelanjutan - PKB) (~4493 tok)
- `ADR-S030-leave-management.md` — ADR-S030: Leave Management (Manajemen Cuti & Izin Guru/Staff) (~6279 tok)
- `ADR-S031-staff-payroll.md` — ADR-S031: Staff Payroll (Penggajian Guru & Staff) (~7228 tok)
- `ADR-S032-teacher-substitution.md` — ADR-S032: Teacher Schedule & Substitution (Jadwal Piket & Penggantian Guru) (~6982 tok)
- `ADR-S033-dormitory-management.md` — ADR-S033: Dormitory Management / Manajemen Asrama (Kamar & Penghuni) (~3689 tok)
- `ADR-S034-dormitory-activity-attendance.md` — ADR-S034: Dormitory Activity & Attendance / Aktivitas & Kehadiran Asrama (~3931 tok)
- `ADR-S035-dormitory-discipline-health.md` — ADR-S035: Dormitory Discipline & Health / Disiplin & Kesehatan Asrama (~4657 tok)
- `ADR-S036-canteen-management.md` — ADR-S036: Canteen Management / Manajemen Kantin (~3571 tok)
- `ADR-S037-canteen-transaction-billing.md` — ADR-S037: Canteen Transaction & Billing / Transaksi & Tagihan Kantin (~4459 tok)
- `ADR-S038-library-management.md` — ADR-S038: Library Management / Perpustakaan (~4193 tok)
- `ADR-S039-laboratory-management.md` — ADR-S039: Laboratory Management (Manajemen Laboratorium) (~5640 tok)
- `ADR-S040-asset-inventory.md` — ADR-S040: Asset & Inventory Management (Manajemen Aset & Inventaris Sekolah) (~5583 tok)
- `ADR-S041-room-facility-booking.md` — ADR-S041: Room & Facility Booking (Peminjaman Ruangan & Fasilitas) (~6181 tok)
- `ADR-S042-parent-portal-dashboard.md` — ADR-S042: Parent Portal & Dashboard (Portal Orang Tua) (~3424 tok)
- `ADR-S043-communication-messaging.md` — ADR-S043: Communication & Messaging (Komunikasi & Pesan) (~4276 tok)
- `ADR-S044-notification-system.md` — ADR-S044: Notification System (Sistem Notifikasi Multi-Channel) (~4443 tok)
- `ADR-S045-announcement-news.md` — ADR-S045: Announcement & News (Pengumuman & Berita Sekolah) (~3969 tok)
- `ADR-S046-correspondence-management.md` — ADR-S046: Correspondence Management (Manajemen Surat Menyurat TU) (~5572 tok)
- `ADR-S047-approval-workflow.md` — ADR-S047: Approval Workflow Engine (Mesin Persetujuan Universal) (~6825 tok)
- `ADR-S048-school-profile-accreditation.md` — ADR-S048: School Profile & Accreditation (Profil Sekolah & Akreditasi) (~4845 tok)
- `ADR-S049-school-committee.md` — ADR-S049: School Committee (Komite Sekolah) (~5575 tok)
- `ADR-S050-school-budget-rkas.md` — ADR-S050: School Budget Management (RKAS) (~4214 tok)
- `ADR-S051-payment-gateway.md` — ADR-S051: School Payment Gateway (Gateway Pembayaran) (~5350 tok)
- `ADR-S052-transportation-management.md` — ADR-S052: Transportation Management (Manajemen Transportasi / Antar Jemput) (~4335 tok)
- `ADR-S053-elearning-integration.md` — ADR-S053: E-Learning Integration (Integrasi Pembelajaran Daring) (~4761 tok)
- `ADR-S054-reporting-analytics.md` — ADR-S054: Reporting & Analytics Dashboard (Laporan & Analitik) (~5595 tok)
- `ADR-S055-dapodik-integration.md` — ADR-S055: Dapodik Integration (Integrasi Dapodik) (~5598 tok)
- `ADR-S056-alumni-management.md` — ADR-S056: Alumni Management (Manajemen Alumni) (~5305 tok)
- `ADR-S057-scholarship-management.md` — ADR-S057: Scholarship Management (Manajemen Beasiswa) (~5704 tok)
- `ADR-S058-school-event-management.md` — ADR-S058: School Event Management (Manajemen Kegiatan Sekolah) (~5792 tok)
- `README.md` — Project documentation (~2442 tok)

## .claude/worktrees/agent-a04f6200/docs/developer/

- `01-getting-started.md` — 01 — Getting Started (~2316 tok)
- `02-architecture-overview.md` — 02 — Architecture Overview (~3262 tok)
- `03-adding-cqrs-domain.md` — 03 — Adding a CQRS Domain (~6976 tok)
- `04-adding-vernon-domain.md` — 04 — Adding a Vernon Domain (~4593 tok)
- `05-multi-tenant-guide.md` — 05 — Multi-Tenant Guide (~3140 tok)
- `06-testing-guide.md` — 06 — Testing Guide (~4493 tok)
- `07-environment-configuration.md` — 07 — Environment Configuration (~3032 tok)
- `README.md` — Project documentation (~524 tok)

## .claude/worktrees/agent-a04f6200/docs/user-manual/

- `01-introduction.md` — 01 — Introduction (~2078 tok)
- `02-deployment-guide.md` — 02 — Deployment Guide (~2723 tok)
- `03-configuration-reference.md` — 03 — Configuration Reference (~1895 tok)
- `04-api-usage-guide.md` — 04 — API Usage Guide (~3337 tok)
- `05-monitoring-observability.md` — 05 — Monitoring & Observability (~2521 tok)
- `06-web-dashboard-guide.md` — 06 — Web Dashboard Guide (~2283 tok)
- `README.md` — Project documentation (~478 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/

- `.gitignore` — Git ignore rules (~84 tok)
- `CLAUDE.md` — web-dashboard — Boilerplate (~1029 tok)
- `eslint.config.js` — ESLint flat configuration (~176 tok)
- `index.html` — Dashboard (~193 tok)
- `package-lock.json` — npm lock file (~56638 tok)
- `package.json` — Node.js package manifest (~419 tok)
- `tsconfig.app.json` (~230 tok)
- `tsconfig.json` — TypeScript configuration (~34 tok)
- `tsconfig.node.json` (~187 tok)
- `vite.config.ts` — Vite build configuration (~156 tok)
- `vitest.config.ts` — Vitest test configuration (~172 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/

- `main.tsx` (~95 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/__ui_tests__/

- `setup.ts` (~60 tok)
- `test-utils.tsx` — createTestQueryClient (~344 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/__ui_tests__/mocks/

- `handlers.ts` — API routes: POST, GET (3 endpoints) (~217 tok)
- `server.ts` — Exports server (~36 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/app/

- `App.tsx` — ThemeInitializer — uses useEffect (~144 tok)
- `ProtectedRoute.tsx` — Only platform superusers can access /su/* routes (~1034 tok)
- `providers.tsx` — queryClient (~155 tok)
- `routes.tsx` — LoginPage (~1025 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/config/

- `app.config.ts` — When true, enables multi-tenant routing: (~199 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/contexts/

- `FlowInfoContext.tsx` — FlowInfoProvider — uses useCallback (~197 tok)
- `useFlowInfo.ts` — Exports FlowInfoContextValue, FlowInfoContext, useFlowInfo (~166 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/hooks/

- `useAutoBreadcrumbs.ts` — Generates breadcrumbs from the current pathname. (~763 tok)
- `useBreakpoint.ts` — Exports useBreakpoint (~211 tok)
- `useClickOutside.ts` — Fires `handler` when a click occurs outside all provided refs. (~282 tok)
- `useClipboard.ts` — Exports useClipboard (~312 tok)
- `useCompanyPath.ts` — Returns a helper to build paths scoped to the current company context. (~126 tok)
- `useCountdown.ts` — Exports useCountdown (~530 tok)
- `useDashboardContext.ts` — Hook untuk mendeteksi konteks dashboard saat ini (HQ atau Company). (~242 tok)
- `useDataSource.ts` — Exports useDataSource (~613 tok)
- `useDebounce.ts` — Exports useDebounce (~90 tok)
- `useEventListener.ts` — Overload 1: window target (element omitted or undefined) (~539 tok)
- `useForm.ts` — Exports useForm (~854 tok)
- `useHQPath.ts` — Returns a helper to build paths scoped to the HQ/Group context. (~73 tok)
- `useIntersectionObserver.ts` — Fire only once, then disconnect (~421 tok)
- `useInterval.ts` — Exports useInterval (~178 tok)
- `useKeyboard.ts` — Prevent default browser action (~485 tok)
- `useLocalStorage.ts` — Like useState but persists to localStorage. (~312 tok)
- `useModuleAccess.ts` — HQ: transaksi hanya bisa dilihat (~377 tok)
- `usePermission.ts` — Exports usePermission (~228 tok)
- `usePrevious.ts` — Returns the value from the previous render. (~92 tok)
- `useSuperuserPath.ts` — Returns a helper to build paths scoped to the Superuser context. (~78 tok)
- `useTheme.ts` — Exports useTheme (~81 tok)
- `useTimeout.ts` — Exports useTimeout (~232 tok)
- `useToggle.ts` — Exports useToggle (~150 tok)
- `useWindowSize.ts` — Exports useWindowSize (~245 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/layouts/AppNavbar/

- `AppNavbar.module.css` — Styles: 59 rules, 2 media queries, 1 animations (~2631 tok)
- `AppNavbar.tsx` — NAV_ITEMS_DEFAULT — uses useNavigate, useState, useEffect (~2619 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/layouts/AppShell/

- `AppShell.module.css` — Styles: 4 rules, 2 media queries (~116 tok)
- `AppShell.tsx` — Multi-tenant context — controls navbar color and nav items. (~164 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/layouts/PageHeader/

- `PageHeader.module.css` — Styles: 12 rules, 1 media queries (~347 tok)
- `PageHeader.tsx` — PageHeader (~373 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/AuditLog/

- `AuditLogPage.module.css` — Styles: 47 rules, 1 animations (~2028 tok)
- `AuditLogPage.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~2426 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/ChangePassword/

- `ChangePasswordPage.module.css` — Styles: 36 rules, 1 animations (~1413 tok)
- `ChangePasswordPage.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~2612 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/ChooseCompany/

- `ChooseCompanyPage.module.css` — Styles: 22 rules, 1 animations (~1048 tok)
- `ChooseCompanyPage.tsx` — ChooseCompanyPage — uses useNavigate (~811 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/Dashboard/

- `DashboardPage.module.css` — Styles: 15 rules, 3 media queries (~581 tok)
- `DashboardPage.tsx` — ─── Stat card component ────────────────────────────────────────────────────── (~625 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/Login/

- `LoginPage.module.css` — Styles: 28 rules, 2 animations (~1256 tok)
- `LoginPage.tsx` — LoginPage — renders form — uses useNavigate, useState (~1227 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/Profile/

- `ProfilePage.module.css` — Styles: 37 rules, 1 media queries, 1 animations (~1649 tok)
- `ProfilePage.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~1735 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/pages/errors/

- `ErrorPage.module.css` — Styles: 6 rules (~337 tok)
- `ForbiddenPage.tsx` — ForbiddenPage (~142 tok)
- `NotFoundPage.tsx` — NotFoundPage (~146 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/services/

- `api.client.ts` — Exports apiClient, apiClientWithToken (~982 tok)
- `audit-log.service.ts` — ─── Paginated response (inline — not yet in api.types) ─────────────────────── (~165 tok)
- `auth.service.ts` — API routes: GET (1 endpoints) (~121 tok)
- `chat.service.ts` — ─── Types ────────────────────────────────────────────────────────────────────── (~2331 tok)
- `company-group.service.ts` — Exports companyGroupService (~95 tok)
- `createEntityService.ts` — Exports ListParams, createEntityService (~486 tok)
- `dashboard.service.ts` — Exports DashboardSummary, DashboardActivity, dashboardService (~494 tok)
- `media.service.ts` — Media service untuk menangani file upload ke endpoint /api/v1/media/upload (~792 tok)
- `query-keys.ts` — QK — Centralized Query Key constants. (~324 tok)
- `superuser-company-group.service.ts` — Exports superuserCompanyGroupService (~101 tok)
- `tenant-owner.service.ts` — Exports tenantOwnerService (~96 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/stores/

- `auth.store.ts` — ─── State ──────────────────────────────────────────────────────────────────── (~802 tok)
- `chat.store.ts` — Exports useChatStore (~2081 tok)
- `notification.store.ts` — Exports NotificationItem, useNotificationStore (~393 tok)
- `ui.store.ts` — Exports useUiStore (~409 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/theme/

- `motion.css` — Styles: 15 rules, 1 media queries, 10 animations (~441 tok)
- `reset.css` — Styles: 1 rules, 2 vars, 1 media queries (~372 tok)
- `typography.css` — Styles: 1 rules (~218 tok)
- `variables.css` — Styles: 102 vars (~1132 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/types/

- `api.types.ts` — Exports PaginatedResponse, AppApiError, RequestConfig (~142 tok)
- `audit-log.types.ts` — Exports AuditAction, AuditLog, AuditLogFilters (~176 tok)
- `auth.types.ts` — ─── User roles ─────────────────────────────────────────────────────────────── (~463 tok)
- `entity.types.ts` — Shared base types for all entities. (~96 tok)
- `navigation.types.ts` — Exports NavItem, SubNavItem, DropdownNavItem, Breadcrumb (~175 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/utils/

- `cn.ts` — Exports cn (~36 tok)
- `export.ts` — Exports ExportFormat, ExportColumn, exportToCSV, exportToJSON, exportToPDF (~750 tok)
- `format.ts` — Exports formatCurrency, formatNumber, formatPercent, formatDate + 4 more (~644 tok)
- `storage.ts` — Typed localStorage / sessionStorage wrapper. (~1015 tok)
- `validation.ts` — Reusable field validators for use with `useForm`. (~943 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Accordion/

- `Accordion.module.css` — Styles: 21 rules (~913 tok)
- `Accordion.tsx` — Accordion (~771 tok)
- `index.ts` (~12 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ActionMenu/

- `ActionMenu.module.css` — Styles: 12 rules, 1 animations (~628 tok)
- `ActionMenu.tsx` — Custom trigger — defaults to MoreHorizontal (3-dot) button (~936 tok)
- `index.ts` (~27 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Avatar/

- `Avatar.module.css` — Styles: 14 rules (~416 tok)
- `Avatar.tsx` — SIZE_PX — uses useState (~686 tok)
- `index.ts` (~26 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Badge/

- `Badge.module.css` — Styles: 13 rules (~501 tok)
- `Badge.tsx` — Maps a status string to a variant using a provided lookup map (~384 tok)
- `index.ts` (~29 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Breadcrumb/

- `Breadcrumb.module.css` — Styles: 8 rules (~268 tok)
- `Breadcrumb.tsx` — Breadcrumb (~518 tok)
- `index.ts` (~27 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ChartCard/

- `ChartCard.module.css` — Styles: 7 rules (~222 tok)
- `ChartCard.tsx` — ChartCard — renders chart (~207 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ChartWidget/

- `ChartWidget.module.css` — Styles: 3 rules, 1 animations (~207 tok)
- `ChartWidget.tsx` — Display name in legend and tooltip (~2865 tok)
- `index.ts` (~39 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ChatWidget/

- `ChannelFormModal.module.css` — Styles: 28 rules (~840 tok)
- `ChannelFormModal.tsx` — ─── Types ────────────────────────────────────────────────────────────────────── (~1526 tok)
- `ChannelSettingsPanel.module.css` — Styles: 47 rules (~1893 tok)
- `ChannelSettingsPanel.tsx` — ─── Types ────────────────────────────────────────────────────────────────────── (~3420 tok)
- `chat.types.ts` — Exports MessageStatus, AttachmentType, DocumentModule, MessageAttachment + 6 more (~381 tok)
- `ChatButton.module.css` — Styles: 7 rules, 1 animations (~456 tok)
- `ChatButton.tsx` — ChatButton (~208 tok)
- `ChatInput.module.css` — Styles: 25 rules (~1153 tok)
- `ChatInput.tsx` — MAX_CHARS — uses useState, useEffect (~3085 tok)
- `ChatMessageList.module.css` — Styles: 40 rules (~1386 tok)
- `ChatMessageList.tsx` — CURRENT_USER_ID — uses useEffect (~2019 tok)
- `ChatPanel.module.css` — Styles: 50 rules, 2 media queries (~2261 tok)
- `ChatPanel.tsx` — ─── RBAC Constants ────────────────────────────────────────────────────────── (~1929 tok)
- `ChatWidget.tsx` — ChatWidget (~83 tok)
- `DocumentPickerModal.module.css` — Styles: 33 rules (~1361 tok)
- `DocumentPickerModal.tsx` — MODULE_LABELS — uses useState, useEffect (~1567 tok)
- `index.ts` (~87 tok)
- `MentionDropdown.module.css` — Styles: 12 rules (~493 tok)
- `MentionDropdown.tsx` — MAX_RESULTS — uses useState, useEffect (~730 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/CheckboxGroup/

- `CheckboxGroup.module.css` — Styles: 19 rules (~675 tok)
- `CheckboxGroup.tsx` — Show "select all" header checkbox (~1075 tok)
- `index.ts` (~30 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/CommandPalette/

- `CommandPalette.module.css` — Styles: 20 rules (~1060 tok)
- `CommandPalette.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~2272 tok)
- `index.ts` (~36 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ConfirmDialog/

- `ConfirmDialog.module.css` — Styles: 22 rules, 2 animations (~897 tok)
- `ConfirmDialog.tsx` — ─── Types ──────────────────────────────────────────────────────────────────── (~980 tok)
- `index.ts` (~45 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/CopyButton/

- `CopyButton.module.css` — Styles: 12 rules (~550 tok)
- `CopyButton.tsx` — RESET_DELAY_MS — uses useState (~606 tok)
- `index.ts` (~16 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DataConnectionWidget/

- `DataConnectionWidget.module.css` — Styles: 12 rules (~580 tok)
- `DataConnectionWidget.tsx` — Legacy: direct click handler. Prefer path + filters instead. (~969 tok)
- `index.ts` (~37 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DataTable/

- `DataTable.module.css` — Styles: 77 rules, 1 media queries, 2 animations (~3363 tok)
- `DataTable.tsx` — API sort param name. Defaults to `key` when not set. Use this when the (~6568 tok)
- `filter.types.ts` — Override default operators for this field type. (~369 tok)
- `filter.utils.ts` — Format lama: spread ke query params (status=active, name_contains=John). (~1533 tok)
- `FilterDialog.module.css` — Styles: 20 rules, 2 animations (~888 tok)
- `FilterDialog.tsx` — FilterDialog — uses useEffect (~651 tok)
- `FilterPanel.module.css` — Styles: 26 rules (~929 tok)
- `FilterPanel.tsx` — ─── Filter row ─────────────────────────────────────────────────────────────── (~1347 tok)
- `FilterValueInput.tsx` — BetweenInput (~1172 tok)
- `InlineFilter.module.css` — Styles: 11 rules (~421 tok)
- `InlineFilter.tsx` — InlineFilter — uses useState (~838 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DatePicker/

- `DatePicker.module.css` — Styles: 21 rules (~973 tok)
- `DatePicker.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~2012 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DateRangePicker/

- `DateRangePicker.module.css` — Styles: 36 rules (~1560 tok)
- `DateRangePicker.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~3686 tok)
- `index.ts` (~36 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DetailPageTemplate/

- `DetailPageTemplate.module.css` — Styles: 89 rules, 2 media queries (~5991 tok)
- `DetailPageTemplate.tsx` — Width of sidebar in pixels. Default: 360 (~3298 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/DomainPageTemplate/

- `DomainPageTemplate.module.css` — Styles: 82 rules (~4004 tok)
- `DomainPageTemplate.tsx` — Horizontal step flow bar shown below header card (~4346 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Drawer/

- `Drawer.module.css` — Styles: 10 rules (~513 tok)
- `Drawer.tsx` — SLIDE — renders modal — uses useEffect (~787 tok)
- `index.ts` (~10 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/EmptyState/

- `EmptyState.module.css` — Styles: 5 rules (~197 tok)
- `EmptyState.tsx` — EmptyState (~205 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ErrorBoundary/

- `ErrorBoundary.module.css` — Styles: 6 rules (~366 tok)
- `ErrorBoundary.tsx` — Custom fallback UI — receives error and retry fn (~744 tok)
- `index.ts` (~20 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/FileUpload/

- `FileUpload.module.css` — Styles: 22 rules (~934 tok)
- `FileUpload.tsx` — IMAGE_TYPES — uses useState, useCallback (~1401 tok)
- `index.ts` (~26 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/FlowWidget/

- `FlowWidget.module.css` — Styles: 43 rules, 1 media queries, 3 animations (~2302 tok)
- `FlowWidget.tsx` — InfoDialog — uses useCallback, useEffect (~1563 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/FocusTrap/

- `FocusTrap.tsx` — Whether to return focus to the previously focused element on unmount (~608 tok)
- `index.ts` (~12 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/FormPageTemplate/

- `CheckboxOption.tsx` — CheckboxOption (~217 tok)
- `Field.tsx` — Field (~191 tok)
- `FieldRow.tsx` — FieldRow (~65 tok)
- `FieldSection.tsx` — FieldSection (~95 tok)
- `FILE_UPLOAD_COMPONENTS.md` — File Upload Components (~1260 tok)
- `FileDropZone.tsx` — Alias untuk FileUploadField untuk backward compatibility (~37 tok)
- `FileUploadField.tsx` — Direct file handler — bypasses mediaService upload (~2455 tok)
- `FormColumn.tsx` — FormColumn (~66 tok)
- `FormGrid.tsx` — FormGrid (~65 tok)
- `FormPageTemplate.module.css` — Styles: 73 rules, 3 media queries (~3801 tok)
- `FormPageTemplate.tsx` — Page title: "Tambah Pelanggan" or "Edit Pelanggan" (~2307 tok)
- `index.ts` (~170 tok)
- `MultiFileUploadField.tsx` — MultiFileUploadField — uses useState (~2194 tok)
- `RadioGroup.tsx` — RadioGroup (~210 tok)
- `Toggle.tsx` — Toggle (~244 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/HierarchyTree/

- `HierarchyTree.module.css` — Styles: 10 rules (~352 tok)
- `HierarchyTree.tsx` — TreeNodeItem — uses useState (~884 tok)
- `index.ts` (~35 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/InlineEditField/

- `index.ts` (~15 tok)
- `InlineEditField.module.css` — Styles: 25 rules (~952 tok)
- `InlineEditField.tsx` — InlineEditField — uses useState, useEffect, useCallback (~1216 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/LazyImage/

- `index.ts` (~12 tok)
- `LazyImage.module.css` — Styles: 10 rules, 1 animations (~396 tok)
- `LazyImage.tsx` — Shown while loading (~613 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/LineTable/

- `index.ts` (~36 tok)
- `LineTable.module.css` — Styles: 36 rules (~1443 tok)
- `LineTable.tsx` — CSS width, e.g. '80px', '1fr', '20%'. Default: auto (~1679 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ListPageTemplate/

- `ListPageTemplate.module.css` — Styles: 25 rules (~1231 tok)
- `ListPageTemplate.tsx` — The API call. Should throw on error. (~2867 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/LoadingBar/

- `index.ts` (~32 tok)
- `LoadingBar.module.css` — Styles: 6 rules, 2 animations (~331 tok)
- `LoadingBar.tsx` — ─── Constants ──────────────────────────────────────────────────────────────── (~832 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Modal/

- `index.ts` (~10 tok)
- `Modal.module.css` — Styles: 8 rules (~564 tok)
- `Modal.tsx` — SIZE_WIDTH — renders modal — uses useEffect (~759 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Modals/

- `DeleteConfirmModal.module.css` — Styles: 14 rules, 2 animations (~744 tok)
- `DeleteConfirmModal.tsx` — ─── Store ──────────────────────────────────────────────────────────────────── (~721 tok)
- `InfoAlertDialog.module.css` — Styles: 10 rules, 2 animations (~534 tok)
- `InfoAlertDialog.tsx` — ─── Store ──────────────────────────────────────────────────────────────────── (~471 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/MultiSelect/

- `index.ts` (~29 tok)
- `MultiSelect.module.css` — Styles: 35 rules, 1 animations (~1560 tok)
- `MultiSelect.tsx` — MultiSelect — uses useState, useCallback (~1828 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/NotificationBell/

- `index.ts` (~39 tok)
- `NotificationBell.module.css` — Styles: 32 rules, 1 animations (~1444 tok)
- `NotificationBell.tsx` — ─── Types ──────────────────────────────────────────────────────────────────── (~1614 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/NumberInput/

- `index.ts` (~13 tok)
- `NumberInput.module.css` — Styles: 17 rules (~747 tok)
- `NumberInput.tsx` — NumberInput — uses useCallback (~870 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/PageWrapper/

- `PageWrapper.module.css` — Styles: 11 rules, 2 animations (~456 tok)
- `PageWrapper.tsx` — PageSkeleton (~340 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Pagination/

- `index.ts` (~12 tok)
- `Pagination.module.css` — Styles: 10 rules (~550 tok)
- `Pagination.tsx` — DEFAULT_PAGE_SIZES (~1106 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/PermissionGate/

- `index.ts` (~32 tok)
- `PermissionGate.tsx` — ─── Types ──────────────────────────────────────────────────────────────────── (~313 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Popover/

- `index.ts` (~11 tok)
- `Popover.module.css` — Styles: 1 rules (~97 tok)
- `Popover.tsx` — DEFAULT_OFFSET — uses useState, useCallback, useEffect (~1262 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ProgressWidget/

- `ProgressWidget.module.css` — Styles: 20 rules, 1 media queries (~906 tok)
- `ProgressWidget.tsx` — Harus cocok dengan nilai currentStatus (~1013 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/QuickLinkCard/

- `QuickLinkCard.module.css` — Styles: 6 rules (~303 tok)
- `QuickLinkCard.tsx` — QuickLinkCard (~203 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/RadioGroup/

- `index.ts` (~26 tok)
- `RadioGroup.module.css` — Styles: 20 rules, 1 animations (~958 tok)
- `RadioGroup.tsx` — RadioGroup (~594 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/RangeInput/

- `index.ts` (~12 tok)
- `RangeInput.module.css` — Styles: 16 rules (~721 tok)
- `RangeInput.tsx` — DEFAULT_FORMAT — uses useCallback (~739 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/ReportIndexCard/

- `ReportIndexCard.module.css` — Styles: 11 rules, 2 media queries (~457 tok)
- `ReportIndexCard.tsx` — ─── ReportIndexCard ────────────────────────────────────────────────────────── (~346 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/SearchableSelect/

- `SearchableSelect.module.css` — Styles: 28 rules, 1 animations (~1131 tok)
- `SearchableSelect.tsx` — ID yang disimpan ke form state (~1959 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/SectionCard/

- `index.ts` (~13 tok)
- `SectionCard.module.css` — Styles: 13 rules (~448 tok)
- `SectionCard.tsx` — Makes the section collapsible (~815 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Skeleton/

- `index.ts` (~24 tok)
- `Skeleton.module.css` — Styles: 9 rules, 1 animations (~496 tok)
- `Skeleton.tsx` — Single skeleton block (~636 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/StatCard/

- `StatCard.module.css` — Styles: 13 rules, 1 animations (~436 tok)
- `StatCard.tsx` — Sparkline — uses useEffect (~1058 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/StatusPills/

- `StatusPills.module.css` — Styles: 3 rules (~148 tok)
- `StatusPills.tsx` — Renders both pills inline — pass into header slots (~272 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Stepper/

- `index.ts` (~27 tok)
- `Stepper.module.css` — Styles: 31 rules (~1183 tok)
- `Stepper.tsx` — getStatus (~832 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Switch/

- `index.ts` (~22 tok)
- `Switch.module.css` — Styles: 24 rules (~630 tok)
- `Switch.tsx` — Switch (~381 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Tabs/

- `index.ts` (~20 tok)
- `Tabs.module.css` — Styles: 15 rules (~750 tok)
- `Tabs.tsx` — Tabs (~482 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/TagInput/

- `index.ts` (~11 tok)
- `TagInput.module.css` — Styles: 15 rules (~732 tok)
- `TagInput.tsx` — Characters that trigger tag creation. Defaults to Enter and comma. (~897 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Timeline/

- `index.ts` (~25 tok)
- `Timeline.module.css` — Styles: 25 rules (~902 tok)
- `Timeline.tsx` — Timeline (~547 tok)

## .claude/worktrees/agent-a04f6200/web-dashboard/src/widgets/Toast/

- `Toast.module.css` — Styles: 15 rules, 1 media queries (~455 tok)
- `Toast.tsx` — useToastStore — uses useEffect (~772 tok)

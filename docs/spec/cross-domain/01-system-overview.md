# 01 — Gambaran Umum Sistem SekolahPro

Dokumen ini menyajikan pandangan tinggi terhadap keseluruhan sistem SekolahPro: diagram sistem lengkap, tabel interaksi antar modul, siklus hidup tenant, peta kepemilikan data, dan ringkasan tumpukan teknologi.

---

## 1. Diagram Sistem Lengkap (ASCII)

```
╔══════════════════════════════════════════════════════════════════════════════════╗
║                           SekolahPro Platform                                    ║
║                   (Go Monolith · Multi-Tenant SaaS + Self-Hosted)                ║
╠══════════════════╦══════════════════════════════════════════════════════════════╣
║                  ║                    HTTP Layer                                  ║
║   KLIEN          ║   Chi Router + JWT Middleware (Two-Phase) + Scope Middleware   ║
║                  ║                                                                ║
║  ┌───────────┐   ║  ┌──────────────┐  ┌─────────────────┐  ┌──────────────────┐ ║
║  │ Web Admin │──▶║  │ Sekolah      │  │ Koperasi        │  │ Core / Shared    │ ║
║  │ Dashboard │   ║  │ Handlers     │  │ Handlers        │  │ Handlers         │ ║
║  └───────────┘   ║  └──────┬───────┘  └──────┬──────────┘  └────────┬─────────┘ ║
║                  ║         │                  │                       │            ║
║  ┌───────────┐   ║         ▼                  ▼                       ▼            ║
║  │ Parent    │──▶║  ┌──────────────────────────────────────────────────────────┐  ║
║  │ Portal    │   ║  │                  USECASE LAYER                            │  ║
║  └───────────┘   ║  │  Command Handlers + Query Handlers + Port Interfaces     │  ║
║                  ║  └────────┬──────────────┬────────────────────┬─────────────┘  ║
║  ┌───────────┐   ║           │              │                    │                 ║
║  │ Mobile    │──▶║           ▼              ▼                    ▼                 ║
║  │ App       │   ║  ┌──────────────┐ ┌────────────┐  ┌──────────────────────────┐ ║
║  └───────────┘   ║  │ Repository   │ │ Vernon     │  │ EventBus Interface        │ ║
║                  ║  │ (Write)      │ │ Reader     │  │ InMemory (dev)            │ ║
║  ┌───────────┐   ║  │ sqlc/sqlx    │ │ (Read,0JN) │  │ NATS JetStream (prod)    │ ║
║  │ Nasabah   │──▶║  └──────┬───────┘ └─────┬──────┘  └──────────┬───────────────┘ ║
║  │ Portal    │   ║         │               │                     │                  ║
║  └───────────┘   ║         ▼               ▼                     ▼                  ║
║                  ║  ┌─────────────────────────────────────────────────────────────┐ ║
╠══════════════════╣  │                  PostgreSQL                                 │ ║
║  EKSTERNAL       ║  │  Sekolah Tables  │  Koperasi Tables  │  Core/Shared Tables  │ ║
║                  ║  │  (tenant_id,     │  (tenant_id,      │  (tenants,companies) │ ║
║  ┌───────────┐   ║  │   company_id)    │   branch_id)      │  users, roles        │ ║
║  │ DAPODIK   │◀──║  └─────────────────────────────────────────────────────────────┘ ║
║  │ Kemendikbud│  ║                                                                   ║
║  └───────────┘   ║  ┌──────────────────────────────────────────────────────────────┐ ║
║                  ║  │                      Redis Cache                              │ ║
║  ┌───────────┐   ║  └──────────────────────────────────────────────────────────────┘ ║
║  │ Midtrans/ │◀──║                                                                    ║
║  │ Xendit    │──▶║  ┌─────────────────────────────────────────────────────────────┐  ║
║  └───────────┘   ║  │     NATS JetStream (Event Bus Production)                   │  ║
║                  ║  │  Stream: DOMAIN_EVENTS (7d)  │  Stream: SYNC_ENGINE (24h)   │  ║
║  ┌───────────┐   ║  └─────────────────────────────────────────────────────────────┘  ║
║  │ WA/SMS    │──▶║                                                                    ║
║  │ Provider  │   ║  ┌─────────────────────────────────────────────────────────────┐  ║
║  └───────────┘   ║  │     Object Storage (S3 / MinIO)                              │  ║
║                  ║  │  Dokumen siswa · Foto · Rapor PDF · Statement Koperasi       │  ║
║  ┌───────────┐   ║  └─────────────────────────────────────────────────────────────┘  ║
║  │ LMS       │◀──║                                                                    ║
║  │ (Moodle,  │──▶║                                                                    ║
║  │  GClass)  │   ║                                                                    ║
║  └───────────┘   ║                                                                    ║
╚══════════════════╩════════════════════════════════════════════════════════════════════╝
```

### Hubungan Antar Domain Group

```
┌──────────────────────────────────────────────────────────────────────┐
│                         CORE / SHARED                                 │
│  tenants · companies · branches · warehouses · users · roles          │
│  event bus abstraction · Vernon SyncEngine · audit logs               │
└──────────┬────────────────────────────────────────────┬──────────────┘
           │ menyediakan fondasi                         │ menyediakan fondasi
           │ multi-tenant + auth                         │ multi-tenant + auth
           ▼                                             ▼
┌─────────────────────────┐              ┌───────────────────────────────┐
│   DOMAIN SEKOLAH        │              │   DOMAIN KOPERASI             │
│  (58 ADR S001-S058)     │              │  (40 ADR K001-K040)           │
│                         │              │                               │
│  students               │──events──▶  │  nasabah                      │
│  teachers               │──events──▶  │  payroll_deduction            │
│  student_finance (SPP)  │  (terpisah) │  rekening · transaksi         │
│  academic_years         │              │  pinjaman · deposito          │
│  class_rooms            │              │  toko/kantin                  │
│  rapor                  │              │  SHU · COA · jurnal           │
│  PPDB / admission       │              │  zakat (Islamic mode)         │
│  parent portal (S042)   │◀──events── │  parent portal (K023)         │
│  notification (S044)    │◀──────────▶│  notification (K022)          │
│  payment gw (S051)      │              │  payment gw (K024)            │
│  dapodik (S055)         │              │  OJK / Dinas Koperasi (K017)  │
└─────────────────────────┘              └───────────────────────────────┘
         │  school_entity_id                       ▲
         └─────────────────────────────────────────┘
           nasabah.school_entity_id → students.id / teachers.id
           (Sekolah adalah source of truth untuk data person)
```

---

## 2. Tabel Interaksi Modul

| Sumber | Tujuan | Mekanisme | Tipe | Keterangan |
|--------|--------|-----------|------|------------|
| Sekolah (S001 students) | Koperasi (K001 nasabah) | Domain Event: `StudentActivated` | ASYNC | Trigger pembuatan application nasabah koperasi |
| Sekolah (ADR-012 teachers) | Koperasi (K020 payroll) | Domain Event: `TeacherPayrollDataUpdated` | ASYNC | Data gaji guru untuk potongan simpanan/angsuran |
| Sekolah (S009 SPP) | Core Payment (S051) | SYNC | SYNC | Invoice SPP diteruskan ke payment gateway |
| Core Payment (S051) | Sekolah (S009) | Webhook callback | ASYNC | Konfirmasi pembayaran → update invoice |
| Core Payment (S051) | Sekolah (S037 kantin) | Webhook callback | ASYNC | Top-up dompet kantin siswa |
| Sekolah (S008 absensi) | Sekolah (S044 notif) | Domain Event: `StudentAbsent` | ASYNC | Notifikasi ke orang tua |
| Sekolah (S009 SPP) | Sekolah (S044 notif) | Domain Event: `InvoiceDue` | ASYNC | Notifikasi tagihan jatuh tempo |
| Sekolah (S042 parent portal) | Koperasi (K023 portal) | Satu JWT user | SYNC | User sama, context berbeda dalam token |
| Koperasi (K022 notif) | Sekolah (S044 notif) | Shared notification infra | ASYNC | Konsolidasi notifikasi, mencegah spam |
| Sekolah (S055 dapodik) | Kemendikbud DAPODIK | Manual export / API future | ASYNC | Sinkronisasi data pendidikan nasional |
| Sekolah (S053 e-learning) | LMS Eksternal | REST API + Webhook | ASYNC | Sinkronisasi tugas & nilai |
| Koperasi (K017 laporan) | OJK / Dinas Koperasi | Manual file / e-filing | ASYNC | Pelaporan regulasi |
| Koperasi (K024) | Midtrans/Xendit | REST API | SYNC/ASYNC | Virtual account, QRIS, callback |
| Core SyncEngine | Semua domain Vernon | Domain Event + Batch UPDATE | ASYNC | Sinkronisasi `_data` JSONB cache |

---

## 3. Siklus Hidup Tenant (Onboarding Sekolah + Koperasi)

```
FASE 1 — REGISTRASI TENANT
─────────────────────────────────────────────────────────────────────
1. Super Admin buat Tenant baru
   └── Set: school_type = "general"|"islamic"
             coop_type  = "general"|"islamic"
   ⚠ IMMUTABLE setelah diaktifkan

2. Buat Company (sekolah unit)
   └── Set NPSN (Nomor Pokok Sekolah Nasional)

3. Buat Branch (opsional — untuk kampus cabang)

4. Buat Warehouse (untuk Koperasi — kasir/gudang)

5. System auto-seed:
   └── System roles default (admin, kepsek, guru, bendahara, orang_tua, ...)
   └── COA default Koperasi (sesuai coop_type)
   └── Produk koperasi default (simpanan pokok, simpanan wajib, tabungan)


FASE 2 — SETUP SEKOLAH
─────────────────────────────────────────────────────────────────────
6. Setup Academic Year (Tahun Ajaran) — status: planning → active
   └── Set semester1_start/end, semester2_start/end

7. Import / input Guru & Staff (template Excel atau manual)
   └── Link user_id untuk yang butuh akses sistem

8. Setup Class Rooms (Kelas)
   └── Assign wali kelas per kelas

9. Import / PPDB Siswa
   └── Mode import Excel: batch processing + Vernon sync
   └── Mode PPDB online: applicant → student → class placement

10. Setup SPP: template tagihan, nominal per jenjang
    └── Generate invoice per siswa per bulan


FASE 3 — SETUP KOPERASI
─────────────────────────────────────────────────────────────────────
11. Konfigurasi Produk Koperasi (tabungan, deposito, pinjaman)
    └── Set rate/nisbah, limit, akad (Islamic mode)

12. Registrasi Nasabah
    Mode A — Auto dari Sekolah:
      └── Event StudentActivated → buat nasabah_application otomatis
      └── Tetap melalui approval flow (tidak langsung aktif)
    Mode B — Manual:
      └── Teller input data via form

13. Buka rekening simpanan pokok + simpanan wajib (otomatis saat nasabah approved)

14. Konfigurasi Teller + Kas Awal
    └── Set denominasi uang, posisi kas awal

15. Setup integrasi pembayaran (VA, QRIS)


FASE 4 — OPERASIONAL
─────────────────────────────────────────────────────────────────────
16. Sistem berjalan normal:
    ├── Sekolah: absensi, nilai, SPP, rapor, PPDB
    ├── Koperasi: simpanan, pinjaman, teller session, SHU
    └── Cross: notifikasi terpadu, parent portal terintegrasi
```

---

## 4. Peta Kepemilikan Data (Source of Truth)

| Entitas | Domain Pemilik | Tabel Utama | Consumer |
|---------|---------------|-------------|----------|
| Profil Siswa | **Sekolah** | `students` | Koperasi (nasabah.school_entity_id) |
| Data Guru/Staff | **Sekolah** | `teachers` | Koperasi (payroll_deduction) |
| Data Orang Tua/Wali | **Sekolah** | `student_guardians` | Notifikasi Koperasi (via shared user) |
| Tahun Ajaran | **Sekolah** | `academic_years` | Referensi lintas domain Sekolah |
| Kelas | **Sekolah** | `class_rooms` | Absensi, Nilai, Jadwal |
| Nasabah Koperasi | **Koperasi** | `nasabah` | Rekening, Transaksi, SHU |
| Rekening Keuangan | **Koperasi** | `rekening` | Transaksi, Laporan |
| Transaksi Keuangan | **Koperasi** | `transaksi` | Jurnal, SHU |
| SPP / Invoice Sekolah | **Sekolah** | `student_invoices` | Payment Gateway (S051) |
| Pembayaran Online | **Sekolah (S051)** | `payment_transactions` | Invoice S009, Kantin S037 |
| User Akun Login | **Core** | `users` | Sekolah + Koperasi (shared JWT) |
| Role & Permission | **Core** | `roles`, `user_roles` | Seluruh sistem |
| Tenant & Company | **Core** | `tenants`, `companies` | Seluruh sistem |
| Jurnal Akuntansi | **Koperasi** | `jurnal`, `jurnal_line` | Laporan keuangan koperasi |
| COA | **Koperasi** | `coa` | Jurnal, laporan |
| Notifikasi Log (Sekolah) | **Sekolah** | `notification_logs` | Audit trail |
| Notifikasi Log (Koperasi) | **Koperasi** | `notifikasi_log` | Audit trail |

**Prinsip Utama Kepemilikan:**
- Sekolah adalah **source of truth** untuk semua data person (siswa, guru, staf, orang tua)
- Koperasi mengonsumsi data person via `school_entity_id` reference — **tidak pernah duplikasi data**
- Core adalah pemilik identitas autentikasi (`users`) — kedua domain bergantung pada data ini
- Setiap domain mengelola data keuangannya sendiri: Sekolah → SPP/invoice, Koperasi → rekening/transaksi

---

## 5. Ringkasan Tumpukan Teknologi

### Backend
| Komponen | Teknologi | Keterangan |
|----------|-----------|------------|
| Runtime | **Go** | Single monolith binary |
| Framework HTTP | **Chi** | Lightweight router |
| Dependency Injection | **Uber FX** | Lifecycle management + wiring |
| Database | **PostgreSQL** | Primary datastore, semua domain |
| Query Layer | **sqlc + sqlx** | Type-safe queries, zero SQL string injection |
| Schema Migration | **golang-migrate** | Versioned migrations |
| Cache | **Redis** | Session, rate limiting, temp data |
| Event Bus (dev) | **InMemory** | Zero-config untuk development |
| Event Bus (prod) | **NATS JetStream** | At-least-once, durable, replay |
| Primary Key | **UUID v7** | Time-sortable, globally unique, B-tree friendly |
| Object Storage | **S3 / MinIO** | Dokumen, foto, PDF |

### Arsitektur Pola
| Pola | Penerapan |
|------|-----------|
| **Go Clean Architecture** | Seluruh sistem — domain → usecase → adapter → infrastructure |
| **CQRS** | Seluruh sistem — command (write) terpisah dari query (read) |
| **Vernon Pattern** | Domain read-heavy (≥3 JOIN, read:write ≥10:1) |
| **Standard CQRS** | Domain sederhana (auth, users, permissions, notifications) |
| **Multi-Tenant RLS** | Application layer — `tenant_id` wajib di semua query |
| **Two-Phase JWT** | Auth: Phase 1 (5 menit) → Phase 2 (1 jam, full scope) |
| **Strategy Pattern** | Business rule berbeda per `school_type`/`coop_type` |
| **Event-Driven Sync** | SyncEngine memperbarui Vernon `_data` secara async |

### Observabilitas
| Komponen | Teknologi |
|----------|-----------|
| Tracing | OpenTelemetry |
| Metrics | Prometheus + `/metrics` endpoint |
| Logging | zerolog |
| Health Check | `/healthz` (liveness) + `/readyz` (readiness) |

### Deployment Mode
| Mode | Trigger | Keterangan |
|------|---------|------------|
| `DEPLOYMENT_MODE=single` | Self-hosted sekolah | Skip tenant selection |
| `DEPLOYMENT_MODE=saas` | Multi-tenant SaaS | Two-phase JWT penuh |

### Dual-Mode Institution
| Kombinasi | `school_type` | `coop_type` | Target Pasar |
|-----------|--------------|-------------|--------------|
| 1 | `general` | `general` | Sekolah umum + Koperasi konvensional |
| 2 | `general` | `islamic` | Sekolah umum + BMT |
| 3 | `islamic` | `general` | Pesantren + Koperasi konvensional |
| 4 | `islamic` | `islamic` | **Pesantren + BMT (target utama)** |

---

## Key Constraints

1. **Tidak ada cross-domain DB query** — domain Koperasi tidak boleh langsung JOIN ke tabel domain Sekolah
2. **Semua integrasi lintas domain via event bus** atau via FK `school_entity_id` yang di-resolve melalui API call
3. **`school_type` dan `coop_type` immutable** setelah tenant aktif
4. **Satu user bisa memiliki role di kedua domain** — JWT Phase 2 menyertakan semua permissions dari kedua domain
5. **Eventual consistency di read path** — Vernon `_data` bisa stale maksimal 30 detik (normal), 1 detik (critical)

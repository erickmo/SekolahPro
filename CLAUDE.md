# CLAUDE.md — Ekosistem Digital Sekolah (EDS)

Platform SaaS multi-tenant untuk digitalisasi penuh operasional sekolah — dari administrasi, akademik, keuangan, kesehatan, hingga ekosistem bisnis sekolah. Satu platform, ribuan sekolah.

> **Terakhir diperbarui**: 20 Maret 2026
> **Versi dokumen**: 3.0
> **Perubahan v3.0**: PostgreSQL 17, integrasi module-vernon-hrm & module-vernon-accounting, Developer Portal + Listing Marketplace, split docs per domain.

---

## Navigasi Dokumen

Dokumentasi dibagi per domain. Baca file yang relevan dengan tugas yang sedang dikerjakan.

| File | Domain | Konten |
|------|--------|--------|
| `CLAUDE.md` | **Index** | Arsitektur, stack, prinsip, roadmap (file ini) |
| `docs/01-saas-platform.md` | **SaaS & Bisnis** | Developer Portal, pricing, tenant management, listing berbayar |
| `docs/02-core-modules.md` | **Core (M01–M12)** | SIMS, kartu siswa, koperasi, akademik, ujian AI, ekskul, website, PPDB, laporan, dashboard |
| `docs/03-komunikasi.md` | **Komunikasi (M13–M17)** | Portal ortu, notifikasi WA, forum, video conference, majalah |
| `docs/04-keuangan.md` | **Keuangan (M18–M22, M51–M52)** | SPP online, kantin, aset, perpustakaan, beasiswa, RKAS, dana komite |
| `docs/05-ai-analitik.md` | **AI & Analitik (M23–M26b, M58)** | EWS, learning path, prediktif, AI tutor, plagiarisme, sentiment, data warehouse |
| `docs/06-kesehatan.md` | **Kesehatan (M27–M30)** | UKS, konseling BK, anti-bullying, gizi |
| `docs/07-operasional.md` | **Operasional (M31–M35, M56–M57)** | SDM guru, smart gate, booking, transportasi, helpdesk, seragam, event |
| `docs/08-pedagogis.md` | **Pedagogis (M48–M50, M55)** | RPP digital, penilaian proyek/KKTP, remedial, ABK |
| `docs/09-karir-alumni.md` | **Karir & Alumni (M36–M39)** | Alumni, portofolio, prakerin, pelatihan guru |
| `docs/10-komunitas.md` | **Komunitas (M53–M54)** | Komite sekolah, volunteer |
| `docs/11-integrasi.md` | **Integrasi (M40–M44, M59–M61)** | Dapodik, LMS, white-label, yayasan, open API, IoT, gamifikasi, marketplace konten |
| `docs/12-platform-infra.md` | **Infrastruktur (M45–M47)** | Audit log, backup/recovery, security center |
| `docs/13-rbac-auth.md` | **Auth & RBAC** | Role hierarchy, akses per modul, JWT, SSO |
| `docs/14-database.md` | **Database** | Prisma schema lengkap semua domain |
| `docs/15-dev-guide.md` | **Dev Guide** | Quick start, konvensi, env vars, testing, deployment |
| `docs/16-vernon-hrm.md` | **Vernon HRM** | Integrasi module-vernon-hrm → M31 SDM & Presensi Guru |
| `docs/17-vernon-accounting.md` | **Vernon Accounting** | Integrasi module-vernon-accounting → M03, M18, M51, M52 |

---

## Visi & Model Bisnis

EDS adalah **SaaS platform** dengan dua lapisan bisnis:

### 1. Langganan Sekolah (B2B)
Sekolah berlangganan akses ke modul-modul EDS. Setiap sekolah mendapat subdomain sendiri (`namaSekolah.eds.id`) dengan branding kustom.

```
Tier FREE    → M01, M02, M07, M08 (basic)
Tier BASIC   → +M03, M04, M05 (AI soal terbatas), M09, M12
              Rp 299.000/bulan
Tier PRO     → Semua modul Core + Komunikasi + Keuangan + AI
              Rp 799.000/bulan
Tier ENTERPRISE → Semua modul + white-label domain + priority support
              Custom pricing
```

### 2. Listing Berbayar (B2B2C Marketplace)
Portal EDS menjadi marketplace untuk penyedia layanan yang berkaitan dengan ekosistem sekolah. Mereka membayar biaya listing untuk muncul di halaman sekolah yang relevan.

```
Kategori Listing:
├── Guru Les & Bimbingan Belajar  (M11)
├── Tempat Kursus & Lembaga Pelatihan
├── Jasa Antar-Jemput Sekolah     (M34)
├── Catering & Katering Sekolah   (M19)
├── Penerbit & Toko Buku          (M10)
├── Seragam & Atribut Sekolah     (M56)
├── Vendor Kegiatan (fotografer, dekorasi, EO)
├── Asuransi Pendidikan
└── Tabungan & Investasi Pendidikan
```

> Detail lengkap: lihat `docs/01-saas-platform.md`

---

## Arsitektur Sistem

```
┌──────────────────────────────────────────────────────────────────────┐
│                          EDS PLATFORM                                │
│                                                                      │
│  ┌─────────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │  Developer Portal│  │  Admin Panel │  │  School Portals       │  │
│  │  (SaaS ops,      │  │  (EDS team)  │  │  web + mobile app     │  │
│  │   listing mgmt)  │  │              │  │  per tenant           │  │
│  └────────┬─────────┘  └──────┬───────┘  └──────────┬────────────┘  │
│           │                   │                     │               │
│  ┌────────▼───────────────────▼─────────────────────▼────────────┐  │
│  │              API Gateway + Auth & IAM (RBAC + JWT)             │  │
│  └────────┬────────────────────────────────────────┬─────────────┘  │
│           │                                        │               │
│  ┌────────▼──────────┐              ┌──────────────▼────────────┐  │
│  │  Core Services    │              │  AI Engine                │  │
│  │  (62 modul)       │              │  Claude API               │  │
│  └────────┬──────────┘              └──────────────┬────────────┘  │
│           │                                        │               │
│  ┌────────▼────────────────────────────────────────▼────────────┐  │
│  │                    DATABASE LAYER                              │  │
│  │  PostgreSQL 17 (RLS per tenant) + Redis 7 + ClickHouse (BI)  │  │
│  └───────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
```

### Stack Teknologi

| Layer | Teknologi |
|-------|-----------|
| **Frontend Web** | Next.js 14 (App Router) + TypeScript |
| **Developer Portal** | Next.js 14 — terpisah dari school portal |
| **Frontend Mobile** | React Native / Expo |
| **Backend API** | Node.js + Fastify (REST + tRPC) |
| **AI Engine** | Anthropic Claude API (`claude-sonnet-4-20250514`) |
| **Database** | PostgreSQL 17 + Prisma ORM |
| **Multi-tenant isolation** | PostgreSQL Row Level Security (RLS) |
| **Cache** | Redis 7 |
| **Auth** | NextAuth.js + JWT + RBAC |
| **File Storage** | MinIO (self-hosted) / S3-compatible |
| **Notifikasi WA** | WhatsApp Business API (360dialog / Wati) |
| **Payment** | Midtrans / Xendit (SPP) + Stripe (SaaS billing) |
| **Real-time** | Socket.io |
| **Video** | Jitsi Meet self-hosted |
| **BI / Warehouse** | ClickHouse + Apache Superset |
| **IoT** | MQTT + Node-RED |
| **HRM Engine** | module-vernon-hrm (M31 SDM & Presensi Guru) |
| **Accounting Engine** | module-vernon-accounting (M03, M18, M51, M52) |
| **Deployment** | Docker Compose → Kubernetes |
| **CI/CD** | GitHub Actions |

---

## Peta Modul Lengkap (62 Modul)

### Tier & Harga per Modul

| Modul | Nama | Tier |
|-------|------|------|
| M01 | Administrasi Siswa (SIMS) | FREE |
| M02 | Kartu Siswa Digital | FREE |
| M03 | Koperasi & Tabungan | BASIC |
| M04 | Akademik & Kurikulum | BASIC |
| M05 | Ujian & AI Generate Soal ⭐ | PRO (AI quota) |
| M06 | Ekstrakurikuler | BASIC |
| M07 | Website Sekolah | FREE |
| M08 | PPDB Online | BASIC |
| M09 | Laporan Dinas | BASIC |
| M10 | Stakeholder & Publisher | PRO |
| M11 | Guru Les / Bimbel | LISTING |
| M12 | Dashboard Kepala Sekolah | BASIC |
| M13 | Portal Orang Tua | BASIC |
| M14 | Messaging & Notifikasi WA | PRO (quota WA) |
| M15 | Forum & Komunitas | BASIC |
| M16 | Rapat & Video Conference | PRO |
| M17 | Blog & Majalah Sekolah | BASIC |
| M18 | Pembayaran SPP Online | PRO (% transaksi) |
| M19 | Kantin Digital & Cashless | LISTING + PRO |
| M20 | Manajemen Aset | BASIC |
| M21 | Perpustakaan Digital | BASIC |
| M22 | Beasiswa & Bantuan Siswa | FREE |
| M23 | Early Warning System | PRO (AI) |
| M24 | Personalized Learning Path | PRO (AI) |
| M25 | Analitik Prediktif | ENTERPRISE (AI) |
| M26 | AI Tutor & Anti-Plagiarisme | PRO (AI quota) |
| M26b | Sentiment Analysis | PRO (AI) |
| M27 | UKS & Rekam Medis | BASIC |
| M28 | Konseling & BK Digital | BASIC |
| M29 | Anti-Bullying & Safety | FREE |
| M30 | Pemantauan Gizi | BASIC |
| M31 | SDM & Presensi Guru | BASIC |
| M32 | Smart Gate & Keamanan | PRO (hardware opt) |
| M33 | Booking Ruang & Fasilitas | BASIC |
| M34 | Transportasi Sekolah | LISTING + PRO |
| M35 | Helpdesk & Ticketing | BASIC |
| M36 | Alumni & Tracer Study | BASIC |
| M37 | Portofolio Digital Siswa | FREE |
| M38 | Prakerin / Magang (SMK) | PRO |
| M39 | Pelatihan & CPD Guru | BASIC |
| M40 | Dapodik / EMIS / ARKAS | BASIC |
| M41 | LMS Integration | PRO |
| M42 | Multi-tenant & White-label | ENTERPRISE |
| M43 | Dashboard Yayasan | ENTERPRISE |
| M44 | Open API & Marketplace | ENTERPRISE |
| M45 | Audit Log & Compliance | ALL (on by default) |
| M46 | Backup & Recovery | ALL (on by default) |
| M47 | Security Center | ALL (on by default) |
| M48 | RPP Digital | BASIC |
| M49 | Penilaian Proyek & KKTP | PRO |
| M50 | Remedial & Pengayaan | BASIC |
| M51 | Anggaran & RKAS | BASIC |
| M52 | Dana Komite | BASIC |
| M53 | Komite Sekolah Digital | BASIC |
| M54 | Volunteer & Komunitas | FREE |
| M55 | Manajemen ABK | PRO |
| M56 | Seragam & Atribut | LISTING |
| M57 | Event & Kepanitiaan | BASIC |
| M58 | Data Warehouse & BI | ENTERPRISE |
| M59 | IoT & Smart Campus | ENTERPRISE (hardware) |
| M60 | Gamifikasi & Loyalitas | PRO |
| M61 | Marketplace Konten Edukatif | LISTING + revenue share |

> Tier FREE = akses gratis, BASIC = paket Rp 299k/bln, PRO = Rp 799k/bln,
> ENTERPRISE = custom, LISTING = vendor bayar ke EDS bukan sekolah.

---

## Struktur Repositori

```
eds/
├── CLAUDE.md                    ← file ini (index)
├── docs/
│   ├── 01-saas-platform.md      ← Developer Portal, pricing, listing
│   ├── 02-core-modules.md       ← M01–M12
│   ├── 03-komunikasi.md         ← M13–M17
│   ├── 04-keuangan.md           ← M18–M22, M51–M52
│   ├── 05-ai-analitik.md        ← M23–M26b, M58
│   ├── 06-kesehatan.md          ← M27–M30
│   ├── 07-operasional.md        ← M31–M35, M56–M57
│   ├── 08-pedagogis.md          ← M48–M50, M55
│   ├── 09-karir-alumni.md       ← M36–M39
│   ├── 10-komunitas.md          ← M53–M54
│   ├── 11-integrasi.md          ← M40–M44, M59–M61
│   ├── 12-platform-infra.md     ← M45–M47
│   ├── 13-rbac-auth.md          ← Auth & RBAC lengkap
│   ├── 14-database.md           ← Prisma schema semua domain
│   ├── 15-dev-guide.md          ← Quick start, env vars, testing
│   ├── 16-vernon-hrm.md         ← module-vernon-hrm → M31
│   └── 17-vernon-accounting.md  ← module-vernon-accounting → M03, M18, M51, M52
│
├── apps/
│   ├── web/                     # School portals (Next.js)
│   ├── developer-portal/        # SaaS ops + listing mgmt (Next.js)
│   ├── mobile/                  # React Native
│   ├── ppdb/                    # PPDB standalone
│   └── foundation/              # Dashboard yayasan
│
├── packages/
│   ├── api/
│   │   └── modules/             # 62 modul, dikelompokkan per domain
│   ├── database/                # Prisma + migrations (PostgreSQL 17)
│   ├── warehouse/               # ETL + ClickHouse
│   ├── ai/                      # Claude wrappers & prompts
│   ├── iot/                     # MQTT + Node-RED
│   ├── notifications/           # WA, push, email, SMS
│   ├── billing/                 # Stripe + Midtrans subscription
│   ├── listing/                 # Listing marketplace engine
│   ├── hrm/                     # module-vernon-hrm adapter
│   ├── accounting/              # module-vernon-accounting adapter
│   ├── ui/                      # Shared components
│   └── config/
│
└── infrastructure/
    ├── docker-compose.yml        # postgres:17, redis:7, clickhouse, dll
    ├── kubernetes/
    └── nginx/
```

---

## Prinsip Desain

1. **SaaS-first** — setiap fitur harus bisa diaktifkan/dimatikan per tenant dari Developer Portal.
2. **Multi-tenant isolation** — PostgreSQL RLS wajib. Query tanpa `schoolId` adalah bug, bukan fitur.
3. **Listing is a first-class citizen** — vendor listing bukan fitur sampingan; mereka adalah revenue stream utama kedua.
4. **Privacy by design** — data siswa (NIK, kesehatan, catatan BK, ABK) dienkripsi AES-256 at-rest.
5. **Security by default** — M45 Audit Log + M47 Security Center aktif di semua tier termasuk FREE.
6. **Offline-first mobile** — absensi & nilai bisa diinput offline, sync saat koneksi tersedia.
7. **Mobile-first** — 80% user akses via smartphone. Desain untuk layar kecil + koneksi lambat.
8. **AI sebagai asisten** — output AI selalu bisa di-override manusia. Tidak ada keputusan otomatis tentang nilai atau kenaikan kelas.
9. **Graceful degradation** — jika AI atau layanan eksternal down, sistem tetap berjalan.
10. **Zero downtime deploy** — blue-green deployment. Deploy tidak boleh saat jam 07.00–16.00.
11. **Immutable audit** — log M45 tidak bisa diedit/dihapus oleh siapapun termasuk superadmin.
12. **Transparent finance** — keuangan sekolah (M51, M52) dapat diakses ortu sesuai wewenang.
13. **PDPA compliant** — UU PDP No. 27/2022. Data siswa tidak dijual atau digunakan untuk iklan.
14. **Inclusive** — M55 ABK terintegrasi ke semua modul relevan, bukan add-on terpisah.

---

## Roadmap

```
Phase 1 — Foundation (Bln 1-3)
  Auth + RBAC + Multi-tenant + Developer Portal (basic)
  M01, M02, M07, M08 (FREE tier)
  M45 Audit Log + M47 Security (semua tier)

Phase 2 — Core BASIC (Bln 3-6)
  M04 Akademik + M05 AI Soal + M14 Notifikasi WA
  M13 Portal Ortu + M18 SPP Online + M03 Koperasi
  Developer Portal: pricing, modul toggle, listing onboarding

Phase 3 — PRO Features (Bln 6-10)
  M23 EWS + M26 AI Tutor + M24 Learning Path
  M27 UKS + M29 Anti-Bullying + M19 Kantin
  M48 RPP + M49 KKTP + M50 Remedial

Phase 4 — Ekosistem (Bln 10-16)
  M31 SDM + M32 Smart Gate + M36 Alumni + M37 Portofolio
  M40 Dapodik + M51 RKAS + M53 Komite
  Listing Marketplace: guru les, kursus, transport, catering

Phase 5 — Enterprise & Growth (Bln 16-24)
  M42 White-label + M43 Yayasan + M44 Open API
  M58 Data Warehouse + M59 IoT + M60 Gamifikasi
  M61 Marketplace Konten (revenue share)
```

---

*Untuk detail setiap domain, baca file docs/ yang sesuai. File ini hanya index dan overview.*

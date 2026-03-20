# 01 — SaaS Platform & Developer Portal

Dokumen ini menjelaskan lapisan bisnis EDS: Developer Portal untuk manajemen tenant & fitur, sistem pricing, dan Listing Marketplace untuk vendor ekosistem sekolah.

---

## Arsitektur SaaS

```
eds.id                         ← Marketing site + sign-up
├── dev.eds.id                 ← Developer Portal (SaaS ops)
├── [slug].eds.id              ← School portal per tenant
└── api.eds.id                 ← Public API (M44)
```

Setiap sekolah yang mendaftar mendapat:
- Subdomain unik: `namaSekolah.eds.id`
- Opsi custom domain: `siakad.namaSekolah.sch.id` (tier ENTERPRISE)
- Isolasi data penuh via PostgreSQL RLS per `schoolId`
- Branding kustom: logo, warna primer, nama platform

---

## Developer Portal (`dev.eds.id`)

Developer Portal adalah panel operasional untuk **tim EDS** dan **admin yayasan/mitra resmi** — bukan untuk kepala sekolah. Di sinilah seluruh lifecycle tenant dikelola.

### Akses Developer Portal

| Role | Akses |
|------|-------|
| `EDS_SUPERADMIN` | Full access semua tenant |
| `EDS_SUPPORT` | Read semua tenant, bisa impersonate untuk support |
| `EDS_SALES` | Onboarding sekolah baru, lihat billing |
| `YAYASAN_ADMIN` | Kelola tenant milik yayasannya saja (M43) |
| `LISTING_MANAGER` | Kelola vendor listing (approve, suspend) |

---

### Fitur Developer Portal

#### 1. Tenant Management

```
Daftar Sekolah
├── Cari & filter (provinsi, kota, tier, status)
├── Detail Tenant
│   ├── Info dasar: nama, NPSN, subdomain, kontak PIC
│   ├── Status: active / suspended / trial / churned
│   ├── Tier saat ini + tanggal renewal
│   ├── Usage stats: siswa aktif, guru, storage, WA quota
│   └── Riwayat pembayaran
├── Tambah Sekolah Baru
│   ├── Form onboarding (nama, NPSN, alamat, PIC, tier)
│   ├── Generate subdomain otomatis dari nama sekolah
│   ├── Set user utama sekolah (admin pertama)
│   └── Kirim email welcome + credentials ke PIC
└── Bulk actions: suspend, upgrade tier, extend trial
```

**Onboarding Flow Sekolah Baru**:
```
1. Tim sales isi form di Developer Portal
2. Sistem generate:
   - schoolId (CUID)
   - subdomain slug
   - admin user (email + temp password)
   - RLS policy di PostgreSQL
3. Kirim email welcome ke PIC sekolah
4. PIC login, ganti password, isi data sekolah
5. Status: trial (30 hari) → pilih tier → bayar → active
```

#### 2. Modul & Feature Flag Management

Setiap modul bisa diaktifkan/dinonaktifkan per tenant secara granular.

```typescript
interface TenantFeatureConfig {
  schoolId: string;
  modules: {
    [moduleId: string]: {
      enabled: boolean;
      tier: 'free' | 'basic' | 'pro' | 'enterprise';
      config?: {
        aiQuotaMonthly?: number;     // M05, M23, M26 — jumlah request AI/bulan
        waQuotaMonthly?: number;     // M14 — jumlah pesan WA/bulan
        storageGb?: number;          // batas storage upload
        maxStudents?: number;        // batas siswa aktif
        customDomain?: string;       // M42
        transactionFeePercent?: number; // M18 — fee payment gateway
      };
      expiresAt?: Date;              // override per-modul untuk trial fitur
    };
  };
}
```

**UI Feature Flag di Developer Portal**:
```
Sekolah: SMAN 1 Malang
Tier: PRO
────────────────────────────────────────────────
Modul         Status    Konfigurasi
────────────────────────────────────────────────
M01 SIMS      ✅ ON     —
M05 AI Soal   ✅ ON     Quota: 500 soal/bulan
M14 Notif WA  ✅ ON     Quota: 2.000 pesan/bulan
M25 Prediktif ⛔ OFF    (butuh ENTERPRISE)
M42 White-label ⛔ OFF  (butuh ENTERPRISE)
────────────────────────────────────────────────
[Override Modul]  [Reset ke Default Tier]
```

#### 3. Billing & Subscription Management

```
Billing Dashboard
├── MRR (Monthly Recurring Revenue) total
├── Per Tenant:
│   ├── Tier & harga
│   ├── Addons: WA quota extra, storage extra, AI quota extra
│   ├── Status pembayaran (lunas / jatuh tempo / gagal)
│   ├── Tanggal renewal berikutnya
│   └── Riwayat invoice
├── Upgrade / downgrade tier (efektif akhir periode)
├── Suspend otomatis jika gagal bayar > 7 hari
└── Laporan keuangan SaaS (MRR, churn, ARPU)
```

**Pricing Structure**:

```
┌─────────────────────────────────────────────────────────────┐
│  TIER FREE — Rp 0/bulan                                     │
│  Modul: M01, M02, M07, M08, M22, M29, M37, M45, M46, M47  │
│  Limit: maks 100 siswa, 1 GB storage                       │
├─────────────────────────────────────────────────────────────┤
│  TIER BASIC — Rp 299.000/bulan                              │
│  Semua FREE + M03, M04, M06, M09, M12, M13, M15, M17,     │
│  M20, M21, M27, M28, M30, M31, M33, M35, M36, M39,        │
│  M40, M48, M50, M51, M52, M53, M57                         │
│  Limit: maks 500 siswa, 10 GB storage                      │
├─────────────────────────────────────────────────────────────┤
│  TIER PRO — Rp 799.000/bulan                                │
│  Semua BASIC + M05 (500 soal AI/bln), M10, M14 (1000 WA), │
│  M16, M18, M19, M23, M24, M26, M26b, M32, M38, M41,       │
│  M49, M55, M60                                              │
│  Limit: maks 2.000 siswa, 50 GB storage                    │
├─────────────────────────────────────────────────────────────┤
│  TIER ENTERPRISE — Custom pricing                           │
│  Semua PRO + M25, M42, M43, M44, M58, M59                  │
│  Unlimited siswa, custom domain, SLA 99.9%, dedicated DB   │
└─────────────────────────────────────────────────────────────┘

Add-ons (bisa dibeli terpisah):
  +500 soal AI/bulan      → Rp  50.000
  +1.000 WA message/bulan → Rp  35.000
  +10 GB storage          → Rp  25.000
  Custom domain           → Rp 100.000/bulan
```

#### 4. User Utama Sekolah (Admin Seeding)

Saat mendaftarkan sekolah baru, Developer Portal men-generate user utama:

```typescript
interface SchoolAdminSeed {
  schoolId: string;
  adminUser: {
    name: string;
    email: string;          // email PIC sekolah
    role: 'ADMIN_SEKOLAH';
    tempPassword: string;   // auto-generate, harus ganti saat login pertama
  };
  welcomeEmail: {
    to: string;
    schoolName: string;
    subdomain: string;
    loginUrl: string;
    tempPassword: string;
    expiresAt: Date;        // temp password expire 7 hari
  };
}
```

Admin sekolah pertama ini kemudian bisa:
- Menambah user lain (guru, TU, dll) dari dalam portal sekolah
- Mengatur konfigurasi sekolah (logo, nama, NPSN)
- Mengaktifkan modul yang sudah di-enable dari Developer Portal

#### 5. Impersonation (Support Tool)

Tim support EDS bisa masuk ke akun sekolah tanpa tahu password mereka, untuk membantu troubleshooting.

```
[Impersonate Tenant] → pilih sekolah → pilih user → buka portal sekolah
Semua aksi saat impersonate dicatat di audit log (M45) dengan flag `impersonated_by`
Impersonation session timeout: 30 menit
Notifikasi email otomatis ke admin sekolah: "Tim support EDS mengakses portal Anda"
```

#### 6. Health & Monitoring Dashboard

```
System Health
├── API response time (p50, p95, p99)
├── Database connections per tenant
├── AI API usage: token/hari, cost estimate
├── WA delivery rate per tenant
├── Error rate per modul
├── Active users realtime
└── Upcoming renewals (next 30 hari)
```

---

## Listing Marketplace

Listing adalah fitur di mana penyedia layanan (vendor) membayar EDS untuk tampil di halaman sekolah yang relevan. Ini adalah **revenue stream terpisah** dari subscription sekolah.

### Konsep

```
Vendor (guru les, katering, dll)
    → Daftar & bayar listing di Developer Portal
    → Profil vendor tampil di portal sekolah yang dipilih
    → Orang tua & siswa bisa lihat & menghubungi vendor
    → Sekolah mendapat komisi atau listing gratis (tergantung paket)
```

### Kategori Listing

| Kategori | Modul Terkait | Model Bisnis |
|----------|--------------|--------------|
| Guru Les & Bimbel | M11 | Listing fee + transaksi booking |
| Tempat Kursus & Lembaga | M11 | Listing fee bulanan |
| Jasa Antar-Jemput | M34 | Listing fee + verifikasi armada |
| Katering & Kantin | M19 | Listing fee + % transaksi |
| Penerbit & Toko Buku | M10 | Listing fee + % penjualan |
| Seragam & Atribut | M56 | Listing fee + % penjualan |
| Vendor Event (EO, foto) | M57 | Listing fee per event |
| Asuransi Pendidikan | — | CPA (cost per acquisition) |
| Tabungan & Investasi | — | CPA / lead fee |
| Supplier Alat Tulis | M03 (koperasi) | Listing fee |

### Pricing Listing Vendor

```
LISTING BASIC      → Rp 199.000/bulan
  Tampil di 1 kota, tanpa prioritas, profil standar

LISTING PRO        → Rp 499.000/bulan
  Tampil di 1 provinsi, prioritas listing, foto galeri,
  verifikasi badge, fitur booking langsung

LISTING PREMIUM    → Rp 999.000/bulan
  Nasional, prioritas tertinggi, featured placement,
  integrasi penuh (booking, pembayaran in-app, review)

LISTING PER EVENT  → Rp 99.000/event (vendor event)
  Tampil di halaman event sekolah tertentu
```

### Alur Vendor di Developer Portal

```
1. Vendor daftar di dev.eds.id/listing
2. Isi profil: nama usaha, kategori, deskripsi, foto, harga, area
3. Upload dokumen verifikasi (KTP/NPWP/izin usaha)
4. Tim EDS review & approve (SLA: 2 hari kerja)
5. Vendor pilih paket listing & bayar
6. Listing aktif → muncul di portal sekolah yang sesuai area
7. Vendor kelola: update profil, lihat statistik, tanggapi review
8. Auto-suspend jika bermasalah atau tidak bayar
```

### Manajemen Listing di Developer Portal

```
Listing Management
├── Daftar vendor (filter: kategori, area, status, tier)
├── Queue approval (pending review)
│   ├── Lihat dokumen upload
│   ├── Approve / Reject + alasan
│   └── Request dokumen tambahan
├── Statistik listing
│   ├── Revenue listing per kategori
│   ├── Click-through rate per vendor
│   ├── Booking conversion rate
│   └── Review & rating rekapitulasi
├── Moderasi konten & review
└── Bulk actions: extend, suspend, featured toggle
```

### Tampilan Listing di Portal Sekolah

Listing vendor muncul di konteks yang relevan:
- **M11** (Guru Les): tab "Guru Les Terdekat" di portal siswa
- **M34** (Transportasi): tab "Jasa Antar-Jemput" di portal ortu
- **M19** (Kantin): tab "Katering Partner" di halaman kantin
- **M08** (PPDB): sidebar "Persiapkan Sekolah Barumu" (seragam, alat tulis, asuransi)
- **M57** (Event): sidebar vendor event saat panitia merencanakan kegiatan

Sekolah bisa:
- Memilih kategori listing mana yang mau ditampilkan (opt-in per kategori)
- Mendapat komisi jika ada transaksi dari listing (jika tier PRO)
- Menyembunyikan listing tertentu jika tidak relevan

---

## Database Schema: SaaS Layer

```prisma
model Tenant {
  id            String        @id @default(cuid())
  schoolId      String        @unique  // FK ke School
  subdomain     String        @unique
  customDomain  String?       @unique
  tier          TenantTier    @default(FREE)
  status        TenantStatus  @default(TRIAL)
  trialEndsAt   DateTime?
  billingEmail  String
  features      TenantFeature[]
  subscription  Subscription?
  addons        TenantAddon[]
  createdAt     DateTime      @default(now())
  updatedAt     DateTime      @updatedAt
}

model TenantFeature {
  id          String   @id @default(cuid())
  tenantId    String
  tenant      Tenant   @relation(fields: [tenantId], references: [id])
  moduleId    String   // 'M05', 'M14', dll
  enabled     Boolean  @default(false)
  config      Json?    // quota, limits, dll
  overrideTier TenantTier?  // override manual oleh EDS support
  expiresAt   DateTime?
  updatedBy   String   // userId EDS support yang mengubah
  updatedAt   DateTime @updatedAt
  @@unique([tenantId, moduleId])
}

model Subscription {
  id              String             @id @default(cuid())
  tenantId        String             @unique
  tenant          Tenant             @relation(fields: [tenantId], references: [id])
  stripeSubId     String?            @unique
  tier            TenantTier
  status          SubscriptionStatus
  currentPeriodStart DateTime
  currentPeriodEnd   DateTime
  cancelAtPeriodEnd  Boolean         @default(false)
  invoices        Invoice[]
  createdAt       DateTime           @default(now())
}

model ListingVendor {
  id            String          @id @default(cuid())
  businessName  String
  category      ListingCategory
  description   String
  photos        String[]
  contactPhone  String
  contactEmail  String
  website       String?
  coverageArea  String[]        // kota/kab yang dilayani
  tier          ListingTier     @default(BASIC)
  status        ListingStatus   @default(PENDING_REVIEW)
  verifiedAt    DateTime?
  verifiedBy    String?
  listings      ListingPlacement[]
  reviews       ListingReview[]
  billing       ListingBilling?
  createdAt     DateTime        @default(now())
}

model ListingPlacement {
  id          String        @id @default(cuid())
  vendorId    String
  vendor      ListingVendor @relation(fields: [vendorId], references: [id])
  schoolId    String?       // null = semua sekolah di area coverage
  moduleId    String        // modul mana listing ini muncul
  position    Int           @default(0)  // untuk featured/priority
  isActive    Boolean       @default(true)
  clickCount  Int           @default(0)
  bookingCount Int          @default(0)
  createdAt   DateTime      @default(now())
}

enum TenantTier     { FREE BASIC PRO ENTERPRISE }
enum TenantStatus   { TRIAL ACTIVE SUSPENDED CHURNED }
enum SubscriptionStatus { ACTIVE PAST_DUE CANCELED UNPAID }
enum ListingCategory {
  GURU_LES TEMPAT_KURSUS ANTAR_JEMPUT KATERING
  PENERBIT SERAGAM VENDOR_EVENT ASURANSI TABUNGAN ALAT_TULIS
}
enum ListingTier    { BASIC PRO PREMIUM PER_EVENT }
enum ListingStatus  { PENDING_REVIEW ACTIVE SUSPENDED EXPIRED REJECTED }
```

---

## API Endpoints: Developer Portal

```
POST   /dev/tenants                    — Tambah sekolah baru
GET    /dev/tenants                    — List semua tenant
GET    /dev/tenants/:id                — Detail tenant
PATCH  /dev/tenants/:id/features       — Update feature flags
POST   /dev/tenants/:id/impersonate    — Impersonation (support)
POST   /dev/tenants/:id/suspend        — Suspend tenant
POST   /dev/tenants/:id/unsuspend      — Aktifkan kembali

GET    /dev/billing/dashboard          — MRR, churn, overview
GET    /dev/billing/:tenantId          — Detail billing tenant
POST   /dev/billing/:tenantId/override — Override tier manual

POST   /dev/listings/vendors           — Daftar vendor baru
GET    /dev/listings/vendors           — List vendor
PATCH  /dev/listings/vendors/:id/approve — Approve vendor
PATCH  /dev/listings/vendors/:id/reject  — Reject vendor
GET    /dev/listings/analytics         — Statistik listing

GET    /dev/health                     — System health
GET    /dev/metrics                    — Usage metrics
GET    /dev/audit                      — Audit log Developer Portal
```

---

## Webhook: Event-Driven Billing

```typescript
// Events yang dikirim ke billing system
type BillingEvent =
  | { type: 'tenant.created';          tenantId: string; tier: TenantTier }
  | { type: 'tenant.upgraded';         tenantId: string; from: TenantTier; to: TenantTier }
  | { type: 'tenant.payment_failed';   tenantId: string; daysOverdue: number }
  | { type: 'tenant.suspended';        tenantId: string }
  | { type: 'addon.quota_exceeded';    tenantId: string; addon: string }
  | { type: 'listing.payment_success'; vendorId: string; amount: number }
  | { type: 'listing.expired';         vendorId: string; listingId: string }
```

---

## Integrasi dengan Modul Sekolah

Feature flag di Developer Portal langsung dikonsumsi oleh middleware API:

```typescript
// packages/api/middleware/feature-gate.ts
export async function featureGate(moduleId: string) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const { schoolId } = req.user;
    const feature = await getFeatureConfig(schoolId, moduleId);

    if (!feature?.enabled) {
      return res.status(403).json({
        success: false,
        error: {
          code: 'FEATURE_DISABLED',
          message: `Modul ${moduleId} tidak aktif untuk sekolah ini.`,
          upgradeUrl: `https://eds.id/upgrade?school=${schoolId}`
        }
      });
    }

    // Inject quota info ke request
    req.featureConfig = feature.config;
    next();
  };
}

// Penggunaan di route
app.post('/api/v1/exam/generate',
  authenticate,
  featureGate('M05'),          // cek apakah M05 aktif
  checkAIQuota('M05'),         // cek sisa quota soal AI bulan ini
  generateExamHandler
);
```

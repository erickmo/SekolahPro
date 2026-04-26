# ADR-K036: Mobile App Strategy

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

K021 (E-Wallet) dan K023 (Dashboard Portal) menyediakan web-based interfaces. Namun di Indonesia, **mobile adalah channel utama** — terutama untuk:

- Orang tua yang ingin top-up uang saku anak
- Siswa yang menggunakan e-wallet di kantin
- Guru/staff yang cek saldo dan angsuran
- Nasabah yang ingin cek riwayat transaksi

Saat ini belum ada ADR yang mengatur:
- Arsitektur mobile app (native vs hybrid vs PWA)
- Fitur offline untuk area dengan koneksi terbatas
- Keamanan mobile (biometric, certificate pinning)
- Distribusi app (store, APK, school-specific)

## Decision

### 1. Platform Decision: Flutter Hybrid

```
mobile_architecture:
├── FRAMEWORK: Flutter
│   ├── Single codebase → iOS + Android
│   ├── Performance mendekati native
│   ├── Rich UI components
│   └── Team sudah familiar (ref flutter-project-init skill)
│
├── BACKEND: Existing Go API (vernon-api pattern)
│   ├── REST API yang sama dengan web dashboard
│   ├── JWT token authentication
│   └── Vernon read cache untuk performance
│
├── STATE MANAGEMENT: Riverpod + Flutter Hooks
│   ├── Consistent dengan web dashboard stack
│   └── Offline-first architecture
│
└── LOCAL STORAGE: Hive / Isar
    ├── Encrypted local database
    ├── Offline transaction queue
    └── Biometric credential storage
```

### 2. App Variants

```
app_variants:
├── NASABAH_APP (Member App)
│   ├── Untuk: anggota koperasi, guru, staff
│   ├── Fitur:
│   │   ├── Cek saldo & riwayat transaksi
│   │   ├── Pembayaran angsuran
│   │   ├── Pengajuan pinjaman
│   │   ├── Notifikasi jatuh tempo
│   │   ├── SHU statement
│   │   └── Profil & pengaturan
│   └── Distribusi: Play Store / App Store
│
├── PARENT_APP (Parent Portal App)
│   ├── Untuk: orang tua siswa
│   ├── Fitur:
│   │   ├── Top-up uang saku digital (K021)
│   │   ├── Monitor spending anak
│   │   ├── Set spending limits
│   │   ├── Notifikasi transaksi anak
│   │   └── Pembayaran SPP (link ke S009)
│   └── Distribusi: Play Store / App Store
│
├── STUDENT_EWALLET (Student E-Wallet)
│   ├── Untuk: siswa (pembayaran di kantin/toko)
│   ├── Fitur:
│   │   ├── QR Code payment
│   │   ├── NFC tap-to-pay (jika hardware support)
│   │   ├── Cek saldo
│   │   └── Riwayat transaksi
│   └── Distribusi: School APK sideload (tidak public store)
│
└── COLLECTOR_APP (Field Collection App)
    ├── Untuk: petugas penagihan (ref K033)
    ├── Fitur:
    │   ├── Daftar nasabah telat bayar
    │   ├── GPS tracking kunjungan
    │   ├── Photo proof kunjungan
    │   ├── Log collection activity
    │   └── Offline mode untuk area tanpa signal
    └── Distribusi: Internal APK sideload
```

### 3. Offline-First Architecture

```
offline_strategy:
├── READ OPERATIONS
│   ├── Last-known data cached locally (Hive/Isar)
│   ├── Stale-while-revalidate pattern
│   ├── Cache TTL: saldo 5 menit, riwayat 1 jam, profil 24 jam
│   └── Clear visual indicator when showing stale data
│
├── WRITE OPERATIONS (transaction queue)
│   ├── Queue system untuk transaksi offline
│   ├── Auto-retry saat koneksi恢复
│   ├── Max queue size: 10 transaksi
│   ├── Max offline period: 24 jam (setelah itu, sync required)
│   └── Conflict resolution: server wins (last-write-wins for saldo)
│
├── SYNC PROTOCOL
│   ├── Incremental sync (only changed data since last sync)
│   ├── Background sync when app is in foreground
│   ├── Manual pull-to-refresh
│   └── Sync status indicator in UI
│
└── SECURITY
    ├── Queued transactions encrypted at rest
    ├── Cannot queue high-value transactions offline (> threshold)
    ├── Biometric required before viewing cached financial data
    └── Auto-logout after 5 minutes inactivity (configurable)
```

### 4. Security Requirements

```
mobile_security:
├── AUTHENTICATION
│   ├── JWT access token (15 min) + refresh token (7 days)
│   ├── Biometric login (fingerprint/face) after first login
│   ├── PIN as fallback (6-digit)
│   └── Device registration (max 2 devices per nasabah)
│
├── NETWORK
│   ├── Certificate pinning (prevent MITM)
│   ├── TLS 1.3 minimum
│   ├── API request signing (HMAC)
│   └── No sensitive data in URL parameters
│
├── DATA AT REST
│   ├── Encrypted local storage (AES-256)
│   ├── No caching of sensitive data (password, PIN, full KTP)
│   ├── Screenshots disabled on sensitive screens (Android)
│   └── Clear cache on logout
│
├── APP INTEGRITY
│   ├── Root/jailbreak detection → warn + limited features
│   ├── App tampering detection
│   ├── Obfuscation (ProGuard/R8)
│   └── Play Integrity API / Device Check
│
└── PUSH NOTIFICATIONS
    ├── FCM (Firebase Cloud Messaging)
    ├── No sensitive data in push payload (only notification ID)
    ├── App fetches details securely after receiving push
    └── User can disable push per notification type
```

### 5. App Configuration per Tenant

```
mobile_app_config:
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Branding ──
├── app_name              VARCHAR         ← Nama app di home screen
├── primary_color         VARCHAR         ← Hex color
├── logo_url              VARCHAR         ← App logo
├── splash_screen_url     VARCHAR (nullable)
│
├── ── Feature Flags ──
├── features_enabled      JSONB
│   ├── ewallet_enabled      BOOLEAN
│   ├── loan_application     BOOLEAN
│   ├── installment_payment  BOOLEAN
│   ├── qr_payment           BOOLEAN
│   ├── nfc_payment          BOOLEAN
│   ├── biometric_login      BOOLEAN
│   └── push_notifications   BOOLEAN
│
├── ── Limits ──
├── max_offline_transactions INT DEFAULT 10
├── offline_period_hours     INT DEFAULT 24
├── session_timeout_minutes  INT DEFAULT 5
├── max_devices_per_user     INT DEFAULT 2
│
├── ─── Audit ──
├── created_at            TIMESTAMPTZ
└── updated_at            TIMESTAMPTZ
```

### 6. Dual-Mode Differences

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|---|---|---|
| Terminology | Bunga, denda | Margin/bagi hasil, ta'zir |
| Product names | Tabungan, Deposito | Tabungan, Deposito Syariah |
| SHU display | SHU | SHU |
| Color scheme | Tenant branding | + Islamic design elements (opsional) |

### 7. Authorization — RBAC

| Permission | Nasabah | Parent | Student | Collector |
|---|---|---|---|---|
| View own balance | v | v (child) | v (own) | - |
| Transfer/top-up | v | v | - | - |
| QR payment | - | - | v | - |
| Apply loan | v | - | - | - |
| Pay installment | v | - | - | - |
| View collection cases | - | - | - | v |
| Log collection activity | - | - | - | v |
| Set spending limits | - | v | - | - |
| Receive notifications | v | v | - | - |

### 8. Consequences

**Keuntungan:**
- Mobile-first untuk market Indonesia
- Single codebase (Flutter) mengurangi development cost
- Offline support untuk area terpencil
- Tenant branding yang customizable

**Risiko:**
- Flutter app size relatif besar (~15-20MB minimum)
- App store approval process bisa lambat
- Offline sync conflict bisa terjadi

**Mitigasi:**
- Deferred components untuk reduce initial app size
- Fastlane untuk automated store deployment
- Clear conflict resolution policy (server wins)

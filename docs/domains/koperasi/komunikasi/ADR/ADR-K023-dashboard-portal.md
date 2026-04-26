# ADR-K023: Dashboard & Self-Service Portal

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Sistem koperasi sekolah memiliki banyak stakeholder dengan kebutuhan informasi yang berbeda-beda:

- **Admin/Manager** membutuhkan dashboard eksekutif: overview aset, rasio keuangan, dan performa cabang
- **Supervisor** membutuhkan monitoring operasional cabang: approval queue, angsuran jatuh tempo, posisi kas
- **Teller** membutuhkan workspace transaksi: sesi aktif, quick actions, pencarian nasabah
- **Nasabah/Anggota** membutuhkan self-service portal: lihat saldo, riwayat transaksi, download statement, ajukan rekening/pinjaman
- **Orang Tua/Wali** membutuhkan parent portal: monitoring belanja anak, top-up, spending limit, analytics
- **Kepala Sekolah** membutuhkan oversight dashboard: kesehatan koperasi, statistik keanggotaan, compliance

Saat ini semua informasi harus diakses melalui menu navigasi manual — tidak ada landing page yang menampilkan ringkasan data sesuai role. Nasabah juga harus datang ke cabang untuk cek saldo atau download statement.

Dashboard harus mengakomodasi **dual-mode** terminologi (koperasi konvensional vs BMT) dan **multi-tenant** — setiap tenant bisa mengkonfigurasi widget mana yang aktif per role.

## Decision

### 1. Dashboard per Role — Widget Breakdown

Setiap role memiliki dashboard default dengan widget yang relevan. Widget bisa dikonfigurasi (enable/disable) per tenant.

**ADMIN / MANAGER (Executive Dashboard):**

```
executive_dashboard
├── Total aset koperasi (real-time aggregate)
├── Total simpanan seluruh nasabah
├── Total pinjaman outstanding
├── NPL ratio, CAR, liquidity ratio (traffic light: hijau/kuning/merah)
├── Member growth (anggota baru vs keluar — bulan/kuartal/tahun)
├── Pending approvals count (nasabah, rekening, pinjaman)
├── Cash position per branch (real-time dari K014)
├── SHU projection (estimasi tahun berjalan dari K016)
├── Regulatory report status (submitted / overdue)
├── Top 10 debtors by outstanding amount
└── Branch performance comparison (chart)
```

**SUPERVISOR (Branch Operations Dashboard):**

```
branch_operations_dashboard
├── Total simpanan & pinjaman cabang
├── Active teller sessions (dari K012)
├── Pending approvals in queue (nasabah, rekening, pinjaman)
├── Overdue installments — angsuran jatuh tempo (dari K008)
├── Dormant accounts list (dari K002)
├── Daily transaction summary (jumlah & nominal)
├── Simpanan wajib collection rate (% anggota yang sudah bayar bulan ini)
└── Cash position cabang (dari K014)
```

**TELLER (Transaction Workspace):**

```
teller_workspace
├── Active session info:
│   ├── Kas awal (opening balance)
│   ├── Running total (kas saat ini)
│   └── Expected cash (kas awal + setoran - penarikan)
├── Today's transactions list (ringkasan per transaksi)
├── Quick actions:
│   ├── Setoran (link ke form setoran)
│   ├── Penarikan (link ke form penarikan)
│   ├── Angsuran (link ke form angsuran)
│   └── Buka Rekening (link ke form pengajuan)
├── Nasabah quick search:
│   ├── By name (partial match)
│   ├── By member number
│   └── By identity number (KTP/kartu pelajar)
└── Pending cash count (jika sesi mendekati closing)
```

**NASABAH / ANGGOTA (Self-Service Portal):**

```
nasabah_portal
├── My accounts summary (semua rekening dengan saldo)
├── Recent transactions (10 transaksi terakhir, semua rekening)
├── My loan status:
│   ├── Outstanding pokok
│   ├── Tanggal angsuran berikutnya
│   └── Nominal angsuran berikutnya
├── My SHU history (distribusi tahunan dari K016)
├── Transaction history (search & filter per rekening, per periode)
├── Statement download (PDF, per rekening, per periode)
├── Profile view (read-only, request update via form)
├── Pengajuan rekening baru (submit application — K002)
└── Pengajuan pinjaman baru (submit application — K007)
```

**ORANG TUA / WALI (Parent Portal):**

```
parent_portal
├── Child's account summary (semua rekening anak yang di-link)
├── Child's spending today / this week / this month
├── Top-up child's tabungan (creates setoran transaction — K011)
├── Set/update spending limits:
│   ├── Daily maximum
│   ├── Per-transaction maximum
│   └── Category restrictions (kantin, toko, dll)
├── Spending analytics:
│   ├── Pie chart by category (kantin, toko, transfer, dll)
│   └── Trend by week/month (line chart)
├── Child's savings progress (goal tracking)
└── Notification preferences for child's account (K022)
```

**KEPALA SEKOLAH (Oversight Dashboard):**

```
oversight_dashboard
├── Koperasi health summary (key ratios: NPL, CAR, liquidity)
├── Membership statistics by category:
│   ├── Guru/Ustadz
│   ├── Siswa/Santri
│   ├── Orang Tua/Wali
│   └── Staff
├── Total simpanan & pinjaman trend (chart — 12 bulan terakhir)
├── NPL status summary
├── SHU projection for next RAT (dari K016)
├── Regulatory compliance status (submitted / pending / overdue)
└── Comparison with previous year (YoY growth)
```

### 2. Widget Architecture

Dashboard menggunakan arsitektur **widget-based composable** — setiap widget adalah unit independen yang bisa dikonfigurasi per role per tenant.

```
Widget Lifecycle:
┌─────────────────────────┐
│    Widget Catalog        │  Predefined oleh sistem
│    (dashboard_widget)    │
└────────┬────────────────┘
         │ Tenant enable/disable
         v
┌─────────────────────────┐
│    Dashboard Layout      │  Per role, per tenant
│    (dashboard_layout)    │
└────────┬────────────────┘
         │ User opens dashboard
         v
┌─────────────────────────┐
│    Widget Render         │  Fetch data → render component
│    - Data source query   │
│    - Refresh interval    │
│    - Cache (if enabled)  │
└─────────────────────────┘
```

**Aturan widget:**
- Setiap widget memiliki **data_source** independen — query atau API endpoint yang mengembalikan data untuk widget tersebut
- Admin tenant bisa **enable/disable** widget per role — menggunakan `dashboard_layout`
- Widget catalog bersifat **system-defined** — tenant tidak bisa membuat widget baru, hanya enable/disable
- Widget refresh interval configurable: 30 detik, 1 menit, 5 menit, atau manual refresh
- Widget yang gagal fetch data menampilkan **error state** tanpa mempengaruhi widget lain (fault isolation)
- Widget mendukung **loading state** — skeleton screen saat data sedang di-fetch

### 3. Self-Service Portal Features

Nasabah dapat mengakses portal tanpa harus datang ke cabang:

```
Self-Service Capabilities:
├── VIEW (read-only):
│   ├── Saldo semua rekening
│   ├── Riwayat transaksi (search, filter, paginate)
│   ├── Status pinjaman (outstanding, jadwal angsuran)
│   ├── History SHU yang diterima
│   └── Profile data pribadi
│
├── DOWNLOAD:
│   ├── Statement per rekening per periode (PDF)
│   ├── Transaction history export (CSV)
│   └── SHU certificate (PDF)
│
├── SUBMIT (application — requires approval):
│   ├── Pengajuan rekening baru → flow K002
│   ├── Pengajuan pinjaman baru → flow K007
│   └── Profile update request → approval oleh Teller/Supervisor
│
└── CANNOT DO (by design):
    ├── Edit saldo atau transaksi
    ├── Edit data pribadi langsung (harus melalui request)
    ├── Approve/reject apapun
    └── Lihat data nasabah lain
```

**Aturan:**
- Nasabah **hanya bisa melihat data milik sendiri** — strict row-level security
- Profile update bukan direct edit — nasabah submit **request** yang harus di-approve oleh Teller/Supervisor (data integrity)
- Pengajuan rekening/pinjaman via portal mengikuti flow yang sama dengan pengajuan via Teller (K002, K007) — hanya channel yang berbeda
- Statement download di-generate on-demand — format profesional dengan header/logo koperasi

### 4. Parent Portal Features

Orang tua/wali dapat memantau dan mengelola akun anak:

```
Parent-Child Linking:
┌──────────────────┐     ┌──────────────────┐
│  Nasabah Parent  │────>│  Nasabah Child    │
│  (orang_tua)     │     │  (siswa/santri)   │
│                  │     │                   │
│  school_relation │     │  school_relation  │
│  = parent        │     │  = student        │
└──────────────────┘     └──────────────────┘
         │
         │ Link via:
         ├── school_entity_id (otomatis dari data sekolah)
         └── Manual link (Admin assign parent ↔ child)
```

**Fitur parent portal:**

| Fitur | Deskripsi | Trigger Notifikasi (K022) |
|-------|-----------|---------------------------|
| View child accounts | Lihat saldo semua rekening anak | - |
| Spending monitoring | Belanja hari ini / minggu / bulan | Daily summary (configurable) |
| Remote top-up | Setoran ke tabungan anak | Top-up confirmation |
| Daily spending limit | Max belanja per hari | Alert jika limit exceeded |
| Per-transaction limit | Max per transaksi | Alert jika limit exceeded |
| Category restriction | Blokir kategori tertentu (misal: toko) | Alert jika restricted category |
| Spending analytics | Pie chart kategori, trend mingguan/bulanan | - |
| Savings goal | Target menabung anak (progress bar) | Goal reached notification |

**Aturan:**
- Parent hanya bisa melihat data anak yang **terhubung** — tidak bisa lihat anak nasabah lain
- Remote top-up menghasilkan **transaksi setoran** di rekening tabungan anak (flow K011)
- Spending limit disimpan di `spending_control` ([ADR-K021](./ADR-K021-ewallet.md)) — di-enforce di transaction layer (K011, K019, K021)
- Link parent-child bisa **1-to-many** — satu orang tua bisa punya beberapa anak
- Link juga bisa **many-to-1** — satu anak bisa punya ayah dan ibu sebagai wali
- Analytics menggunakan data dari transaksi anak — aggregation di backend, bukan di client

### 5. Data Access Control

Akses data menggunakan **strict row-level security** yang di-enforce di API layer:

```
Access Control Matrix:
┌─────────────────┬───────────────────────────────────────────┐
│ Role            │ Data Scope                                │
├─────────────────┼───────────────────────────────────────────┤
│ Nasabah         │ Own data ONLY (own rekening, transaksi)   │
│ Orang Tua       │ Own data + linked children's data ONLY    │
│ Teller          │ Own branch, all nasabah in branch         │
│ Supervisor      │ Own branch, all nasabah in branch         │
│ Manager         │ Own branch + subordinate branches         │
│ Admin           │ All data within tenant                    │
│ Kepala Sekolah  │ Aggregate/summary data only (no PII)     │
└─────────────────┴───────────────────────────────────────────┘
```

**Enforcement layers:**

```
Request → API Gateway
           │
           v
    ┌──────────────────┐
    │ 1. AuthN (JWT)   │  Verify identity (K022 pkg/jwt)
    └────────┬─────────┘
             v
    ┌──────────────────┐
    │ 2. Tenant Check  │  tenant_id dari JWT must match resource
    └────────┬─────────┘
             v
    ┌──────────────────┐
    │ 3. Role Check    │  Permission sesuai RBAC table
    └────────┬─────────┘
             v
    ┌──────────────────┐
    │ 4. Row-Level     │  Filter query by:
    │    Security      │  - nasabah_id (untuk nasabah)
    │                  │  - parent_child_link (untuk orang tua)
    │                  │  - branch_id (untuk teller/supervisor)
    │                  │  - tenant_id (untuk admin)
    └──────────────────┘
```

**Aturan:**
- Setiap API endpoint **wajib** menyertakan tenant_id filter — tidak ada endpoint yang bisa akses cross-tenant
- Nasabah endpoint menambahkan `WHERE nasabah_id = :current_user_nasabah_id` secara otomatis
- Parent endpoint menambahkan `WHERE nasabah_id IN (SELECT child_id FROM parent_child_link WHERE parent_id = :current_user_nasabah_id)`
- Kepala Sekolah hanya mendapatkan **aggregate data** — tidak bisa drill down ke data nasabah individual (privacy)
- Teller dan Supervisor dibatasi oleh `branch_id` — tidak bisa lihat data cabang lain

### 6. Mobile Responsiveness

Dashboard dan portal harus accessible di berbagai device:

```
Device Support:
├── Desktop (primary)    → Full dashboard layout, semua widget
├── Tablet               → Responsive layout, widget reflow
├── Mobile browser       → Simplified layout, critical widgets only
└── Native mobile app    → Phase 2 (out of scope MVP)
```

**Aturan:**
- Dashboard menggunakan **responsive design** — CSS Grid/Flexbox, breakpoint-based layout
- Mobile view menampilkan **critical widgets** terlebih dahulu — saldo, transaksi terakhir, quick actions
- Chart/graph menggunakan library yang mendukung responsive rendering
- Touch-friendly: button sizes minimum 44x44px, form inputs optimized untuk mobile
- **Phase 2**: native mobile app (Flutter) dengan offline capability — didokumentasikan sebagai future scope, tidak masuk MVP
- Portal nasabah dan parent portal **diprioritaskan** untuk mobile — user base utama (nasabah, orang tua) lebih sering akses via smartphone

### 7. Export & Download

Sistem menyediakan fitur export data untuk berbagai keperluan:

```
Export Capabilities:
├── CSV/Excel:
│   ├── Transaction history (per rekening, per periode)
│   ├── Member list (Admin/Manager only)
│   ├── Overdue installment list (Supervisor+)
│   └── Custom tabular data dari widget dashboard
│
├── PDF:
│   ├── Account statement (per rekening, per periode)
│   ├── SHU certificate
│   ├── Loan schedule (jadwal angsuran)
│   └── Daily/monthly summary report
│
└── Bulk Export:
    ├── Date range filter (max 1 tahun per request)
    ├── Async generation (queue-based untuk file besar)
    └── Download link via notifikasi (K022) saat selesai
```

**Statement format:**

```
┌─────────────────────────────────────────────┐
│           [Logo Koperasi/BMT]               │
│        Koperasi ABC / BMT XYZ               │
│     Jl. Raya Sekolah No. 1, Jakarta         │
│           Tel: (021) 123-4567               │
├─────────────────────────────────────────────┤
│  REKENING KORAN / STATEMENT                 │
│                                             │
│  Nasabah: Ahmad Fauzi                       │
│  No. Anggota: KOP-2026-JKT-000001          │
│  No. Rekening: TB-2026-JKT-00000001        │
│  Periode: 01/01/2026 - 31/03/2026          │
├──────┬──────────┬────────┬────────┬─────────┤
│ Tgl  │ Keterangan│ Debit  │ Kredit │ Saldo   │
├──────┼──────────┼────────┼────────┼─────────┤
│ ...  │ ...      │ ...    │ ...    │ ...     │
├──────┴──────────┴────────┴────────┴─────────┤
│  Saldo Akhir: Rp 1.500.000                 │
│  Dicetak: 15/04/2026 14:30 WIB             │
└─────────────────────────────────────────────┘
```

**Aturan:**
- Statement PDF menggunakan **header koperasi** yang configurable per tenant (logo, nama, alamat)
- Export CSV/Excel menggunakan format standar yang bisa di-import ke spreadsheet
- Bulk export untuk data besar (> 1000 row) diproses **asynchronous** — user mendapat notifikasi (K022) saat file siap download
- File export memiliki **expiry** (default 7 hari) — setelah itu auto-delete dari storage
- Export di-log ke audit trail — siapa download apa, kapan

### 8. Data Model

**dashboard_widget (Widget Catalog):**

```
dashboard_widget
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── code                  VARCHAR UNIQUE per tenant (misal: EXEC_TOTAL_ASET)
├── name                  VARCHAR NOT NULL
├── description           TEXT (nullable)
├── category              VARCHAR (misal: financial, operational, membership)
│
├── ── Data Source ──
├── data_source           VARCHAR NOT NULL (endpoint atau query identifier)
├── data_params           JSONB DEFAULT '{}' (parameter tambahan)
├── refresh_interval_sec  INTEGER DEFAULT 300 (5 menit)
│
├── ── Display ──
├── default_size          ENUM (small, medium, large, full_width)
├── chart_type            ENUM (number, table, bar, line, pie, donut, traffic_light) DEFAULT 'number'
├── display_order         INTEGER DEFAULT 0
│
├── ── Availability ──
├── target_roles          JSONB NOT NULL (list role yang bisa melihat)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**dashboard_layout (Role-based Widget Config per Tenant):**

```
dashboard_layout
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
│
├── ── Identitas ──
├── role                  VARCHAR NOT NULL (misal: admin, supervisor, teller, nasabah, parent, kepsek)
├── layout_name           VARCHAR DEFAULT 'default'
│
├── ── Widget Configuration ──
├── widgets               JSONB NOT NULL
│   │   Array of:
│   │   {
│   │     "widget_id": "018f...",
│   │     "position":  { "row": 0, "col": 0 },
│   │     "size":      "medium",
│   │     "visible":   true,
│   │     "refresh_interval_sec": 60,
│   │     "custom_params": {}
│   │   }
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**portal_session (Self-Service Login Tracking):**

```
portal_session
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── user_id               UUID (FK → user)
│
├── ── Session Info ──
├── session_type          ENUM (nasabah_portal, parent_portal)
├── device_type           VARCHAR (mobile, tablet, desktop)
├── user_agent            VARCHAR
├── ip_address            INET
│
├── ── Timing ──
├── started_at            TIMESTAMPTZ NOT NULL
├── last_active_at        TIMESTAMPTZ NOT NULL
├── ended_at              TIMESTAMPTZ (nullable)
├── duration_seconds      INTEGER (nullable, calculated on end)
│
├── ── Activity ──
├── pages_visited         JSONB DEFAULT '[]'
├── actions_performed     JSONB DEFAULT '[]'
│   │   Array of:
│   │   {
│   │     "action": "view_balance",
│   │     "timestamp": "2026-04-15T10:30:00Z",
│   │     "metadata": {}
│   │   }
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    └── created_by        UUID
```

**Parent Portal — Spending Control:**

Parent mengatur spending limit anak melalui `spending_control` (K021):
- View dan edit daily_limit, per_transaction_limit, weekly_limit, monthly_limit
- Set category_restrictions (canteen only, toko only, semua)
- Set time_restrictions (jam sekolah saja)
- Changes berlaku real-time (POS K019 selalu query latest config)
- Audit trail: setiap perubahan tercatat (updated_by = parent nasabah_id)

Spending limits TIDAK diduplikasi di sini — gunakan `spending_control` dari [ADR-K021](./ADR-K021-ewallet.md).

**parent_child_config (simplified — hanya link dan notifikasi):**

```
parent_child_config
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── parent_nasabah_id     UUID (FK → nasabah) NOT NULL
├── child_nasabah_id      UUID (FK → nasabah) NOT NULL
│
├── ── Savings Goal ──
├── savings_goal_amount   NUMERIC(15,2) (nullable)
├── savings_goal_label    VARCHAR (nullable, misal: "Tabungan Liburan")
├── savings_goal_deadline DATE (nullable)
│
├── ── Notification Config ──
├── notification_config   JSONB (preferensi notifikasi: daily_summary, per_transaction, dll)
├── is_active             BOOLEAN DEFAULT true
│
├── ── Vernon Fields ──
├── _rels                 JSONB NOT NULL DEFAULT '{}'
├── _data                 JSONB NOT NULL DEFAULT '{}'
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 9. Vernon _rels dan _data Structure

**dashboard_widget — minimal _data** karena widget catalog jarang di-list dengan JOIN:

```json
// _rels
{
  "tenant_id": "018f..."
}

// _data — minimal
{}
```

**dashboard_layout:**

```json
// _rels
{
  "tenant_id": "018f..."
}

// _data — minimal
{}
```

**parent_child_config:**

```json
// _rels
{
  "tenant_id":         "018f...",
  "parent_nasabah_id": "018f...",
  "child_nasabah_id":  "018f..."
}

// _data
{
  "parent": {
    "id":            "018f...",
    "full_name":     "Budi Santoso",
    "member_number": "KOP-2026-JKT-000010"
  },
  "child": {
    "id":            "018f...",
    "full_name":     "Andi Santoso",
    "member_number": "KOP-2026-JKT-000042"
  }
}
```

**SyncEngine triggers:**
- `NasabahUpdatedEvent` → update `_data.parent` atau `_data.child` di `parent_child_config` yang terkait

**portal_session — tidak menggunakan Vernon pattern** karena:
- Session write-heavy, read by query (bukan listing)
- Tidak perlu denormalized read-cache
- Query by: tenant_id, user_id, session_type, date range

### 10. Authorization — RBAC

| Permission | Nasabah | Orang Tua | Teller | Supervisor | Manager | Admin | Kepsek |
|------------|---------|-----------|--------|------------|---------|-------|--------|
| View own dashboard | v | v | v | v | v | v | v |
| View own portal (nasabah) | v | - | - | - | - | - | - |
| View parent portal | - | v | - | - | - | - | - |
| View executive dashboard | - | - | - | - | v | v | - |
| View oversight dashboard | - | - | - | - | - | - | v |
| Download own statement | v | - | - | - | - | - | - |
| Download child statement | - | v | - | - | - | - | - |
| Export data (CSV/Excel) | - | - | - | v | v | v | - |
| Configure widget layout | - | - | - | - | v | v | - |
| Manage widget catalog | - | - | - | - | - | v | - |
| Set child spending limit | - | v | - | - | - | - | - |
| Remote top-up child | - | v | - | - | - | - | - |
| Submit rekening application | v | - | v | v | v | v | - |
| Submit pinjaman application | v | - | v | v | v | v | - |
| Submit profile update request | v | - | - | - | - | - | - |
| View portal session log | - | - | - | - | v | v | - |

**Catatan:**
- Nasabah dan Orang Tua hanya bisa akses data sendiri / anak sendiri — **tidak ada akses cross-nasabah**
- Kepala Sekolah mendapat **aggregate summary** — tidak bisa drill down ke data individual nasabah
- Export data membutuhkan minimal **Supervisor** — mencegah data leak dari Teller/Nasabah
- Configure widget layout membutuhkan **Manager+** — layout mempengaruhi semua user dengan role tersebut di tenant
- Remote top-up oleh Orang Tua menghasilkan transaksi yang di-log sama seperti setoran biasa (audit trail K011)

### 11. Dual-Mode Terminology

Widget content dan label bervariasi per mode:

| Aspek | `coop_type = "general"` | `coop_type = "islamic"` |
|-------|------------------------|------------------------|
| Dashboard title | Dashboard Koperasi | Dashboard BMT |
| Member label | Anggota | Nasabah |
| Loan widget title | Pinjaman Outstanding | Pembiayaan Outstanding |
| Interest display | Bunga | Bagi Hasil / Margin |
| SHU label | SHU (Sisa Hasil Usaha) | SHU (Sisa Hasil Usaha) |
| Statement header | Koperasi [Nama] | BMT [Nama] |
| Greeting (portal) | Selamat datang, Anggota | Assalamu'alaikum, Nasabah |
| Report label | Laporan Koperasi | Laporan BMT |

**Aturan:**
- Terminologi di-resolve di **render layer** berdasarkan `coop_type` tenant — backend mengirim data netral, frontend menampilkan label sesuai mode
- Widget `data_source` sama untuk kedua mode — hanya label/display yang berbeda
- Statement PDF menggunakan template yang berbeda per `coop_type` (header, salam, terminologi)

## Consequences

### Positif

- **Role-appropriate** — setiap stakeholder mendapat informasi yang relevan tanpa noise
- **Self-service** — nasabah bisa cek saldo, download statement, dan ajukan rekening/pinjaman tanpa datang ke cabang
- **Parent empowerment** — orang tua bisa memantau dan mengelola keuangan anak secara real-time
- **Configurable** — widget-based architecture memungkinkan tenant mengkustomisasi dashboard per role
- **Secure** — strict row-level security mencegah akses data cross-nasabah
- **Mobile-ready** — responsive design memastikan portal accessible di smartphone
- **Auditable** — portal session tracking mencatat aktivitas nasabah untuk audit dan analytics
- **Dual-mode compliant** — terminologi otomatis menyesuaikan koperasi konvensional vs BMT
- **Export capable** — statement dan report bisa di-download dalam format profesional

### Negatif

- **Widget maintenance** — setiap widget memerlukan data source query yang harus di-maintain dan di-optimize
- **Performance concern** — executive dashboard aggregating data dari banyak tabel bisa lambat di tenant besar
- **Session tracking overhead** — pencatatan aktivitas portal menambah write load
- **Parent-child complexity** — linking, spending limit, dan analytics menambah kompleksitas domain
- **Mobile limitation** — responsive web bukan pengganti native app untuk UX optimal
- **Export security** — file yang di-download keluar dari kontrol sistem (data leak risk)

### Mitigasi

- Widget data source menggunakan **materialized view atau cache** untuk aggregate data — refresh periodik, bukan real-time query
- Executive dashboard widget menggunakan **pre-computed metrics** yang di-update oleh background job (misal: NPL ratio dihitung setiap jam, bukan setiap request)
- Portal session tracking bersifat **fire-and-forget** — async write, tidak blocking user experience
- Parent-child config reuse pattern **nasabah_beneficiary** (K001) untuk linking — proven pattern
- Mobile native app direncanakan di Phase 2 — responsive web cukup untuk MVP
- Export file di-encrypt dan memiliki **expiry 7 hari** — auto-delete dari storage setelah expired
- Export di-log ke audit trail — compliance team bisa review siapa download apa

## Alternatives Considered

### A. Static Dashboard (Hardcoded per Role)

Dashboard dengan layout dan widget yang di-hardcode per role, tanpa konfigurasi per tenant.

**Ditolak** karena: setiap koperasi memiliki prioritas dan kebutuhan informasi yang berbeda. Koperasi kecil mungkin tidak perlu branch comparison, sementara koperasi besar membutuhkan multi-branch view. Widget-based configurable dashboard lebih fleksibel tanpa menambah kompleksitas development yang signifikan.

### B. Separate App untuk Self-Service Portal

Membangun aplikasi terpisah (separate codebase, separate deployment) untuk portal nasabah dan parent portal.

**Ditolak** karena: menambah complexity deployment dan maintenance. Shared authentication, shared API, dan shared data model membuat integrasi dalam satu aplikasi lebih efisien. Portal nasabah cukup sebagai section/module dalam aplikasi utama dengan route dan permission yang berbeda.

### C. Native Mobile App dari Awal (Phase 1)

Membangun native mobile app (Flutter) bersamaan dengan web dashboard di Phase 1.

**Ditolak** karena: scope MVP sudah cukup besar dengan web responsive. Native mobile app membutuhkan effort tambahan yang signifikan (offline sync, push notification native, app store submission). Responsive web portal sudah cukup untuk kebutuhan nasabah dan orang tua di Phase 1. Native app direncanakan di Phase 2 setelah web portal proven dan feedback dikumpulkan.

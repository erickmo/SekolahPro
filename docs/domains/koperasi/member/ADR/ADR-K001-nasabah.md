# ADR-K001: Nasabah (Anggota / Member)

Status:     Accepted
Date:       2026-04-15
Deciders:   Erick Mo

## Context

Nasabah adalah core entity dalam sistem koperasi/BMT. Setiap transaksi — tabungan, pinjaman, angsuran — terikat pada nasabah. Sistem harus mengakomodasi:

- **Anggota internal sekolah**: siswa/santri, guru/ustadz, staf, orang tua/wali
- **Anggota eksternal**: masyarakat umum di sekitar sekolah
- **Dual-mode terminologi**: Anggota (koperasi konvensional) vs Nasabah (BMT) — lihat [ADR-009](../core/ADR-009-dual-mode-institution-type.md)

Proses pendaftaran nasabah baru harus melalui formulir aplikasi formal dengan persetujuan, sesuai regulasi koperasi yang mewajibkan dokumentasi keanggotaan.

## Decision

### 1. Registration Flow — Application Form + Approval

Pendaftaran nasabah **wajib** melalui formulir aplikasi yang harus di-approve sebelum nasabah aktif.

```
Applicant mengisi form
        │
        v
┌─────────────────┐
│  Status: DRAFT  │  Applicant bisa edit sebelum submit
└────────┬────────┘
         │ submit
         v
┌─────────────────┐
│ Status: PENDING │  Menunggu approval
└────────┬────────┘
         │
    ┌────┴────┐
    v         v
┌────────┐ ┌──────────┐
│APPROVED│ │ REJECTED │  Reviewer bisa beri catatan penolakan
└───┬────┘ └──────────┘
    │ auto-create
    v
┌─────────────────┐
│ Status: ACTIVE  │  Nasabah aktif, bisa bertransaksi
└─────────────────┘
```

**Aturan:**
- Form yang sudah di-submit tidak bisa di-edit oleh applicant (harus ditolak lalu ajukan ulang)
- Approval menghasilkan pembuatan record nasabah secara otomatis
- Rejection wajib menyertakan alasan (`rejection_reason`)
- Riwayat approval tersimpan di audit log

### 2. Customizable Form Template

Admin tenant dapat meng-customize template formulir aplikasi, terutama bagian **Terms & Conditions**.

```
form_template
├── sections[]
│   ├── personal_info      ← System-defined, tidak bisa dihapus
│   ├── address             ← System-defined, tidak bisa dihapus
│   ├── identity_documents  ← System-defined, tidak bisa dihapus
│   ├── custom_section_*    ← Tenant-defined, bisa tambah/hapus
│   └── terms_conditions    ← System-defined, konten bisa di-edit
└── version                 ← Setiap perubahan template increment version
```

**Aturan template:**
- Section `personal_info`, `address`, `identity_documents`, dan `terms_conditions` adalah **system-defined** — tidak bisa dihapus, tapi konten T&C bisa di-edit
- Tenant bisa menambahkan **custom sections** untuk kebutuhan spesifik
- Setiap perubahan template menghasilkan **version baru**
- Aplikasi yang sudah di-submit menyimpan referensi ke `template_version` yang digunakan saat itu (immutable snapshot)
- Template default disediakan per `coop_type` (general / islamic)

### 3. Authorization — RBAC

Akses dan approval menggunakan **Role-Based Access Control**:

| Permission | Teller | Supervisor | Manager | Admin |
|------------|--------|------------|---------|-------|
| Buat aplikasi (atas nama applicant) | v | v | v | v |
| Review & Approve/Reject | - | v | v | v |
| Edit form template | - | - | v | v |
| Edit T&C content | - | - | v | v |
| Deactivate nasabah | - | - | v | v |
| Reactivate nasabah | - | - | - | v |
| View data nasabah | v | v | v | v |
| Edit data nasabah | v | v | v | v |
| Update foto nasabah | v | v | v | v |
| Update specimen tanda tangan | - | v | v | v |
| Upgrade KYC level | - | v | v | v |
| Transfer nasabah antar cabang | - | - | v | v |
| Merge duplikat nasabah | - | - | - | v |

**Catatan:**
- Teller bisa input aplikasi (walk-in customer) tapi **tidak bisa approve sendiri** — mencegah fraud
- Approval membutuhkan minimal role **Supervisor**
- Edit template dan deactivation membutuhkan minimal role **Manager**
- Merge duplikat hanya **Admin** — operasi berisiko tinggi dengan preview + reversible 30 hari
- Transfer antar cabang: approval Manager di **branch asal** saja, branch tujuan cukup notifikasi

### 4. Member Number — Auto-generate with Configuration

Nomor anggota di-generate otomatis menggunakan format yang bisa dikonfigurasi per tenant:

```
member_number_config:
  mode: "auto" | "manual"          # Tenant config, default: "auto"
  prefix: "KOP"                     # Configurable prefix
  separator: "-"                    # Configurable separator
  sequence_digits: 6                # Jumlah digit sequence
  include_year: true                # Sertakan tahun
  include_branch: true              # Sertakan kode cabang
```

**Contoh hasil:**
```
auto + all options:    KOP-2026-JKT-000001
auto + minimal:        KOP-000001
manual:                Operator input manual (harus unique, validated)
```

**Aturan:**
- Mode `auto` adalah default — `manual` hanya tersedia jika diaktifkan di tenant config
- Nomor anggota **immutable** setelah assigned — tidak bisa diubah
- Uniqueness di-enforce di database level (unique constraint per tenant)
- Sequence auto-increment, tidak ada reuse nomor yang sudah dipakai

### 5. Nasabah Data Model

```
nasabah
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
├── member_number         VARCHAR UNIQUE per tenant
├── registration_date     DATE
│
├── ── Personal Info ──
├── full_name             VARCHAR NOT NULL
├── nickname              VARCHAR
├── birth_place           VARCHAR
├── birth_date            DATE
├── gender                ENUM (male, female)
├── religion              VARCHAR
├── marital_status        ENUM (single, married, divorced, widowed)
├── education_level       VARCHAR
├── occupation            VARCHAR
│
├── ── Identity ──
├── identity_type         ENUM (ktp, sim, passport, kartu_pelajar)
├── identity_number       VARCHAR UNIQUE per tenant
├── identity_expiry       DATE (nullable, KTP seumur hidup = null)
├── identity_photo_url    VARCHAR
│
├── ── Visual Verification ──
├── photo_url             VARCHAR (foto wajah terbaru, bisa di-update)
├── signature_specimen_url VARCHAR (scan tanda tangan untuk verifikasi teller)
│
├── ── Contact ──
├── phone                 VARCHAR
├── email                 VARCHAR (nullable)
├── address               TEXT
├── province              VARCHAR
├── city                  VARCHAR
├── district              VARCHAR
├── village               VARCHAR
├── postal_code           VARCHAR
│
├── ── School Relation ──
├── school_relation_type  ENUM (student, teacher, staff, parent, external)
├── school_entity_id      UUID (nullable, FK → siswa/guru/staf jika internal)
├── sponsor_id            UUID (nullable, FK → nasabah, anggota yang mereferensikan)
│
├── ── KYC ──
├── kyc_level             ENUM (basic, full) DEFAULT 'basic'
├── kyc_verified_at       TIMESTAMPTZ (nullable, kapan di-upgrade ke full)
├── kyc_verified_by       UUID (nullable, FK → user, siapa yang verifikasi)
│
├── ── Membership ──
├── status                ENUM (active, inactive, suspended)
├── join_date             DATE
├── exit_date             DATE (nullable)
├── exit_reason           TEXT (nullable)
│
├── ── Metadata ──
├── application_id        UUID (FK → application, sumber pendaftaran)
├── template_version      INTEGER (versi form saat didaftarkan)
├── custom_fields         JSONB (data dari custom sections)
│
├── ── Vernon Fields ──
├── _rels                 JSONB
├── _data                 JSONB
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 6. Application Data Model

```
nasabah_application
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── branch_id             UUID (FK → branch)
│
├── ── Applicant Data ──
├── form_data             JSONB (snapshot seluruh isian form)
├── sponsor_id            UUID (nullable, FK → nasabah, jika tenant wajibkan sponsor)
├── template_version      INTEGER (versi template yang digunakan)
├── terms_accepted        BOOLEAN NOT NULL
├── terms_accepted_at     TIMESTAMPTZ
│
├── ── Workflow ──
├── status                ENUM (draft, pending, approved, rejected)
├── submitted_at          TIMESTAMPTZ (nullable)
├── reviewed_at           TIMESTAMPTZ (nullable)
├── reviewed_by           UUID (nullable, FK → user)
├── rejection_reason      TEXT (nullable)
│
├── ── Result ──
├── nasabah_id            UUID (nullable, FK → nasabah, filled on approval)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

### 7. Deactivation Rules

Nasabah hanya bisa di-deactivate jika **semua saldo <= 0**:

```
Deactivation Pre-check:
├── Total saldo tabungan     = 0  ✓
├── Total saldo deposito     = 0  ✓
├── Sisa pokok pinjaman      = 0  ✓
├── Tunggakan angsuran       = 0  ✓
├── Denda outstanding        = 0  ✓
└── Simpanan pokok & wajib   = 0  ✓ (harus ditarik/dikembalikan dulu)
    │
    ALL ZERO → Deactivation allowed
    ANY > 0  → Deactivation blocked, return list of blocking items
```

**Aturan:**
- Sistem melakukan **pre-check otomatis** sebelum deactivation — menolak jika ada saldo > 0
- Response menampilkan daftar rekening/pinjaman yang masih memiliki saldo (agar operator tahu apa yang harus diselesaikan)
- Deactivation meng-set `status = inactive`, `exit_date = NOW()`, wajib isi `exit_reason`
- Data nasabah **tidak dihapus** (soft deactivation) — tetap bisa diakses untuk audit dan pelaporan
- Reactivation dimungkinkan oleh Admin — menghasilkan application baru (proses approval ulang)

### 8. Dual-Mode Terminology

| Field/Label | `coop_type = "general"` | `coop_type = "islamic"` |
|-------------|------------------------|------------------------|
| Entity name | Anggota | Nasabah |
| Member number label | No. Anggota | No. Nasabah |
| Registration form title | Formulir Pendaftaran Anggota | Formulir Pendaftaran Nasabah |
| T&C section | Syarat & Ketentuan Keanggotaan | Syarat & Ketentuan Keanggotaan + Akad Wakalah |

BMT mode menambahkan **Akad Wakalah** (perjanjian perwakilan) di T&C sebagai bagian dari compliance syariah.

### 9. Ahli Waris / Beneficiary

Setiap nasabah **wajib** mencantumkan minimal 1 ahli waris saat pendaftaran. Ini merupakan kewajiban regulasi koperasi dan OJK untuk LKM.

```
nasabah_beneficiary
├── id                    UUID v7 (PK)
├── tenant_id             UUID (FK → tenant)
├── nasabah_id            UUID (FK → nasabah) NOT NULL
│
├── ── Data Ahli Waris ──
├── full_name             VARCHAR NOT NULL
├── relationship          ENUM (spouse, child, parent, sibling, other)
├── relationship_detail   VARCHAR (nullable, jika other — jelaskan)
├── identity_number       VARCHAR (KTP ahli waris)
├── phone                 VARCHAR (nullable)
├── address               TEXT (nullable)
│
├── ── Pembagian ──
├── share_percentage      NUMERIC(5,2) NOT NULL (misal: 50.00 = 50%)
├── is_primary            BOOLEAN DEFAULT false (contact person utama saat klaim)
│
├── ── Dokumen Klaim ──
├── claim_document_url    VARCHAR (nullable, diisi saat klaim — surat kematian dll)
├── claim_status          ENUM (none, claimed, settled) DEFAULT 'none'
├── claimed_at            TIMESTAMPTZ (nullable)
│
└── ── Audit ──
    ├── created_at        TIMESTAMPTZ
    ├── created_by        UUID
    ├── updated_at        TIMESTAMPTZ
    └── updated_by        UUID
```

**Aturan:**
- Minimal **1 ahli waris** saat pendaftaran — aplikasi tidak bisa di-submit tanpa ahli waris
- Bisa **multiple** ahli waris per nasabah
- `SUM(share_percentage)` untuk semua ahli waris satu nasabah **harus = 100%** — di-enforce di application layer
- Tepat **1 ahli waris** harus ber-flag `is_primary = true` — sebagai contact person utama saat klaim
- Data ahli waris bisa di-update kapan saja oleh nasabah (melalui teller) — perubahan tercatat di audit log
- Field `claim_document_url` dan `claim_status` diisi saat proses klaim (nasabah meninggal)

### 10. Specimen Tanda Tangan

Tanda tangan nasabah disimpan sebagai scan/foto untuk verifikasi visual oleh teller saat penarikan di atas threshold.

**Aturan:**
- Field `signature_specimen_url` di tabel nasabah — upload saat pendaftaran
- Teller **wajib verifikasi visual** untuk penarikan di atas threshold
- Threshold verifikasi **configurable per tenant** di tenant config (default: Rp 1.000.000)
- Update specimen membutuhkan minimal role **Supervisor** — mencegah penggantian ilegal
- Specimen lama disimpan di audit log sebelum di-replace (riwayat specimen)

```
signature_verification_config:
  enabled: true
  threshold_amount: 1000000       # Configurable per tenant
  require_manager_above: 5000000  # Di atas nominal ini, perlu approval Manager
```

### 11. KYC Level / Limit Tier

Nasabah memiliki level verifikasi identitas yang menentukan batas transaksi bulanan.

| Level | Persyaratan | Limit Transaksi/Bulan | Default |
|-------|-------------|----------------------|---------|
| `basic` | KTP saja | Configurable (default: Rp 10.000.000) | Ya, semua nasabah baru |
| `full` | KTP + verifikasi tambahan (foto, specimen, dokumen pendukung) | Unlimited | Upgrade oleh Supervisor+ |

**Aturan:**
- Semua nasabah baru dimulai dari level **basic**
- Upgrade ke `full` dilakukan oleh **Supervisor+** setelah verifikasi dokumen — tanpa approval flow terpisah, langsung update + audit log
- Limit di-enforce di **transaction layer** (K011), bukan di nasabah layer — nasabah hanya menyimpan `kyc_level`
- Default limit basic **configurable per tenant** — angka 10jt adalah default, bisa dinaikkan/diturunkan
- Downgrade dari `full` ke `basic` dimungkinkan oleh Manager+ jika ada masalah verifikasi

### 12. Transfer Antar Cabang

Nasabah bisa dipindahkan ke branch lain jika pindah domisili atau ada kebutuhan operasional.

```
Transfer Request
        │
        v
┌─────────────────────────┐
│ Manager branch ASAL     │  Approval wajib
│ approve transfer        │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Atomik dalam 1 DB TRX:  │
│ 1. Update nasabah.branch_id             │
│ 2. Update semua rekening.branch_id      │
│ 3. Propagasi Vernon _data & _rels       │
│ 4. Snapshot saldo semua rekening → audit │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ Notifikasi ke branch    │  Tidak perlu approval
│ TUJUAN                  │
└─────────────────────────┘
```

**Aturan:**
- Approval hanya dari **Manager di branch asal** — branch tujuan cukup notifikasi
- Transfer dilakukan dalam **satu database transaction** (atomik) — consistency lebih penting dari performance
- **Semua rekening** ikut pindah — tidak bisa partial transfer
- Snapshot saldo seluruh rekening dicatat di **audit log** saat transfer (untuk reconciliation)
- Riwayat branch sebelumnya tercatat di audit trail
- Nasabah tidak bisa di-transfer jika ada rekening dalam status **FROZEN** — selesaikan dulu

### 13. Merge Duplikat Nasabah

Mekanisme untuk menggabungkan dua record nasabah yang merujuk ke orang yang sama.

```
Admin pilih 2 nasabah
        │
        v
┌─────────────────────────┐
│ STEP 1: Preview Report  │  Tampilkan semua yang akan di-reassign:
│ - Rekening (jumlah, saldo) │
│ - Transaksi (jumlah)    │
│ - Ahli waris            │
│ - Beneficiary           │
└────────┬────────────────┘
         │ Admin confirm
         v
┌─────────────────────────┐
│ STEP 2: Auto-Reconcile  │  Verifikasi:
│ SUM saldo SEBELUM merge │  = SUM saldo SESUDAH merge
└────────┬────────────────┘
         │ Match ✓
         v
┌─────────────────────────┐
│ STEP 3: Execute Merge   │
│ - Pilih "primary" nasabah (data yang dipertahankan)  │
│ - Reassign semua rekening duplicate → primary         │
│ - Reassign semua transaksi duplicate → primary        │
│ - Merge ahli waris (gabungkan, re-normalize share)    │
│ - Soft-delete duplicate (status → merged)             │
│ - Simpan merged_into_id di record duplicate           │
└────────┬────────────────┘
         │
         v
┌─────────────────────────┐
│ STEP 4: Reversible      │  Bisa di-undo selama 30 hari
│ Setelah 30 hari →       │  permanent soft-delete
└─────────────────────────┘
```

**Aturan:**
- Hanya **Admin** yang bisa execute merge
- **Preview report** wajib ditampilkan sebelum eksekusi — Admin harus review
- **Auto-reconciliation**: SUM saldo semua rekening sebelum merge HARUS = SUM sesudah merge — jika tidak match, merge dibatalkan
- Duplicate di-set `status = merged`, `merged_into_id = primary.id` — bukan hard delete
- **Reversible selama 30 hari** — Admin bisa undo, semua reassignment dikembalikan
- Setelah 30 hari, merge menjadi permanent (tapi record duplicate tetap tersimpan sebagai audit)
- Member number duplicate di-retire — tidak di-reuse
- Audit trail lengkap: siapa yang merge, kapan, apa saja yang di-reassign

### 14. Vernon _rels dan _data Structure

Nasabah menggunakan Vernon pattern karena listing nasabah membutuhkan data branch dan sponsor (≥ 3 JOIN untuk listing lengkap).

**_rels:**
```json
{
  "tenant_id":  "018f...",
  "branch_id":  "018f...",
  "sponsor_id": "018f..."
}
```

**_data:**
```json
{
  "branch": {
    "id":   "018f...",
    "name": "Cabang Jakarta Pusat",
    "code": "JKT"
  },
  "sponsor": {
    "id":            "018f...",
    "full_name":     "Haji Ahmad",
    "member_number": "KOP-2026-JKT-000001"
  }
}
```

**Catatan keamanan:**
- **Tidak ada data sensitif** di `_data` — no `identity_number`, no saldo, no data finansial
- `_data` hanya berisi informasi yang dibutuhkan untuk listing/display
- Data sensitif harus diambil dari kolom asli nasabah via query detail (bukan listing)

**SyncEngine triggers:**
- `BranchUpdatedEvent` → update `_data.branch` di semua nasabah cabang tersebut
- `NasabahUpdatedEvent` (sponsor) → update `_data.sponsor` di semua nasabah yang punya sponsor tersebut

### 15. Referral / Sponsor

Beberapa koperasi mensyaratkan anggota baru memiliki sponsor (anggota existing yang mereferensikan).

**Aturan:**
- Field `sponsor_id` (nullable, FK → nasabah) di application form dan nasabah
- **Configurable per tenant** — 3 mode:
  - `disabled`: field sponsor tidak ditampilkan (default)
  - `optional`: field sponsor ditampilkan tapi tidak wajib
  - `required`: field sponsor wajib diisi, aplikasi tidak bisa di-submit tanpa sponsor
- Sponsor harus nasabah **ACTIVE** di tenant yang sama
- Sponsor tercatat permanen di nasabah — berguna untuk:
  - Tracking growth/referral network
  - Bonus SHU berdasarkan kontribusi referral (jika koperasi menerapkan)
- Jika sponsor di-deactivate, link tetap tersimpan (historical reference)

### 16. Future: Family Linking (Phase 2)

**Ditunda ke fase 2** — tidak masuk MVP. Didokumentasikan di sini agar tim future tahu arahnya.

**Rencana:**
- Tabel `nasabah_family_link` (nasabah_id_1, nasabah_id_2, relation_type)
- Relation types: spouse, parent_child, sibling
- Use cases: laporan SHU keluarga, pinjaman joint, analisis keanggotaan keluarga
- Tidak ada perubahan schema sekarang — data model nasabah saat ini tidak menutup pintu untuk fitur ini

### 17. Dormant di Level Nasabah

**Tidak ada flag dormant di level nasabah** — dormant hanya berlaku di level rekening (lihat [ADR-K002](./ADR-K002-rekening.md) bagian Dormant Account Handling).

**Alasan:**
- Dormant adalah kondisi rekening, bukan kondisi orang
- Satu nasabah bisa punya rekening dormant dan rekening aktif sekaligus
- Nasabah tetap ACTIVE selama statusnya active

**Biaya keanggotaan tahunan:**
- Sebagai config **opsional** per tenant (`annual_membership_fee`)
- Jika diaktifkan, nominal dan mekanisme collection diputuskan di [ADR-K011](./ADR-K011-transaksi.md)
- Tidak mempengaruhi status nasabah — jika tidak bayar, penanganan diserahkan ke kebijakan koperasi (bukan otomatis deactivate)

## Consequences

### Positif

- **Auditable** — setiap nasabah punya jejak dari aplikasi hingga approval
- **Compliant** — formulir formal, ahli waris wajib, dan KYC level memenuhi regulasi koperasi/OJK
- **Flexible** — template customizable, sponsor configurable, threshold configurable per tenant
- **Safe deactivation** — pre-check mencegah nasabah keluar dengan saldo menggantung
- **Fraud prevention** — teller tidak bisa approve sendiri, specimen verifikasi, separation of duties
- **Data integrity** — merge duplikat dengan preview, auto-reconciliation, dan reversibility 30 hari
- **Multi-branch ready** — transfer antar cabang atomik dengan audit trail lengkap
- **Beneficiary ready** — ahli waris tercatat sejak awal, mempercepat proses klaim

### Negatif

- **Onboarding lebih lambat** — butuh approval, ahli waris wajib, dan upload dokumen
- **Template versioning complexity** — harus maintain snapshot per aplikasi
- **RBAC overhead** — lebih banyak permission yang harus di-setup
- **Merge complexity** — reassignment rekening dan transaksi lintas tabel berisiko jika ada bug
- **Transfer atomik** — operasi berat yang bisa lock banyak row jika nasabah punya banyak rekening

### Mitigasi

- Supervisor/Manager bisa approve secara batch untuk efisiensi
- Template versioning otomatis (increment on save), tidak perlu manage manual
- Default role set disediakan saat tenant onboarding — termasuk semua permission baru
- Merge dilindungi oleh preview report + auto-reconciliation + reversibility 30 hari
- Transfer menggunakan row-level lock (bukan table lock) — impact terbatas pada nasabah yang di-transfer

## Alternatives Considered

### A. Direct Registration (tanpa approval)

Nasabah langsung aktif setelah mengisi form, tanpa proses approval.

**Ditolak** karena: tidak memenuhi regulasi koperasi yang mewajibkan persetujuan pengurus. Juga membuka risiko data nasabah tidak valid karena tidak ada tahap verifikasi.

### B. ACL (Access Control List) instead of RBAC

Per-user permission tanpa grouping ke role.

**Ditolak** karena: koperasi memiliki struktur organisasi yang jelas (Teller → Supervisor → Manager → Admin). RBAC lebih natural untuk konteks ini dan lebih mudah di-maintain. ACL terlalu granular untuk kebutuhan saat ini — bisa dipertimbangkan jika ada kebutuhan permission per-entity di masa depan.

### C. Hard Delete on Deactivation

Menghapus data nasabah saat deactivation.

**Ditolak** karena: data transaksi historis dan laporan regulasi membutuhkan referensi ke nasabah. Hard delete akan memutus referential integrity dan melanggar ketentuan retensi data koperasi.

### D. Deactivation dengan saldo > 0 (force close)

Mengizinkan deactivation meskipun masih ada saldo, dengan proses settlement otomatis.

**Ditolak** karena: settlement otomatis (pengembalian simpanan, write-off pinjaman) membutuhkan keputusan manusia dan otorisasi khusus. Risiko finansial terlalu tinggi untuk diotomasi.

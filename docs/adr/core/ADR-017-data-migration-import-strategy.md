# ADR-017: Strategi Migrasi Data & Import

**Status**: Proposed
**Date**: 2026-04-15
**Deciders**: Erick Mo, COO Review Panel

## Context

Sekolah yang mengadopsi SekolahPro sudah memiliki data existing di berbagai sumber:

| Sumber Data | Format | Contoh |
|-------------|--------|--------|
| **Data siswa** (biodata, NIS, NISN) | Excel, export Dapodik | `.xlsx`, `.csv` |
| **Data guru/staff** | Excel | Spreadsheet kepegawaian |
| **Data keuangan** (SPP, tagihan) | Excel, software akuntansi | Rekap bendahara |
| **Data kelas & tahun ajaran** | Excel, sistem lama | Jibas, AdminSekolah |
| **Data rapor/nilai** | Excel, arsip kertas | Rekap wali kelas |

**Masalah utama**: Tanpa tools import, onboarding 1 sekolah bisa memakan **2-4 minggu** input manual. Ini menjadi bottleneck terbesar untuk customer acquisition. Sekolah dengan 500+ siswa tidak akan mau migrasi jika prosesnya manual.

Sistem lama yang umum dipakai sekolah di Indonesia:
- **Dapodik** (data pokok pendidikan, export CSV/Excel)
- **Jibas** (open-source SIM sekolah)
- **AdminSekolah** (SaaS)
- **Spreadsheet custom** (mayoritas sekolah)

## Decision

### 1. Excel/CSV Import sebagai Jalur Utama

Admin sekolah mengupload file Excel/CSV menggunakan template standar yang disediakan SekolahPro. Ini dipilih karena **semua sekolah bisa menggunakan Excel** — tidak perlu integrasi API ke sistem lama.

### 2. Template-Based Import

SekolahPro menyediakan template Excel yang bisa didownload per entitas:

| Entitas | Template | Kolom Utama | Validasi |
|---------|----------|-------------|----------|
| **Siswa** | `template_siswa.xlsx` | NIS, NISN, nama, kelas, jenis_kelamin, tanggal_lahir, alamat | NIS unique per company, NISN 10 digit, kelas harus ada |
| **Guru/Staff** | `template_guru.xlsx` | NIP/NIK, nama, jabatan, mata_pelajaran, email, no_telp | NIP unique per company, email format valid |
| **Tipe Biaya** | `template_tipe_biaya.xlsx` | kode, nama, nominal, frekuensi (bulanan/tahunan/sekali) | Kode unique, nominal > 0 |
| **Tagihan SPP** | `template_tagihan.xlsx` | NIS_siswa, tipe_biaya, bulan, tahun, nominal, status | NIS harus ada di sistem, bulan 1-12 |
| **Kelas** | `template_kelas.xlsx` | kode_kelas, nama, tingkat, wali_kelas_NIP, kapasitas | Kode unique per tahun ajaran |
| **Tahun Ajaran** | `template_tahun_ajaran.xlsx` | kode, nama, tanggal_mulai, tanggal_selesai, is_active | Tanggal valid, tidak overlap |

### 3. Alur Import

```
Download template → Isi data → Upload file → Dry-run validasi →
Review error report → Perbaiki data → Re-upload → Konfirmasi import →
Batch insert → Vernon sync → Laporan verifikasi
```

### 4. Dry-Run Validation

Setiap upload **wajib** melewati dry-run terlebih dahulu:

- Parsing seluruh baris file
- Validasi format, tipe data, constraint (unique, foreign key)
- Deteksi duplikat terhadap data existing di database
- Return **error report** tanpa menyimpan apapun ke database
- Admin melihat: total baris, valid, error, warning, duplikat
- Admin baru bisa klik "Konfirmasi Import" setelah review

### 5. Batch Processing (Async)

Import dengan **1000+ baris** diproses secara asynchronous:

```sql
CREATE TABLE import_batches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    tenant_id       UUID NOT NULL,
    company_id      UUID NOT NULL,

    -- Metadata
    entity_type     VARCHAR(30) NOT NULL,  -- 'students', 'teachers', etc.
    file_name       VARCHAR(255) NOT NULL,
    uploaded_by     UUID NOT NULL,

    -- Progress
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_rows      INTEGER NOT NULL DEFAULT 0,
    processed_rows  INTEGER NOT NULL DEFAULT 0,
    success_rows    INTEGER NOT NULL DEFAULT 0,
    error_rows      INTEGER NOT NULL DEFAULT 0,

    -- Result
    error_report    JSONB,
    batch_record_ids JSONB,  -- array of inserted record UUIDs

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,

    CONSTRAINT chk_status CHECK (status IN (
        'pending', 'validating', 'validated', 'importing',
        'completed', 'failed', 'rolled_back'
    )),
    CONSTRAINT chk_entity CHECK (entity_type IN (
        'students', 'teachers', 'fee_types',
        'invoices', 'class_rooms', 'academic_years'
    ))
);

CREATE INDEX idx_import_batch_company ON import_batches (tenant_id, company_id);
CREATE INDEX idx_import_batch_status ON import_batches (status) WHERE status NOT IN ('completed', 'rolled_back');
```

Progress tracking via polling atau WebSocket:
```
GET /api/v1/imports/{batch_id}/progress
→ { "status": "importing", "processed": 450, "total": 1200, "percent": 37 }
```

### 6. Conflict Resolution (Duplikat)

Deteksi duplikat berdasarkan key unik per entitas:

| Entitas | Duplicate Key | Opsi |
|---------|--------------|------|
| Siswa | NIS atau NISN | skip, overwrite, merge |
| Guru | NIP atau email | skip, overwrite, merge |
| Kelas | kode_kelas + tahun_ajaran | skip, overwrite |
| Tagihan | NIS + tipe_biaya + bulan + tahun | skip, overwrite |

- **Skip**: Baris duplikat dilewati, dicatat di report
- **Overwrite**: Data existing di-update dengan data baru
- **Merge**: Field kosong di existing diisi dari data baru, field yang sudah ada tidak ditimpa

Admin memilih strategi conflict resolution **sebelum** konfirmasi import.

### 7. Vernon Sync Post-Import

Setelah batch insert selesai, trigger Vernon `_data` sync untuk semua record yang di-import:

1. Batch insert ke tabel utama (e.g., `students`)
2. Emit event `StudentImported` per record (batched, max 100 per event)
3. Vernon sync handler membangun `_data` JSONB untuk setiap record
4. Sync dijalankan async — tidak blocking import completion

### 8. Rollback

Setiap import batch bisa di-rollback oleh admin:

- `batch_record_ids` menyimpan semua UUID record yang dibuat
- Rollback = soft-delete semua record dalam batch
- Hanya bisa rollback batch dengan status `completed`
- Rollback juga trigger Vernon `_data` cleanup
- Batas waktu rollback: **72 jam** setelah import

### 9. API untuk Migrasi Sistem

REST API untuk migrasi otomatis dari sistem lain:

```
# Import
POST   /api/v1/imports/upload          — Upload file + dry-run
POST   /api/v1/imports/confirm         — Konfirmasi import setelah review
GET    /api/v1/imports/{id}/progress   — Cek progress
POST   /api/v1/imports/{id}/rollback   — Rollback batch
GET    /api/v1/imports                 — List import history

# Templates
GET    /api/v1/imports/templates/{entity}  — Download template Excel

# API Migration (untuk integrasi sistem)
POST   /api/v1/migrations/students     — Bulk create via JSON API
POST   /api/v1/migrations/teachers     — Bulk create via JSON API
POST   /api/v1/migrations/invoices     — Bulk create via JSON API
```

API migration endpoint menerima JSON array dan mengikuti flow yang sama (validasi, batch, Vernon sync).

## Consequences

### Positive

- **Onboarding cepat**: Sekolah bisa migrasi data dalam hitungan jam, bukan minggu.
- **Self-service**: Admin sekolah bisa import sendiri tanpa bantuan tim teknis.
- **Aman**: Dry-run + rollback mencegah data corrupt masuk ke sistem.
- **Vernon consistent**: Post-import sync menjamin `_data` selalu up-to-date.
- **Scalable**: Async processing mendukung import ribuan baris tanpa timeout.
- **Audit trail**: Setiap batch tercatat — siapa upload, kapan, berapa record.

### Negative / Trade-offs

- **Template maintenance**: Setiap perubahan schema domain memerlukan update template.
- **Mapping complexity**: Data dari Dapodik/Jibas formatnya berbeda — user perlu map manual ke template.
- **Vernon sync load**: Import 5000 siswa sekaligus = 5000 Vernon sync — perlu rate limiting.
- **Storage**: File upload memerlukan object storage (S3/MinIO).
- **Error UX**: Menampilkan error report ratusan baris dengan UX yang baik butuh effort di frontend.

## Alternatives Considered

### 1. Manual Input Only (Tanpa Import)
- Ditolak: Onboarding 1 sekolah bisa 2-4 minggu. Tidak scalable untuk target 100+ sekolah.

### 2. Automated Scraping dari Sistem Lama
- Ditolak: Setiap sistem punya format berbeda, banyak yang tidak punya API. Maintenance cost terlalu tinggi untuk ROI yang didapat.

### 3. Professional Services Migration (Tim Teknis Datang ke Sekolah)
- Deferred: Bisa ditawarkan sebagai layanan premium untuk sekolah besar. Tapi jalur utama harus self-service agar scalable. Akan dipertimbangkan setelah MVP.

### 4. Direct Database Migration (ETL Pipeline)
- Ditolak: Memerlukan akses ke database sistem lama. Mayoritas sekolah tidak punya akses DB atau bahkan tidak pakai database (masih Excel).
